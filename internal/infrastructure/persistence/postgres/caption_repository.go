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

// CaptionRepository implements repository.CaptionRepository using PostgreSQL.
type CaptionRepository struct {
	pool *pgxpool.Pool
}

// NewCaptionRepository creates a new CaptionRepository.
func NewCaptionRepository(pool *pgxpool.Pool) *CaptionRepository {
	return &CaptionRepository{pool: pool}
}

func (r *CaptionRepository) Upsert(ctx context.Context, caption *entity.Caption) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO captions (id, video_id, language_code, format, storage_key, status, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (video_id, language_code) DO UPDATE SET
			format = EXCLUDED.format,
			storage_key = EXCLUDED.storage_key,
			status = EXCLUDED.status,
			source = EXCLUDED.source,
			updated_at = EXCLUDED.updated_at
	`, caption.ID, caption.VideoID, caption.LanguageCode, caption.Format,
		caption.StorageKey, string(caption.Status), string(caption.Source),
		caption.CreatedAt, caption.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting caption: %w", err)
	}
	return nil
}

func (r *CaptionRepository) GetByVideoAndLang(ctx context.Context, videoID uuid.UUID, languageCode string) (*entity.Caption, error) {
	db := connFromCtx(ctx, r.pool)
	var c entity.Caption
	var status, source string
	err := db.QueryRow(ctx, `
		SELECT id, video_id, language_code, format, storage_key, status, source, created_at, updated_at
		FROM captions WHERE video_id = $1 AND language_code = $2
	`, videoID, languageCode).Scan(
		&c.ID, &c.VideoID, &c.LanguageCode, &c.Format, &c.StorageKey,
		&status, &source, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting caption: %w", err)
	}
	c.Status = entity.CaptionStatus(status)
	c.Source = entity.CaptionSource(source)
	return &c, nil
}

func (r *CaptionRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Caption, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, video_id, language_code, format, storage_key, status, source, created_at, updated_at
		FROM captions WHERE video_id = $1
		ORDER BY language_code ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing captions: %w", err)
	}
	defer rows.Close()

	var captions []entity.Caption
	for rows.Next() {
		var c entity.Caption
		var status, source string
		if err := rows.Scan(&c.ID, &c.VideoID, &c.LanguageCode, &c.Format, &c.StorageKey,
			&status, &source, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning caption: %w", err)
		}
		c.Status = entity.CaptionStatus(status)
		c.Source = entity.CaptionSource(source)
		captions = append(captions, c)
	}
	return captions, rows.Err()
}

func (r *CaptionRepository) Update(ctx context.Context, caption *entity.Caption) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		UPDATE captions SET status = $2, storage_key = $3, updated_at = $4
		WHERE id = $1
	`, caption.ID, string(caption.Status), caption.StorageKey, caption.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating caption: %w", err)
	}
	return nil
}

func (r *CaptionRepository) DeleteByVideoAndLang(ctx context.Context, videoID uuid.UUID, languageCode string) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM captions WHERE video_id = $1 AND language_code = $2`, videoID, languageCode)
	if err != nil {
		return fmt.Errorf("deleting caption: %w", err)
	}
	return nil
}
