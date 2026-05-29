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

// SeriesWatchProgressRepository implements repository.SeriesWatchProgressRepository using PostgreSQL.
type SeriesWatchProgressRepository struct {
	pool *pgxpool.Pool
}

// NewSeriesWatchProgressRepository creates a new SeriesWatchProgressRepository.
func NewSeriesWatchProgressRepository(pool *pgxpool.Pool) *SeriesWatchProgressRepository {
	return &SeriesWatchProgressRepository{pool: pool}
}

func (r *SeriesWatchProgressRepository) Upsert(ctx context.Context, userID, seriesID, lastEpisodeID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO series_watch_progress (user_id, series_id, last_episode_id, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, series_id) DO UPDATE
		SET last_episode_id = EXCLUDED.last_episode_id,
		    updated_at      = NOW()
	`, userID, seriesID, lastEpisodeID)
	if err != nil {
		return fmt.Errorf("upserting series watch progress: %w", err)
	}
	return nil
}

func (r *SeriesWatchProgressRepository) Get(ctx context.Context, userID, seriesID uuid.UUID) (*entity.SeriesWatchProgress, error) {
	db := connFromCtx(ctx, r.pool)
	var p entity.SeriesWatchProgress
	err := db.QueryRow(ctx, `
		SELECT user_id, series_id, last_episode_id, updated_at
		FROM series_watch_progress
		WHERE user_id = $1 AND series_id = $2
	`, userID, seriesID).Scan(&p.UserID, &p.SeriesID, &p.LastEpisodeID, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting series watch progress: %w", err)
	}
	return &p, nil
}

func (r *SeriesWatchProgressRepository) Delete(ctx context.Context, userID, seriesID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		DELETE FROM series_watch_progress WHERE user_id = $1 AND series_id = $2
	`, userID, seriesID)
	if err != nil {
		return fmt.Errorf("deleting series watch progress: %w", err)
	}
	return nil
}
