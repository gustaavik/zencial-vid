package postgres

import (
	"context"
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
		INSERT INTO user_subscriptions (id, user_id, plan_id, status, started_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, sub.ID, sub.UserID, sub.PlanID, string(sub.Status),
		sub.StartedAt, sub.ExpiresAt, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	var s entity.Subscription
	var status string
	var expiresAt *time.Time

	err := db.QueryRow(ctx, `
		SELECT id, user_id, plan_id, status, started_at, expires_at, created_at, updated_at
		FROM user_subscriptions WHERE id = $1
	`, id).Scan(&s.ID, &s.UserID, &s.PlanID, &status,
		&s.StartedAt, &expiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting subscription by id: %w", err)
	}

	s.Status = entity.SubscriptionStatus(status)
	s.ExpiresAt = expiresAt
	return &s, nil
}

func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)

	var s entity.Subscription
	var status string
	var expiresAt *time.Time

	err := db.QueryRow(ctx, `
		SELECT id, user_id, plan_id, status, started_at, expires_at, created_at, updated_at
		FROM user_subscriptions
		WHERE user_id = $1 AND status = 'active'
		LIMIT 1
	`, userID).Scan(&s.ID, &s.UserID, &s.PlanID, &status,
		&s.StartedAt, &expiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting active subscription: %w", err)
	}

	s.Status = entity.SubscriptionStatus(status)
	s.ExpiresAt = expiresAt
	return &s, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *entity.Subscription) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE user_subscriptions SET plan_id = $2, status = $3, expires_at = $4, updated_at = $5
		WHERE id = $1
	`, sub.ID, sub.PlanID, string(sub.Status), sub.ExpiresAt, sub.UpdatedAt)
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
		SELECT id, user_id, plan_id, status, started_at, expires_at, created_at, updated_at
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
		var s entity.Subscription
		var status string
		var expiresAt *time.Time

		if err := rows.Scan(&s.ID, &s.UserID, &s.PlanID, &status,
			&s.StartedAt, &expiresAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		s.Status = entity.SubscriptionStatus(status)
		s.ExpiresAt = expiresAt
		subs = append(subs, s)
	}

	return subs, nil
}
