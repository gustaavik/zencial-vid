package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SubscriptionRepository defines persistence operations for subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *entity.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)
	Update(ctx context.Context, sub *entity.Subscription) error
	ListSubscriptions(ctx context.Context, page, perPage int) ([]entity.Subscription, int64, error)

	// Plans
	ListPlans(ctx context.Context) ([]entity.Plan, error)
	ListAllPlans(ctx context.Context) ([]entity.Plan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error)
	CreatePlan(ctx context.Context, plan *entity.Plan) error
	UpdatePlan(ctx context.Context, plan *entity.Plan) error
	DeactivatePlan(ctx context.Context, id uuid.UUID) error
}
