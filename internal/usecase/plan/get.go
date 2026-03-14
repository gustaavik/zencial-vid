package plan

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetByID returns a plan by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, *apperror.AppError) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}
	if plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	return plan, nil
}
