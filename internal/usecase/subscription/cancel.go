package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Cancel cancels the active subscription for the given user.
func (s *Service) Cancel(ctx context.Context, userID uuid.UUID) *apperror.AppError {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		s.log.Error("getting subscription for cancel", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return apperror.NotFound(apperror.CodeNoActiveSubscription, "no active subscription", domain.ErrNoActiveSubscription)
	}

	sub.Cancel()
	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("canceling subscription", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to cancel subscription", err)
	}
	return nil
}
