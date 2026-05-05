package subscription

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

// AssignInput holds the data needed to assign a subscription.
type AssignInput struct {
	UserID    uuid.UUID
	PlanID    uuid.UUID
	ExpiresAt *time.Time
}

// Assign creates a new subscription for a user.
func (s *Service) Assign(ctx context.Context, input AssignInput) (*entity.Subscription, *apperror.AppError) {
	// Check plan exists and is active
	plan, err := s.planRepo.GetByID(ctx, input.PlanID)
	if err != nil {
		s.log.Error("getting plan for subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}
	if plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}
	if !plan.IsActive {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "plan is not active", nil)
	}

	// Check for existing active subscription
	existing, err := s.subRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		s.log.Error("checking existing subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check existing subscription", err)
	}
	if existing != nil {
		return nil, apperror.Conflict(apperror.CodeActiveSubscriptionExists, "user already has an active subscription", domain.ErrActiveSubscriptionExists)
	}

	sub := entity.NewSubscription(input.UserID, input.PlanID, input.ExpiresAt)

	if err := s.subRepo.Create(ctx, sub); err != nil {
		s.log.Error("creating subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create subscription", err)
	}

	if err := s.dispatcher.Dispatch(event.SubscriptionAssigned{
		SubscriptionID: sub.ID,
		UserID:         sub.UserID,
		PlanID:         sub.PlanID,
		ActorID:        actor.FromContext(ctx),
		Timestamp:      time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching subscription assigned event", "error", err)
	}

	return sub, nil
}
