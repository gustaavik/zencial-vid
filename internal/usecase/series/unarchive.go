package series

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Unarchive restores an archived series to draft status.
func (s *Service) Unarchive(ctx context.Context, seriesID uuid.UUID) (*entity.Series, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for unarchive", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	series.Unarchive()
	if err := s.seriesRepo.Update(ctx, series); err != nil {
		s.log.Error("updating series status", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to unarchive series", err)
	}

	if err := s.dispatcher.Dispatch(event.SeriesRestored{
		SeriesID:  series.ID,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series restored event", "error", err)
	}

	return series, nil
}
