package subscription

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListPlans returns all available subscription plans.
func (s *Service) ListPlans(ctx context.Context) ([]entity.Plan, *apperror.AppError) {
	plans, err := s.subscriptionRepo.ListPlans(ctx)
	if err != nil {
		s.log.Error("listing plans", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list plans", err)
	}
	return plans, nil
}
