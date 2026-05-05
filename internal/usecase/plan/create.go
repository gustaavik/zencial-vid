package plan

import (
	"context"
	"time"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreateInput holds the data needed to create a plan.
type CreateInput struct {
	Name          string
	Description   string
	Price         float64
	Level         int
	StripePriceID string
}

// Create creates a new plan.
func (s *Service) Create(ctx context.Context, input CreateInput) (*entity.Plan, *apperror.AppError) {
	slug, err := valueobject.NewSlug(input.Name)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid name for slug generation", err)
	}

	exists, err := s.planRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		s.log.Error("checking plan slug existence", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check slug", err)
	}
	if exists {
		return nil, apperror.Conflict(apperror.CodePlanSlugConflict, "plan slug already exists", domain.ErrPlanSlugExists)
	}

	plan := entity.NewPlan(input.Name, slug, input.Description, input.Price, input.Level)
	plan.StripePriceID = input.StripePriceID

	if err := s.planRepo.Create(ctx, plan); err != nil {
		s.log.Error("creating plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create plan", err)
	}

	if err := s.dispatcher.Dispatch(event.PlanCreated{
		PlanID:    plan.ID,
		ActorID:   actor.FromContext(ctx),
		Name:      plan.Name,
		Slug:      plan.Slug.String(),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching plan created event", "error", err)
	}

	return plan, nil
}
