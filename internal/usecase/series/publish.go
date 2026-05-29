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

// Publish transitions a series to published status.
func (s *Service) Publish(ctx context.Context, seriesID uuid.UUID) (*entity.Series, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		s.log.Error("getting series for publish", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	if series.Status == entity.SeriesStatusPublished {
		return series, nil
	}

	series.Publish()
	if err := s.seriesRepo.Update(ctx, series); err != nil {
		s.log.Error("updating series status", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to publish series", err)
	}

	if err := s.dispatcher.Dispatch(event.SeriesPublished{
		SeriesID:  series.ID,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series published event", "error", err)
	}

	return series, nil
}
