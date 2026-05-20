package series

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateWatchProgress records the last-watched episode for a user in a series.
func (s *Service) UpdateWatchProgress(ctx context.Context, userID, seriesID, episodeID uuid.UUID) *apperror.AppError {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for watch progress update", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	video, err := s.videoRepo.GetByID(ctx, episodeID)
	if err != nil {
		s.log.Error("getting video for watch progress update", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return apperror.NotFound(apperror.CodeVideoNotFound, "episode not found", domain.ErrVideoNotFound)
	}
	if video.SeriesID == nil || *video.SeriesID != seriesID {
		return apperror.BadRequest(apperror.CodeBadRequest, "video does not belong to this series", nil)
	}

	if err := s.seriesWpRepo.Upsert(ctx, userID, seriesID, episodeID); err != nil {
		s.log.Error("upserting series watch progress", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to update watch progress", err)
	}

	return nil
}

// GetWatchProgress returns the last-watched episode progress for a user in a series.
func (s *Service) GetWatchProgress(ctx context.Context, userID, seriesID uuid.UUID) (*entity.SeriesWatchProgress, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for watch progress", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	progress, err := s.seriesWpRepo.Get(ctx, userID, seriesID)
	if err != nil {
		s.log.Error("getting series watch progress", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get watch progress", err)
	}
	if progress == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesWatchProgressNotFound, "no watch progress for this series", domain.ErrSeriesWatchProgressNotFound)
	}

	return progress, nil
}
