package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SubscriptionRepository implements repository.SubscriptionRepository using PostgreSQL.
type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *entity.Subscription) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO user_subscriptions (
			id, user_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
			started_at, expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, sub.ID, sub.UserID, sub.PlanID, string(sub.Status),
		nullableString(sub.StripeSubscriptionID), nullableString(sub.StripeCustomerID),
		sub.StartedAt, sub.ExpiresAt, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	row := db.QueryRow(ctx, `
		SELECT id, user_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       started_at, expires_at, created_at, updated_at
		FROM user_subscriptions WHERE id = $1
	`, id)
	sub, scanErr := scanSubscription(row)
	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting subscription by id: %w", scanErr)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	row := db.QueryRow(ctx, `
		SELECT id, user_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       started_at, expires_at, created_at, updated_at
		FROM user_subscriptions
		WHERE stripe_subscription_id = $1
		LIMIT 1
	`, stripeSubscriptionID)
	sub, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting subscription by stripe subscription id: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	row := db.QueryRow(ctx, `
		SELECT id, user_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       started_at, expires_at, created_at, updated_at
		FROM user_subscriptions
		WHERE user_id = $1 AND status = 'active'
		LIMIT 1
	`, userID)
	sub, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting active subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *entity.Subscription) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE user_subscriptions
		SET plan_id = $2, status = $3, stripe_subscription_id = $4,
		    stripe_customer_id = $5, expires_at = $6, updated_at = $7
		WHERE id = $1
	`, sub.ID, sub.PlanID, string(sub.Status), nullableString(sub.StripeSubscriptionID),
		nullableString(sub.StripeCustomerID), sub.ExpiresAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE user_subscriptions SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("cancelling subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `
		SELECT id, user_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       started_at, expires_at, created_at, updated_at
		FROM user_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing subscriptions by user: %w", err)
	}
	defer rows.Close()

	var subs []entity.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		subs = append(subs, *sub)
	}

	return subs, nil
}

type subscriptionScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(scanner subscriptionScanner) (*entity.Subscription, error) {
	var s entity.Subscription
	var status string
	var stripeSubscriptionID sql.NullString
	var stripeCustomerID sql.NullString
	var expiresAt *time.Time

	if err := scanner.Scan(&s.ID, &s.UserID, &s.PlanID, &status,
		&stripeSubscriptionID, &stripeCustomerID, &s.StartedAt, &expiresAt,
		&s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}

	s.Status = entity.SubscriptionStatus(status)
	s.ExpiresAt = expiresAt
	if stripeSubscriptionID.Valid {
		s.StripeSubscriptionID = stripeSubscriptionID.String
	}
	if stripeCustomerID.Valid {
		s.StripeCustomerID = stripeCustomerID.String
	}

	return &s, nil
}
