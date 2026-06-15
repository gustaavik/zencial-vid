package analytics

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// RecordPlaybackInput is one cumulative playback heartbeat from a client.
type RecordPlaybackInput struct {
	SessionID       uuid.UUID
	VideoID         uuid.UUID
	UserID          uuid.UUID
	Source          string
	Platform        string
	CountryCode     string
	PositionSeconds int64
	WatchedSeconds  int64
	WatchedBuckets  []int
	Completed       bool
}

// RecordPlayback persists a playback heartbeat. Ingestion is best-effort:
// session-id conflicts (a session ID reused for another video or user) are
// logged and swallowed rather than surfaced to the player.
func (s *Service) RecordPlayback(ctx context.Context, in *RecordPlaybackInput) *apperror.AppError {
	if in.PositionSeconds < 0 || in.WatchedSeconds < 0 {
		return apperror.BadRequest(apperror.CodeBadRequest, "position_seconds and watched_seconds must be >= 0", nil)
	}

	video, err := s.videoRepo.GetByID(ctx, in.VideoID)
	if err != nil {
		s.log.Error("analytics: getting video for playback event", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	position := in.PositionSeconds
	watched := in.WatchedSeconds
	if video.Duration.Seconds > 0 {
		// Clamp to the video's duration so player over-shoot doesn't skew stats;
		// watched time is allowed up to 3x duration to accommodate rewatching
		// within one session.
		if position > video.Duration.Seconds {
			position = video.Duration.Seconds
		}
		if maxWatched := video.Duration.Seconds * 3; watched > maxWatched {
			watched = maxWatched
		}
	}

	userID := in.UserID
	hb := &repository.PlaybackHeartbeat{
		SessionID:            in.SessionID,
		VideoID:              in.VideoID,
		UserID:               &userID,
		Source:               entity.NormalizePlaybackSource(in.Source),
		Platform:             entity.NormalizePlaybackPlatform(in.Platform),
		CountryCode:          in.CountryCode,
		PositionSeconds:      position,
		WatchedSeconds:       watched,
		ViewThresholdSeconds: entity.ViewThresholdSeconds(video.Duration.Seconds),
		WatchedBuckets:       in.WatchedBuckets,
		Completed:            in.Completed,
	}

	if err := s.playbackRepo.UpsertHeartbeat(ctx, hb); err != nil {
		if errors.Is(err, domain.ErrPlaybackSessionConflict) {
			s.log.Warn("analytics: playback session conflict",
				"session_id", in.SessionID, "video_id", in.VideoID, "user_id", in.UserID)
			return nil
		}
		s.log.Error("analytics: upserting playback heartbeat", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to record playback event", err)
	}
	return nil
}
