package plan

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// List returns a paginated list of all plans (admin).
func (s *Service) List(ctx context.Context, fs filter.FilterSet) ([]entity.Plan, int64, *apperror.AppError) {
	plans, total, err := s.planRepo.List(ctx, fs)
	if err != nil {
		s.log.Error("listing plans", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list plans", err)
	}

	return plans, total, nil
}

// ListActive returns all active plans (public).
func (s *Service) ListActive(ctx context.Context) ([]entity.Plan, *apperror.AppError) {
	plans, err := s.planRepo.ListActive(ctx)
	if err != nil {
		s.log.Error("listing active plans", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list active plans", err)
	}

	return plans, nil
}
