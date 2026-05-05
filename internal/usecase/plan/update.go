package plan

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

// UpdateInput holds the data needed to update a plan.
type UpdateInput struct {
	ID            uuid.UUID
	Name          *string
	Description   *string
	Price         *float64
	Level         *int
	StripePriceID *string
	IsActive      *bool
}

// Update updates an existing plan.
func (s *Service) Update(ctx context.Context, input UpdateInput) (*entity.Plan, *apperror.AppError) {
	plan, err := s.planRepo.GetByID(ctx, input.ID)
	if err != nil {
		s.log.Error("getting plan for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}
	if plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	if input.Name != nil {
		newSlug, err := valueobject.NewSlug(*input.Name)
		if err != nil {
			return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid name for slug generation", err)
		}
		if newSlug.String() != plan.Slug.String() {
			exists, err := s.planRepo.ExistsBySlug(ctx, newSlug)
			if err != nil {
				s.log.Error("checking plan slug existence", "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
			}
			if exists {
				return nil, apperror.Conflict(apperror.CodePlanSlugConflict, "plan slug already exists", domain.ErrPlanSlugExists)
			}
			plan.Slug = newSlug
		}
		plan.Name = *input.Name
	}

	if input.Description != nil {
		plan.Description = *input.Description
	}
	if input.Price != nil {
		plan.Price = *input.Price
	}
	if input.Level != nil {
		plan.Level = *input.Level
	}
	if input.StripePriceID != nil {
		plan.StripePriceID = *input.StripePriceID
	}
	if input.IsActive != nil {
		plan.IsActive = *input.IsActive
	}

	plan.UpdatedAt = time.Now().UTC()

	if err := s.planRepo.Update(ctx, plan); err != nil {
		s.log.Error("updating plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update plan", err)
	}

	if err := s.dispatcher.Dispatch(event.PlanUpdated{
		PlanID:    plan.ID,
		ActorID:   actor.FromContext(ctx),
		Name:      plan.Name,
		Slug:      plan.Slug.String(),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching plan updated event", "error", err)
	}

	return plan, nil
}
