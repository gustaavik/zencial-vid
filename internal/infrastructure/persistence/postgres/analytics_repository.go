package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// AnalyticsRepository implements repository.AnalyticsRepository using PostgreSQL.
// All queries derive statistics from the existing user_watch_progress and videos tables;
// no separate analytics schema is required.
type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepository creates a new AnalyticsRepository.
func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) GetVideoStats(ctx context.Context, videoID uuid.UUID) (*repository.VideoStats, error) {
	db := connFromCtx(ctx, r.pool)

	stats := &repository.VideoStats{VideoID: videoID}
	err := db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT wp.user_id)                                                          AS total_viewers,
			COALESCE(AVG(
				CASE WHEN v.duration > 0
				     THEN LEAST(wp.position_seconds::float / v.duration::float * 100, 100)
				     ELSE 0 END
			), 0)                                                                               AS avg_progress_pct,
			COALESCE(
				COUNT(DISTINCT CASE
					WHEN v.duration > 0 AND wp.position_seconds::float / v.duration::float >= 0.9
					THEN wp.user_id END
				)::float / NULLIF(COUNT(DISTINCT wp.user_id), 0) * 100,
			0)                                                                                  AS completion_rate
		FROM user_watch_progress wp
		JOIN videos v ON v.id = wp.video_id
		WHERE wp.video_id = $1
	`, videoID).Scan(&stats.TotalViewers, &stats.AvgProgressPct, &stats.CompletionRate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return stats, nil
		}
		return nil, fmt.Errorf("getting video stats: %w", err)
	}
	return stats, nil
}

func (r *AnalyticsRepository) GetUploaderSummary(ctx context.Context, uploaderID uuid.UUID) ([]repository.VideoStats, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `
		SELECT
			v.id                                                                                AS video_id,
			COUNT(DISTINCT wp.user_id)                                                          AS total_viewers,
			COALESCE(AVG(
				CASE WHEN v.duration > 0
				     THEN LEAST(wp.position_seconds::float / v.duration::float * 100, 100)
				     ELSE 0 END
			), 0)                                                                               AS avg_progress_pct,
			COALESCE(
				COUNT(DISTINCT CASE
					WHEN v.duration > 0 AND wp.position_seconds::float / v.duration::float >= 0.9
					THEN wp.user_id END
				)::float / NULLIF(COUNT(DISTINCT wp.user_id), 0) * 100,
			0)                                                                                  AS completion_rate
		FROM videos v
		LEFT JOIN user_watch_progress wp ON wp.video_id = v.id
		WHERE v.uploaded_by = $1
		GROUP BY v.id
		ORDER BY total_viewers DESC
	`, uploaderID)
	if err != nil {
		return nil, fmt.Errorf("getting uploader summary: %w", err)
	}
	defer rows.Close()

	var results []repository.VideoStats
	for rows.Next() {
		var s repository.VideoStats
		if err := rows.Scan(&s.VideoID, &s.TotalViewers, &s.AvgProgressPct, &s.CompletionRate); err != nil {
			return nil, fmt.Errorf("scanning video stats: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}
