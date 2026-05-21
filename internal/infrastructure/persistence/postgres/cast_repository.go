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
		INSERT INTO casts (id, name, picture_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, cast.ID, cast.Name, nullableString(cast.PictureKey), cast.CreatedAt, cast.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating cast: %w", err)
	}
	return nil
}

func (r *CastRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error) {
	db := connFromCtx(ctx, r.pool)
	return scanCast(db.QueryRow(ctx, `
		SELECT id, name, picture_key, created_at, updated_at
		FROM casts WHERE id = $1
	`, id))
}

func (r *CastRepository) GetByName(ctx context.Context, name string) (*entity.Cast, error) {
	db := connFromCtx(ctx, r.pool)
	return scanCast(db.QueryRow(ctx, `
		SELECT id, name, picture_key, created_at, updated_at
		FROM casts WHERE name = $1
	`, name))
}

// FindOrCreate returns the existing cast member for name, or inserts one.
// The INSERT is a no-op on conflict so a concurrent caller's row wins.
func (r *CastRepository) FindOrCreate(ctx context.Context, name string) (*entity.Cast, error) {
	db := connFromCtx(ctx, r.pool)
	now := time.Now().UTC()
	id := uuid.New()

	// Attempt insert; skip silently if the unique name already exists.
	_, err := db.Exec(ctx, `
		INSERT INTO casts (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (name) DO NOTHING
	`, id, name, now)
	if err != nil {
		return nil, fmt.Errorf("upserting cast: %w", err)
	}

	return scanCast(db.QueryRow(ctx, `
		SELECT id, name, picture_key, created_at, updated_at
		FROM casts WHERE name = $1
	`, name))
}

func (r *CastRepository) Update(ctx context.Context, cast *entity.Cast) error {
	db := connFromCtx(ctx, r.pool)
	cast.UpdatedAt = time.Now().UTC()
	_, err := db.Exec(ctx, `
		UPDATE casts SET name = $2, picture_key = $3, updated_at = $4
		WHERE id = $1
	`, cast.ID, cast.Name, nullableString(cast.PictureKey), cast.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating cast: %w", err)
	}
	return nil
}

func (r *CastRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM casts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting cast: %w", err)
	}
	return nil
}

// HasVideoWithCaller returns true when castID is credited on at least one
// video whose uploaded_by equals callerID.
func (r *CastRepository) HasVideoWithCaller(ctx context.Context, castID, callerID uuid.UUID) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM video_cast vc
			JOIN videos v ON v.id = vc.video_id
			WHERE vc.cast_id = $1 AND v.uploaded_by = $2
		)
	`, castID, callerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking cast ownership: %w", err)
	}
	return exists, nil
}

// scanCast reads a single row from the casts table.
func scanCast(row pgx.Row) (*entity.Cast, error) {
	c := &entity.Cast{}
	var pictureKey *string
	err := row.Scan(&c.ID, &c.Name, &pictureKey, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning cast: %w", err)
	}
	if pictureKey != nil {
		c.PictureKey = *pictureKey
	}
	return c, nil
}
