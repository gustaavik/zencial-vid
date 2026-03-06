package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminChangePlanInput holds the data required for an admin plan change.
type AdminChangePlanInput struct {
	SubscriptionID uuid.UUID
	PlanID         uuid.UUID
}

// AdminChangePlan changes the plan of a subscription (admin operation).
func (s *Service) AdminChangePlan(ctx context.Context, input AdminChangePlanInput) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetByID(ctx, input.SubscriptionID)
	if err != nil {
		s.log.Error("getting subscription for admin plan change", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "subscription not found", domain.ErrNoActiveSubscription)
	}

	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.PlanID)
	if err != nil || plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	sub.PlanID = plan.ID
	sub.Plan = plan
	sub.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("admin changing plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to change plan", err)
	}
	return sub, nil
}
