package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// SubscriptionWithPlan holds a subscription along with its associated plan.
type SubscriptionWithPlan struct {
	Subscription *entity.Subscription
	Plan         *entity.Plan
}

// GetActiveByUserID returns the user's active subscription with plan info.
func (s *Service) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*SubscriptionWithPlan, *apperror.AppError) {
	sub, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		s.log.Error("getting active subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return nil, nil
	}

	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		s.log.Error("getting plan for subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}

	return &SubscriptionWithPlan{
		Subscription: sub,
		Plan:         plan,
	}, nil
}
