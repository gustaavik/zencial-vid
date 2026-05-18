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

// CastRepository implements repository.CastRepository using PostgreSQL.
type CastRepository struct {
	pool *pgxpool.Pool
}

// NewCastRepository creates a new CastRepository.
func NewCastRepository(pool *pgxpool.Pool) *CastRepository {
	return &CastRepository{pool: pool}
}

func (r *CastRepository) Create(ctx context.Context, cast *entity.Cast) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO video_cast (id, video_id, name, role, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, cast.ID, cast.VideoID, cast.Name, cast.Role, cast.SortOrder, cast.CreatedAt, cast.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating cast: %w", err)
	}
	return nil
}

func (r *CastRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error) {
	db := connFromCtx(ctx, r.pool)
	c := &entity.Cast{}
	err := db.QueryRow(ctx, `
		SELECT id, video_id, name, role, sort_order, created_at, updated_at
		FROM video_cast WHERE id = $1
	`, id).Scan(&c.ID, &c.VideoID, &c.Name, &c.Role, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting cast by id: %w", err)
	}
	return c, nil
}

func (r *CastRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Cast, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, video_id, name, role, sort_order, created_at, updated_at
		FROM video_cast WHERE video_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing cast by video: %w", err)
	}
	defer rows.Close()

	var results []entity.Cast
	for rows.Next() {
		var c entity.Cast
		if err := rows.Scan(&c.ID, &c.VideoID, &c.Name, &c.Role, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning cast row: %w", err)
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

func (r *CastRepository) Update(ctx context.Context, cast *entity.Cast) error {
	db := connFromCtx(ctx, r.pool)
	cast.UpdatedAt = time.Now().UTC()
	_, err := db.Exec(ctx, `
		UPDATE video_cast SET name = $2, role = $3, sort_order = $4, updated_at = $5
		WHERE id = $1
	`, cast.ID, cast.Name, cast.Role, cast.SortOrder, cast.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating cast: %w", err)
	}
	return nil
}

func (r *CastRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM video_cast WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting cast: %w", err)
	}
	return nil
}
