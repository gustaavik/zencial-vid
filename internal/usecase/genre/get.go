package genre

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID returns a genre by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, *apperror.AppError) {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting genre", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get genre", err)
	}
	if genre == nil {
		return nil, apperror.NotFound(apperror.CodeGenreNotFound, "genre not found", domain.ErrGenreNotFound)
	}
	return genre, nil
}
