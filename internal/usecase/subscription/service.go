package subscription

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

type Service struct {
	subscriptionRepo repository.SubscriptionRepository
	log              *slog.Logger
}

func NewService(subscriptionRepo repository.SubscriptionRepository, log *slog.Logger) *Service {
	return &Service{subscriptionRepo: subscriptionRepo, log: log}
}

func (s *Service) ListPlans(ctx context.Context) ([]entity.Plan, *apperror.AppError) {
	plans, err := s.subscriptionRepo.ListPlans(ctx)
	if err != nil {
		s.log.Error("listing plans", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list plans", err)
	}
	return plans, nil
}

func (s *Service) GetCurrent(ctx context.Context, userID uuid.UUID) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		s.log.Error("getting subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get subscription", err)
	}
	if sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "no active subscription", domain.ErrNoActiveSubscription)
	}
	return sub, nil
}

type SubscribeInput struct {
	UserID uuid.UUID
	PlanID uuid.UUID
}

func (s *Service) Subscribe(ctx context.Context, input SubscribeInput) (*entity.Subscription, *apperror.AppError) {
	existing, err := s.subscriptionRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		s.log.Error("checking existing subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check subscription", err)
	}
	if existing != nil && existing.IsAccessible() {
		return nil, apperror.Conflict(apperror.CodeAlreadySubscribed, "already has an active subscription", domain.ErrAlreadySubscribed)
	}

	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.PlanID)
	if err != nil || plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	now := time.Now()
	sub := &entity.Subscription{
		ID:                 uuid.New(),
		UserID:             input.UserID,
		PlanID:             plan.ID,
		Plan:               plan,
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.subscriptionRepo.Create(ctx, sub); err != nil {
		s.log.Error("creating subscription", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create subscription", err)
	}

	return sub, nil
}

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

type ChangePlanInput struct {
	UserID uuid.UUID
	PlanID uuid.UUID
}

func (s *Service) ChangePlan(ctx context.Context, input ChangePlanInput) (*entity.Subscription, *apperror.AppError) {
	sub, err := s.subscriptionRepo.GetActiveByUserID(ctx, input.UserID)
	if err != nil || sub == nil {
		return nil, apperror.NotFound(apperror.CodeNoActiveSubscription, "no active subscription", domain.ErrNoActiveSubscription)
	}

	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.PlanID)
	if err != nil || plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", domain.ErrPlanNotFound)
	}

	sub.PlanID = plan.ID
	sub.Plan = plan
	sub.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.log.Error("changing plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to change plan", err)
	}
	return sub, nil
}
