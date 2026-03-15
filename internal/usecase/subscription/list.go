package subscription

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListByUserID returns all subscriptions for a user.
func (s *Service) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, *apperror.AppError) {
	subs, err := s.subRepo.ListByUserID(ctx, userID)
	if err != nil {
		s.log.Error("listing subscriptions", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list subscriptions", err)
	}

	return subs, nil
}
