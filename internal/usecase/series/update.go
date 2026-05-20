package series

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateInput holds the data needed to update a series.
type UpdateInput struct {
	ID               uuid.UUID
	Title            *string
	Description      *string
	Creator          *string
	CoverImageKey    *string
	GenreIDs         []uuid.UUID
	MinimumPlanLevel *int
	CallerID         uuid.UUID
	CallerRoles      []entity.UserRole
}

// Update updates a series' metadata.
func (s *Service) Update(ctx context.Context, input *UpdateInput) (*entity.Series, *apperror.AppError) {
	series, err := s.seriesRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting series for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get series", err)
	}
	if series == nil {
		return nil, apperror.NotFound(apperror.CodeSeriesNotFound, "series not found", domain.ErrSeriesNotFound)
	}

	if !entity.HasRole(input.CallerRoles, entity.RoleAdmin) && series.UploadedBy != input.CallerID {
		return nil, apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", domain.ErrSeriesOwnershipRequired)
	}

	if input.Title != nil {
		series.Title = *input.Title
		slug, err := valueobject.NewSlug(*input.Title)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug", err)
		}
		series.Slug = slug.WithRandomID()
	}
	if input.Description != nil {
		series.Description = *input.Description
	}
	if input.Creator != nil {
		series.Creator = *input.Creator
	}
	if input.CoverImageKey != nil {
		series.CoverImageKey = *input.CoverImageKey
	}
	if input.MinimumPlanLevel != nil {
		series.MinimumPlanLevel = input.MinimumPlanLevel
	}

	series.UpdatedAt = time.Now().UTC()
	if err := s.seriesRepo.Update(ctx, series); err != nil {
		s.log.Error("updating series", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update series", err)
	}

	if input.GenreIDs != nil {
		if err := s.seriesRepo.SetGenres(ctx, series.ID, input.GenreIDs); err != nil {
			s.log.Error("setting series genres", "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to update series genres", err)
		}
		series.GenreIDs = input.GenreIDs
	}

	if err := s.dispatcher.Dispatch(event.SeriesUpdated{
		SeriesID:  series.ID,
		Field:     "metadata",
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching series updated event", "error", err)
	}

	return series, nil
}
