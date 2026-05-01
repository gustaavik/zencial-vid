package watchprogress

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Service handles watch progress (resume / continue watching) use cases.
type Service struct {
	progressRepo repository.WatchProgressRepository
	videoRepo    repository.VideoRepository
	log          *slog.Logger
}

// NewService creates a new watch progress Service.
func NewService(progressRepo repository.WatchProgressRepository, videoRepo repository.VideoRepository, log *slog.Logger) *Service {
	return &Service{
		progressRepo: progressRepo,
		videoRepo:    videoRepo,
		log:          log,
	}
}

// Upsert records the user's current playback position for a video.
//
// Negative positions are rejected as BadRequest. Positions exceeding the video's
// duration are clamped to the duration so end-of-stream callbacks don't error out.
// Returns NotFound when the video does not exist.
func (s *Service) Upsert(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int64) *apperror.AppError {
	if positionSeconds < 0 {
		return apperror.BadRequest(apperror.CodeBadRequest, "position_seconds must be >= 0", nil)
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("watch_progress: getting video", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	// Clamp to the video's duration so an over-shoot from the player (rounding,
	// final frame, etc.) doesn't cause storage drift.
	if video.Duration.Seconds > 0 && positionSeconds > video.Duration.Seconds {
		positionSeconds = video.Duration.Seconds
	}

	if err := s.progressRepo.Upsert(ctx, userID, videoID, positionSeconds); err != nil {
		s.log.Error("watch_progress: upserting", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to save watch progress", err)
	}
	return nil
}

// Get returns the user's saved progress for a video, also returning the video's
// duration so callers can compute completion percentages without a second lookup.
//
// Returns NotFound (CodeWatchProgressNotFound) when no progress is recorded.
func (s *Service) Get(ctx context.Context, userID, videoID uuid.UUID) (*entity.WatchProgress, int64, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("watch_progress: getting video", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, 0, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	progress, err := s.progressRepo.Get(ctx, userID, videoID)
	if err != nil {
		s.log.Error("watch_progress: getting progress", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to get watch progress", err)
	}
	if progress == nil {
		return nil, 0, apperror.NotFound(apperror.CodeWatchProgressNotFound, "watch progress not found", domain.ErrWatchProgressNotFound)
	}

	return progress, video.Duration.Seconds, nil
}

// Delete removes the user's saved progress for a video.
//
// Returns NotFound when no progress is recorded.
func (s *Service) Delete(ctx context.Context, userID, videoID uuid.UUID) *apperror.AppError {
	if err := s.progressRepo.Delete(ctx, userID, videoID); err != nil {
		if errors.Is(err, domain.ErrWatchProgressNotFound) {
			return apperror.NotFound(apperror.CodeWatchProgressNotFound, "watch progress not found", err)
		}
		s.log.Error("watch_progress: deleting", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete watch progress", err)
	}
	return nil
}

// ListInProgress returns the user's "continue watching" feed.
func (s *Service) ListInProgress(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.VideoWithProgress, int64, *apperror.AppError) {
	items, total, err := s.progressRepo.ListInProgress(ctx, userID, page)
	if err != nil {
		s.log.Error("watch_progress: listing", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list watch progress", err)
	}
	return items, total, nil
}
