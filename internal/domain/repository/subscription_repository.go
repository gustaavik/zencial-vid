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
	GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entity.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)
	Update(ctx context.Context, sub *entity.Subscription) error
	Cancel(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
}
