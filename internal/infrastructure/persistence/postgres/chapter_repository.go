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

// ChapterRepository implements repository.ChapterRepository using PostgreSQL.
type ChapterRepository struct {
	pool *pgxpool.Pool
}

// NewChapterRepository creates a new ChapterRepository.
func NewChapterRepository(pool *pgxpool.Pool) *ChapterRepository {
	return &ChapterRepository{pool: pool}
}

func (r *ChapterRepository) BulkReplace(ctx context.Context, videoID uuid.UUID, chapters []entity.Chapter) error {
	db := connFromCtx(ctx, r.pool)

	if _, err := db.Exec(ctx, `DELETE FROM chapters WHERE video_id = $1`, videoID); err != nil {
		return fmt.Errorf("clearing chapters: %w", err)
	}

	now := time.Now().UTC()
	for i := range chapters {
		chapters[i].VideoID = videoID
		if chapters[i].ID == uuid.Nil {
			chapters[i].ID = uuid.New()
		}
		chapters[i].CreatedAt = now
		chapters[i].UpdatedAt = now

		_, err := db.Exec(ctx, `
			INSERT INTO chapters (id, video_id, start_time_secs, title, source, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, chapters[i].ID, chapters[i].VideoID, chapters[i].StartTimeSecs,
			chapters[i].Title, string(chapters[i].Source),
			chapters[i].CreatedAt, chapters[i].UpdatedAt)
		if err != nil {
			return fmt.Errorf("inserting chapter: %w", err)
		}
	}
	return nil
}

func (r *ChapterRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Chapter, error) {
	db := connFromCtx(ctx, r.pool)
	var c entity.Chapter
	var source string
	err := db.QueryRow(ctx, `
		SELECT id, video_id, start_time_secs, title, source, created_at, updated_at
		FROM chapters WHERE id = $1
	`, id).Scan(&c.ID, &c.VideoID, &c.StartTimeSecs, &c.Title, &source, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting chapter: %w", err)
	}
	c.Source = entity.ChapterSource(source)
	return &c, nil
}

func (r *ChapterRepository) ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Chapter, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, video_id, start_time_secs, title, source, created_at, updated_at
		FROM chapters WHERE video_id = $1
		ORDER BY start_time_secs ASC
	`, videoID)
	if err != nil {
		return nil, fmt.Errorf("listing chapters: %w", err)
	}
	defer rows.Close()

	var chapters []entity.Chapter
	for rows.Next() {
		var c entity.Chapter
		var source string
		if err := rows.Scan(&c.ID, &c.VideoID, &c.StartTimeSecs, &c.Title, &source, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning chapter: %w", err)
		}
		c.Source = entity.ChapterSource(source)
		chapters = append(chapters, c)
	}
	return chapters, rows.Err()
}

func (r *ChapterRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM chapters WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting chapter: %w", err)
	}
	return nil
}

func (r *ChapterRepository) DeleteByVideo(ctx context.Context, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM chapters WHERE video_id = $1`, videoID)
	if err != nil {
		return fmt.Errorf("deleting chapters for video: %w", err)
	}
	return nil
}
