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

// VideoCastRepository implements repository.VideoCastRepository using PostgreSQL.
type VideoCastRepository struct {
	pool *pgxpool.Pool
}

// NewVideoCastRepository creates a new VideoCastRepository.
func NewVideoCastRepository(pool *pgxpool.Pool) *VideoCastRepository {
	return &VideoCastRepository{pool: pool}
}

func (r *VideoCastRepository) Create(ctx context.Context, vc *entity.VideoCast) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO video_cast (id, video_id, cast_id, role, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, vc.ID, vc.VideoID, vc.CastID, vc.Role, vc.SortOrder, vc.CreatedAt, vc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating video cast: %w", err)
	}
	return nil
}

func (r *VideoCastRepository) GetByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) (*entity.VideoCast, error) {
	db := connFromCtx(ctx, r.pool)
	return scanVideoCast(db.QueryRow(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.sort_order,
			vc.created_at, vc.updated_at,
			c.id, c.name, c.picture_key, c.created_at, c.updated_at
		FROM video_cast vc
		JOIN casts c ON c.id = vc.cast_id
		WHERE vc.video_id = $1 AND vc.cast_id = $2
	`, videoID, castID))
}

func (r *VideoCastRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT
			vc.id, vc.video_id, vc.cast_id, vc.role, vc.sort_order,
			vc.created_at, vc.updated_at,
			c.id, c.name, c.picture_key, c.created_at, c.updated_at
		FROM video_cast vc
		JOIN casts c ON c.id = vc.cast_id
		WHERE vc.video_id = $1
		ORDER BY vc.sort_order ASC, vc.created_at ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing video cast: %w", err)
	}
	defer rows.Close()

	var results []entity.VideoCast
	for rows.Next() {
		vc, err := scanVideoCastRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *vc)
	}
	return results, rows.Err()
}

func (r *VideoCastRepository) Update(ctx context.Context, vc *entity.VideoCast) error {
	db := connFromCtx(ctx, r.pool)
	vc.UpdatedAt = time.Now().UTC()
	_, err := db.Exec(ctx, `
		UPDATE video_cast SET role = $3, sort_order = $4, updated_at = $5
		WHERE video_id = $1 AND cast_id = $2
	`, vc.VideoID, vc.CastID, vc.Role, vc.SortOrder, vc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating video cast: %w", err)
	}
	return nil
}

func (r *VideoCastRepository) DeleteByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		DELETE FROM video_cast WHERE video_id = $1 AND cast_id = $2
	`, videoID, castID)
	if err != nil {
		return fmt.Errorf("deleting video cast: %w", err)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVideoCast(row pgx.Row) (*entity.VideoCast, error) {
	return scanVideoCastRow(row)
}

func scanVideoCastRow(row scannable) (*entity.VideoCast, error) {
	vc := &entity.VideoCast{Cast: &entity.Cast{}}
	var pictureKey *string
	err := row.Scan(
		&vc.ID, &vc.VideoID, &vc.CastID, &vc.Role, &vc.SortOrder,
		&vc.CreatedAt, &vc.UpdatedAt,
		&vc.Cast.ID, &vc.Cast.Name, &pictureKey, &vc.Cast.CreatedAt, &vc.Cast.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning video cast: %w", err)
	}
	if pictureKey != nil {
		vc.Cast.PictureKey = *pictureKey
	}
	return vc, nil
}
