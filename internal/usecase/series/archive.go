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

// ArchiveInput holds the data needed to archive a series.
type ArchiveInput struct {
	SeriesID    uuid.UUID
	CallerID    uuid.UUID
	CallerRoles []entity.UserRole
}

// Archive soft-deletes a series by setting it to archived status.
func (s *Service) Archive(ctx context.Context, input *ArchiveInput) *apperror.AppError {
	series, err := s.seriesRepo.GetByID(ctx, input.SeriesID)
	if err != nil {
		s.log.Error("getting series for archive", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) && series.UploadedBy != input.CallerID {
		return apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", domain.ErrSeriesOwnershipRequired)
	}

	if series.Status == entity.SeriesStatusArchived {
		return nil
	}

	series.Archive()
	if err := s.seriesRepo.Update(ctx, series); err != nil {
		s.log.Error("updating series status", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to archive series", err)
	}

	if err := s.dispatcher.Dispatch(event.SeriesArchived{
		SeriesID:  series.ID,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series archived event", "error", err)
	}

	return nil
}
