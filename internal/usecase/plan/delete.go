package plan

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete soft-deletes a plan by deactivating it.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("getting plan for delete", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to get plan", err)
	}
	if plan == nil {
		return apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	plan.Deactivate()
	if err := s.planRepo.Update(ctx, plan); err != nil {
		s.log.Error("deactivating plan", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to deactivate plan", err)
	}

	if err := s.dispatcher.Dispatch(event.PlanDeleted{
		PlanID:    plan.ID,
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching plan deleted event", "error", err)
	}

	return nil
}
