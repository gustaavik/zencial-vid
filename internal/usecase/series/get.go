package series

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID returns a series by its UUID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Series, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting series by ID", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	genreIDs, err := s.seriesRepo.GetGenreIDs(ctx, series.ID)
	if err != nil {
		s.log.Error("getting series genre IDs", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series genres", err)
	}
	series.GenreIDs = genreIDs

	return series, nil
}

// GetBySlug returns a series by its slug.
func (s *Service) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Series, *apperror.AppError) {
	series, err := s.seriesRepo.GetBySlug(ctx, slug)
	if err != nil {
		s.log.Error("getting series by slug", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	genreIDs, err := s.seriesRepo.GetGenreIDs(ctx, series.ID)
	if err != nil {
		s.log.Error("getting series genre IDs", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series genres", err)
	}
	series.GenreIDs = genreIDs

	return series, nil
}
