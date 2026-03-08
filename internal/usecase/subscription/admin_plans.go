package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CreatePlanInput holds data for creating a new subscription plan.
type CreatePlanInput struct {
	Name             string
	Tier             string
	PriceAmount      int64
	PriceCurrency    string
	BillingInterval  string
	MaxQuality       valueobject.VideoQuality
	MaxStreams       int
	DownloadsAllowed bool
}

// UpdatePlanInput holds data for updating an existing subscription plan.
type UpdatePlanInput struct {
	ID               uuid.UUID
	Name             string
	Tier             string
	PriceAmount      int64
	PriceCurrency    string
	BillingInterval  string
	MaxQuality       valueobject.VideoQuality
	MaxStreams       int
	DownloadsAllowed bool
	IsActive         bool
}

// AdminListAllPlans returns all plans including inactive ones.
func (s *Service) AdminListAllPlans(ctx context.Context) ([]entity.Plan, *apperror.AppError) {
	plans, err := s.subscriptionRepo.ListAllPlans(ctx)
	if err != nil {
		s.log.Error("listing all plans", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list plans", err)
	}
	return plans, nil
}

// AdminCreatePlan creates a new subscription plan.
func (s *Service) AdminCreatePlan(ctx context.Context, input CreatePlanInput) (*entity.Plan, *apperror.AppError) {
	plan := &entity.Plan{
		ID:               uuid.New(),
		Name:             input.Name,
		Tier:             entity.PlanTier(input.Tier),
		Price:            valueobject.NewMoney(input.PriceAmount, input.PriceCurrency),
		BillingInterval:  input.BillingInterval,
		MaxQuality:       input.MaxQuality,
		MaxStreams:       input.MaxStreams,
		DownloadsAllowed: input.DownloadsAllowed,
		IsActive:         true,
		CreatedAt:        time.Now(),
	}
	if err := s.subscriptionRepo.CreatePlan(ctx, plan); err != nil {
		s.log.Error("creating plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create plan", err)
	}
	return plan, nil
}

// AdminUpdatePlan updates an existing subscription plan.
func (s *Service) AdminUpdatePlan(ctx context.Context, input UpdatePlanInput) (*entity.Plan, *apperror.AppError) {
	plan, err := s.subscriptionRepo.GetPlanByID(ctx, input.ID)
	if err != nil {
		s.log.Error("fetching plan for update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, fmt.Sprintf("failed to fetch plan: %v", err), nil)
	}
	if plan == nil {
		return nil, apperror.NotFound(apperror.CodePlanNotFound, "plan not found", nil)
	}

	plan.Name = input.Name
	plan.Tier = entity.PlanTier(input.Tier)
	plan.Price = valueobject.NewMoney(input.PriceAmount, input.PriceCurrency)
	plan.BillingInterval = input.BillingInterval
	plan.MaxQuality = input.MaxQuality
	plan.MaxStreams = input.MaxStreams
	plan.DownloadsAllowed = input.DownloadsAllowed
	plan.IsActive = input.IsActive

	if err := s.subscriptionRepo.UpdatePlan(ctx, plan); err != nil {
		s.log.Error("updating plan", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update plan", err)
	}
	return plan, nil
}

// AdminDeactivatePlan deactivates a plan (soft delete).
func (s *Service) AdminDeactivatePlan(ctx context.Context, id uuid.UUID) *apperror.AppError {
	plan, err := s.subscriptionRepo.GetPlanByID(ctx, id)
	if err != nil {
		s.log.Error("fetching plan for deactivation", "error", err)
		return apperror.Internal(apperror.CodeInternalError, fmt.Sprintf("failed to fetch plan: %v", err), nil)
	}
	if plan == nil {
		return apperror.NotFound(apperror.CodePlanNotFound, "plan not found", nil)
	}
	if err := s.subscriptionRepo.DeactivatePlan(ctx, id); err != nil {
		s.log.Error("deactivating plan", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to deactivate plan", err)
	}
	return nil
}
