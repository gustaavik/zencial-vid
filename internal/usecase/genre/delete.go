package genre

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete removes a genre by its ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting genre for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get genre", err)
	}
	if genre == nil {
		return apperror.NotFound(apperror.CodeGenreNotFound, "genre not found", domain.ErrGenreNotFound)
	}

	if err := s.genreRepo.Delete(ctx, id); err != nil {
		s.log.Error("deleting genre", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete genre", err)
	}

	if err := s.dispatcher.Dispatch(event.GenreDeleted{
		GenreID:   id,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching genre deleted event", "error", err)
	}

	return nil
}
