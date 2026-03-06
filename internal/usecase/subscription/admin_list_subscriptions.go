package subscription

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminListSubscriptions returns a paginated list of all subscriptions.
func (s *Service) AdminListSubscriptions(ctx context.Context, page, perPage int) ([]entity.Subscription, int64, *apperror.AppError) {
	subs, total, err := s.subscriptionRepo.ListSubscriptions(ctx, page, perPage)
	if err != nil {
		s.log.Error("listing subscriptions", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list subscriptions", err)
	}
	return subs, total, nil
}
