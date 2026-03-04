package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminCreateSubscriptionInput holds the data required for an admin to create a subscription.
type AdminCreateSubscriptionInput struct {
	UserID uuid.UUID
	PlanID uuid.UUID
}

// AdminCreateSubscription creates a new subscription for a user (admin operation).
func (s *Service) AdminCreateSubscription(ctx context.Context, input AdminCreateSubscriptionInput) (*entity.Subscription, *apperror.AppError) {
	existing, err := s.subscriptionRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		s.log.Error("checking existing subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check subscription", err)
	}
	if existing != nil && existing.IsAccessible() {
		return nil, apperror.Conflict(apperror.CodeAlreadySubscribed, "user already has an active subscription", domain.ErrAlreadySubscribed)
	}

	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.PlanID)
	if err != nil || plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	now := time.Now()
	sub := &entity.Subscription{
		ID:                 uuid.New(),
		UserID:             input.UserID,
		PlanID:             plan.ID,
		Plan:               plan,
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.subscriptionRepo.Create(ctx, sub); err != nil {
		s.log.Error("admin creating subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create subscription", err)
	}
	return sub, nil
}
