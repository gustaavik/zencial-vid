package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Cancel cancels a subscription by its ID.
func (s *Service) Cancel(ctx context.Context, id uuid.UUID) *apperror.AppError {
	sub, err := s.subRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting subscription for cancel", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return apperror.NotFound(apperror.CodeSubscriptionNotFound, "subscription not found", domain.ErrSubscriptionNotFound)
	}

	if err := s.subRepo.Cancel(ctx, id); err != nil {
		s.log.Error("cancelling subscription", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to cancel subscription", err)
	}

	return nil
}
