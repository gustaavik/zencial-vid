package season

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// DeleteSeason removes a season. Fails if it has episodes.
func (s *Service) DeleteSeason(ctx context.Context, seasonID, uploaderID uuid.UUID) *apperror.AppError {
	season, err := s.seasonRepo.GetByID(ctx, seasonID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch season", err)
	}
	if season == nil {
		return apperror.NotFound(apperror.CodeSeasonNotFound, "season not found", nil)
	}

	series, err := s.seriesRepo.GetByID(ctx, season.SeriesID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch series", err)
	}
	if series == nil || series.UploadedBy != uploaderID {
		return apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", nil)
	}

	if err := s.seasonRepo.Delete(ctx, seasonID); err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to delete season", err)
	}
	return nil
}
