package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

var planFilterConfig = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"name":      {DBColumn: "p.name", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
		"is_active": {DBColumn: "p.is_active", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeBool},
	},
	SortColumns: map[string]filter.SortDef{
		"name":       {DBColumn: "p.name"},
		"level":      {DBColumn: "p.level"},
		"created_at": {DBColumn: "p.created_at"},
	},
	DefaultSort: "p.level ASC",
}

// PlanFilterConfig returns the filter configuration for plans.
func PlanFilterConfig() filter.Config {
	return planFilterConfig
}

// PlanRepository implements repository.PlanRepository using PostgreSQL.
type PlanRepository struct {
	pool *pgxpool.Pool
}

// NewPlanRepository creates a new PlanRepository.
func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

func (r *PlanRepository) Create(ctx context.Context, plan *entity.Plan) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO plans (id, name, slug, description, price, level, stripe_price_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, plan.ID, plan.Name, plan.Slug.String(), plan.Description, plan.Price,
		plan.Level, nullableString(plan.StripePriceID), plan.IsActive, plan.CreatedAt, plan.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating plan: %w", err)
	}

	return nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	db := connFromCtx(ctx, r.pool)

	var p entity.Plan
	var slug string
	var stripePriceID sql.NullString

	err := db.QueryRow(ctx, `
		SELECT id, name, slug, description, price, level, stripe_price_id, is_active, created_at, updated_at
		FROM plans WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &slug, &p.Description, &p.Price,
		&p.Level, &stripePriceID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting plan by id: %w", err)
	}

	p.Slug = valueobject.SlugFromTrusted(slug)
	if stripePriceID.Valid {
		p.StripePriceID = stripePriceID.String
	}
	return &p, nil
}

func (r *PlanRepository) GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Plan, error) {
	db := connFromCtx(ctx, r.pool)

	var p entity.Plan
	var slugStr string
	var stripePriceID sql.NullString

	err := db.QueryRow(ctx, `
		SELECT id, name, slug, description, price, level, stripe_price_id, is_active, created_at, updated_at
		FROM plans WHERE slug = $1
	`, slug.String()).Scan(&p.ID, &p.Name, &slugStr, &p.Description, &p.Price,
		&p.Level, &stripePriceID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting plan by slug: %w", err)
	}

	p.Slug = valueobject.SlugFromTrusted(slugStr)
	if stripePriceID.Valid {
		p.StripePriceID = stripePriceID.String
	}
	return &p, nil
}

func (r *PlanRepository) Update(ctx context.Context, plan *entity.Plan) error {
	db := connFromCtx(ctx, r.pool)

	_, err := db.Exec(ctx, `
		UPDATE plans SET name = $2, slug = $3, description = $4, price = $5,
		       level = $6, stripe_price_id = $7, is_active = $8, updated_at = $9
		WHERE id = $1
	`, plan.ID, plan.Name, plan.Slug.String(), plan.Description, plan.Price,
		plan.Level, nullableString(plan.StripePriceID), plan.IsActive, plan.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating plan: %w", err)
	}

	return nil
}

func (r *PlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM plans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) List(ctx context.Context, fs *filter.FilterSet) ([]entity.Plan, int64, error) {
	db := connFromCtx(ctx, r.pool)

	// Count
	countWhere, countArgs, _ := filter.CountSQL(fs, "", 1)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM plans p %s`, countWhere)

	var total int64
	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting plans: %w", err)
	}

	// Data
	queryParts := filter.ToSQL(fs, "", 1)
	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, price, level, stripe_price_id, is_active, created_at, updated_at
		FROM plans p
		%s %s %s
	`, queryParts.WhereClause, queryParts.OrderClause, queryParts.LimitClause)

	rows, err := db.Query(ctx, dataQuery, queryParts.Args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing plans: %w", err)
	}
	defer rows.Close()

	var plans []entity.Plan
	for rows.Next() {
		var p entity.Plan
		var slug string
		var stripePriceID sql.NullString

		if err := rows.Scan(&p.ID, &p.Name, &slug, &p.Description, &p.Price,
			&p.Level, &stripePriceID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning plan: %w", err)
		}
		p.Slug = valueobject.SlugFromTrusted(slug)
		if stripePriceID.Valid {
			p.StripePriceID = stripePriceID.String
		}
		plans = append(plans, p)
	}

	return plans, total, nil
}

func (r *PlanRepository) ListActive(ctx context.Context) ([]entity.Plan, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `
		SELECT id, name, slug, description, price, level, stripe_price_id, is_active, created_at, updated_at
		FROM plans WHERE is_active = true
		ORDER BY level ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("listing active plans: %w", err)
	}
	defer rows.Close()

	var plans []entity.Plan
	for rows.Next() {
		var p entity.Plan
		var slug string
		var stripePriceID sql.NullString

		if err := rows.Scan(&p.ID, &p.Name, &slug, &p.Description, &p.Price,
			&p.Level, &stripePriceID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning active plan: %w", err)
		}
		p.Slug = valueobject.SlugFromTrusted(slug)
		if stripePriceID.Valid {
			p.StripePriceID = stripePriceID.String
		}
		plans = append(plans, p)
	}

	return plans, nil
}

func (r *PlanRepository) ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM plans WHERE slug = $1)`, slug.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking plan slug existence: %w", err)
	}
	return exists, nil
}
