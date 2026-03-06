package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminGetUserSubscription returns the active subscription for the given user (admin view).
func (s *Service) AdminGetUserSubscription(ctx context.Context, userID uuid.UUID) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		s.log.Error("getting user subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "no active subscription", domain.ErrNoActiveSubscription)
	}
	return sub, nil
}
