package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminCancelSubscription cancels a subscription (admin operation).
func (s *Service) AdminCancelSubscription(ctx context.Context, subscriptionID uuid.UUID) *apperror.AppError {
	sub, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		s.log.Error("getting subscription for admin cancel", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return apperror.NotFound(apperror.CodeNoActiveSubscription, "subscription not found", domain.ErrNoActiveSubscription)
	}

	sub.Cancel()
	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("admin canceling subscription", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to cancel subscription", err)
	}
	return nil
}
