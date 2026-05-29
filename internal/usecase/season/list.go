package season

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListSeasons returns all seasons for a series, ordered by season_number.
func (s *Service) ListSeasons(ctx context.Context, seriesID uuid.UUID) ([]entity.Season, *apperror.AppError) {
	seasons, err := s.seasonRepo.ListBySeries(ctx, seriesID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list seasons", err)
	}
	return seasons, nil
}
