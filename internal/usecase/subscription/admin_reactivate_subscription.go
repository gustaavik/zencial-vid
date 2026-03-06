package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminReactivateSubscription reactivates a canceled subscription (admin operation).
func (s *Service) AdminReactivateSubscription(ctx context.Context, subscriptionID uuid.UUID) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		s.log.Error("getting subscription for reactivation", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "subscription not found", domain.ErrNoActiveSubscription)
	}
	if sub.Status != entity.SubscriptionCanceled {
		return nil, apperror.BadRequest("INVALID_STATUS", "only canceled subscriptions can be reactivated", nil)
	}

	sub.Reactivate()

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("reactivating subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to reactivate subscription", err)
	}
	return sub, nil
}
