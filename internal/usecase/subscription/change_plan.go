package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ChangePlanInput holds the data required to change a subscription's plan.
type ChangePlanInput struct {
	UserID uuid.UUID
	PlanID uuid.UUID
}

// ChangePlan changes the plan of the user's active subscription.
func (s *Service) ChangePlan(ctx context.Context, input ChangePlanInput) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil || sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "no active subscription", domain.ErrNoActiveSubscription)
	}

	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.PlanID)
	if err != nil || plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	sub.PlanID = plan.ID
	sub.Plan = plan
	sub.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("changing plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to change plan", err)
	}
	return sub, nil
}
