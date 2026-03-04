package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// SubscriptionRepository implements repository.SubscriptionRepository.
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
		INSERT INTO subscriptions (id, user_id, plan_id, status, external_id, current_period_start,
		                          current_period_end, canceled_at, trial_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.ExternalID,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CanceledAt, sub.TrialEnd,
		sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)
	s := &entity.Subscription{}
	plan := &entity.Plan{}
	err := db.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.external_id, s.current_period_start,
		       s.current_period_end, s.canceled_at, s.trial_end, s.created_at, s.updated_at,
		       p.id, p.name, p.tier, p.price_amount, p.price_currency, p.billing_interval,
		       p.max_quality, p.max_streams, p.downloads_allowed, p.is_active
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.ExternalID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CanceledAt, &s.TrialEnd,
		&s.CreatedAt, &s.UpdatedAt,
		&plan.ID, &plan.Name, &plan.Tier, &plan.Price.Amount, &plan.Price.Currency,
		&plan.BillingInterval, &plan.MaxQuality, &plan.MaxStreams,
		&plan.DownloadsAllowed, &plan.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting subscription by id: %w", err)
	}
	s.Plan = plan
	return s, nil
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	return r.GetActiveByUserID(ctx, userID)
}

func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	db := connFromCtx(ctx, r.pool)
	s := &entity.Subscription{}
	plan := &entity.Plan{}
	err := db.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.external_id, s.current_period_start,
		       s.current_period_end, s.canceled_at, s.trial_end, s.created_at, s.updated_at,
		       p.id, p.name, p.tier, p.price_amount, p.price_currency, p.billing_interval,
		       p.max_quality, p.max_streams, p.downloads_allowed, p.is_active
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status IN ('active', 'trialing', 'canceled')
		ORDER BY s.created_at DESC LIMIT 1
	`, userID).Scan(&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.ExternalID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CanceledAt, &s.TrialEnd,
		&s.CreatedAt, &s.UpdatedAt,
		&plan.ID, &plan.Name, &plan.Tier, &plan.Price.Amount, &plan.Price.Currency,
		&plan.BillingInterval, &plan.MaxQuality, &plan.MaxStreams,
		&plan.DownloadsAllowed, &plan.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting active subscription: %w", err)
	}
	s.Plan = plan
	return s, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *entity.Subscription) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE subscriptions SET plan_id=$2, status=$3, external_id=$4,
		       current_period_start=$5, current_period_end=$6, canceled_at=$7,
		       trial_end=$8, updated_at=$9
		WHERE id = $1
	`, sub.ID, sub.PlanID, sub.Status, sub.ExternalID,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CanceledAt, sub.TrialEnd, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) ListSubscriptions(ctx context.Context, page, perPage int) ([]entity.Subscription, int64, error) {
	db := connFromCtx(ctx, r.pool)
	offset := (page - 1) * perPage

	rows, err := db.Query(ctx, `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.external_id, s.current_period_start,
		       s.current_period_end, s.canceled_at, s.trial_end, s.created_at, s.updated_at,
		       p.id, p.name, p.tier, p.price_amount, p.price_currency, p.billing_interval,
		       p.max_quality, p.max_streams, p.downloads_allowed, p.is_active,
		       u.email,
		       COUNT(*) OVER() AS total
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		JOIN users u ON s.user_id = u.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []entity.Subscription
	var total int64
	for rows.Next() {
		var s entity.Subscription
		plan := &entity.Plan{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.ExternalID,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CanceledAt, &s.TrialEnd,
			&s.CreatedAt, &s.UpdatedAt,
			&plan.ID, &plan.Name, &plan.Tier, &plan.Price.Amount, &plan.Price.Currency,
			&plan.BillingInterval, &plan.MaxQuality, &plan.MaxStreams,
			&plan.DownloadsAllowed, &plan.IsActive,
			&s.UserEmail, &total); err != nil {
			return nil, 0, fmt.Errorf("scanning subscription: %w", err)
		}
		s.Plan = plan
		subs = append(subs, s)
	}
	return subs, total, nil
}

func (r *SubscriptionRepository) ListPlans(ctx context.Context) ([]entity.Plan, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, name, tier, price_amount, price_currency, billing_interval,
		       max_quality, max_streams, downloads_allowed, is_active, created_at
		FROM plans WHERE is_active = true ORDER BY price_amount
	`)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}
	defer rows.Close()

	var plans []entity.Plan
	for rows.Next() {
		var p entity.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Tier, &p.Price.Amount, &p.Price.Currency,
			&p.BillingInterval, &p.MaxQuality, &p.MaxStreams, &p.DownloadsAllowed,
			&p.IsActive, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning plan: %w", err)
		}
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	db := connFromCtx(ctx, r.pool)
	p := &entity.Plan{}
	err := db.QueryRow(ctx, `
		SELECT id, name, tier, price_amount, price_currency, billing_interval,
		       max_quality, max_streams, downloads_allowed, is_active, created_at
		FROM plans WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Tier, &p.Price.Amount, &p.Price.Currency,
		&p.BillingInterval, &p.MaxQuality, &p.MaxStreams, &p.DownloadsAllowed,
		&p.IsActive, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting plan by id: %w", err)
	}
	return p, nil
}
