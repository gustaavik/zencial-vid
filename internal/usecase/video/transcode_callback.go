package video

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// MarkTranscodeComplete transitions a processing video to published.
// Idempotent: if the video is already published, it returns the current state without error.
// Errors if the video is in any other non-processing state (draft, archived, failed) — the
// caller is expected to retry or surface the issue.
func (s *Service) MarkTranscodeComplete(ctx context.Context, id uuid.UUID) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for transcode complete", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	switch video.Status {
	case entity.VideoStatusPublished:
		return video, nil
	case entity.VideoStatusProcessing:
		// proceed
	default:
		s.log.Warn("transcode-complete callback for video not in processing state",
			"video_id", id, "status", video.Status)
		return nil, apperror.BadRequest(apperror.CodeVideoNotTranscoding,
			"video is not in processing state", nil)
	}

	video.MarkTranscoded()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("marking video transcoded", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to mark video transcoded", err)
	}

	if err := s.dispatcher.Dispatch(event.VideoPublished{
		VideoID:   video.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching video published event", "error", err)
	}

	return video, nil
}

// MarkTranscodeFailed records a transcoding failure for a video.
// Idempotent: if the video is already in failed state, the reason is updated but no event re-fires.
// Errors if the video is not in processing or failed state.
func (s *Service) MarkTranscodeFailed(ctx context.Context, id uuid.UUID, reason string) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting video for transcode failure", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if video.Status != entity.VideoStatusProcessing && video.Status != entity.VideoStatusFailed {
		s.log.Warn("transcode-failed callback for video not in processing/failed state",
			"video_id", id, "status", video.Status)
		return nil, apperror.BadRequest(apperror.CodeVideoNotTranscoding,
			"video is not in processing state", nil)
	}

	wasProcessing := video.Status == entity.VideoStatusProcessing
	video.MarkTranscodeFailed(reason)

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("marking video transcode failed", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to record transcode failure", err)
	}

	if wasProcessing {
		if err := s.dispatcher.Dispatch(event.VideoTranscodeFailed{
			VideoID:   video.ID,
			Reason:    reason,
			Timestamp: time.Now().UTC(),
		}); err != nil {
			s.log.Error("dispatching video transcode failed event", "error", err)
		}
	}

	return video, nil
}
