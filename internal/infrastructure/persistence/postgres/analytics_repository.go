package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// AnalyticsRepository implements repository.AnalyticsRepository using PostgreSQL.
// All queries aggregate over the playback_sessions table; rows qualify as a
// "view" when their cumulative watched time crossed the per-video threshold
// (stored as is_view at write time, see PlaybackSessionRepository).
type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepository creates a new AnalyticsRepository.
func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

// scopeFilter renders a PlaybackScope as additional WHERE conditions.
// $1 and $2 are always the window bounds; scope params start at $3.
func scopeFilter(scope repository.PlaybackScope, from, to time.Time) (where string, args []any) {
	where = "ps.started_at >= $1 AND ps.started_at < $2"
	args = []any{from, to}
	if scope.VideoID != nil {
		where += fmt.Sprintf(" AND ps.video_id = $%d", len(args)+1)
		args = append(args, *scope.VideoID)
	}
	if scope.UploaderID != nil {
		where += fmt.Sprintf(" AND v.uploaded_by = $%d", len(args)+1)
		args = append(args, *scope.UploaderID)
	}
	return where, args
}

func (r *AnalyticsRepository) GetTotals(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) (*repository.PlaybackTotals, error) {
	db := connFromCtx(ctx, r.pool)
	where, args := scopeFilter(scope, from, to)

	totals := &repository.PlaybackTotals{}
	err := db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE ps.is_view)                                                  AS views,
			COALESCE(SUM(ps.watched_seconds), 0)                                                AS watched_seconds,
			COUNT(DISTINCT ps.user_id) FILTER (WHERE ps.is_view AND ps.user_id IS NOT NULL)     AS unique_viewers,
			COALESCE(AVG(
				CASE WHEN v.duration > 0
				     THEN LEAST(ps.max_position::float / v.duration::float * 100, 100)
				     ELSE 0 END
			) FILTER (WHERE ps.is_view), 0)                                                     AS avg_percent_watched,
			COALESCE(
				COUNT(*) FILTER (WHERE ps.is_view AND v.duration > 0
				                   AND ps.max_position::float / v.duration::float >= 0.9)::float
				/ NULLIF(COUNT(*) FILTER (WHERE ps.is_view), 0) * 100,
			0)                                                                                  AS finish_rate
		FROM playback_sessions ps
		JOIN videos v ON v.id = ps.video_id
		WHERE `+where,
		args...,
	).Scan(&totals.Views, &totals.WatchedSeconds, &totals.UniqueViewers, &totals.AvgPercentWatched, &totals.FinishRate)
	if err != nil {
		return nil, fmt.Errorf("getting playback totals: %w", err)
	}
	return totals, nil
}

func (r *AnalyticsRepository) GetDailySeries(ctx context.Context, scope repository.PlaybackScope, from, to time.Time) ([]repository.DailyStat, error) {
	db := connFromCtx(ctx, r.pool)
	where, args := scopeFilter(scope, from, to)

	rows, err := db.Query(ctx, `
		SELECT
			date_trunc('day', ps.started_at)     AS day,
			COUNT(*) FILTER (WHERE ps.is_view)   AS views,
			COALESCE(SUM(ps.watched_seconds), 0) AS watched_seconds
		FROM playback_sessions ps
		JOIN videos v ON v.id = ps.video_id
		WHERE `+where+`
		GROUP BY 1
		ORDER BY 1
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("getting daily playback series: %w", err)
	}
	defer rows.Close()

	var series []repository.DailyStat
	for rows.Next() {
		var d repository.DailyStat
		if err := rows.Scan(&d.Day, &d.Views, &d.WatchedSeconds); err != nil {
			return nil, fmt.Errorf("scanning daily stat: %w", err)
		}
		series = append(series, d)
	}
	return series, rows.Err()
}

func (r *AnalyticsRepository) GetTopVideos(ctx context.Context, uploaderID *uuid.UUID, from, to time.Time, limit int) ([]repository.VideoRollup, error) {
	db := connFromCtx(ctx, r.pool)

	where := "TRUE"
	args := []any{from, to}
	if uploaderID != nil {
		where = fmt.Sprintf("v.uploaded_by = $%d", len(args)+1)
		args = append(args, *uploaderID)
	}
	args = append(args, limit)

	rows, err := db.Query(ctx, fmt.Sprintf(`
		SELECT
			v.id, v.title, v.status,
			COUNT(ps.id) FILTER (WHERE ps.is_view)  AS views,
			COALESCE(SUM(ps.watched_seconds), 0)    AS watched_seconds,
			COALESCE(AVG(
				CASE WHEN v.duration > 0
				     THEN LEAST(ps.max_position::float / v.duration::float * 100, 100)
				     ELSE 0 END
			) FILTER (WHERE ps.is_view), 0)         AS avg_percent_watched,
			COALESCE(
				COUNT(ps.id) FILTER (WHERE ps.is_view AND v.duration > 0
				                       AND ps.max_position::float / v.duration::float >= 0.9)::float
				/ NULLIF(COUNT(ps.id) FILTER (WHERE ps.is_view), 0) * 100,
			0)                                      AS finish_rate
		FROM videos v
		LEFT JOIN playback_sessions ps
			ON ps.video_id = v.id AND ps.started_at >= $1 AND ps.started_at < $2
		WHERE %s
		GROUP BY v.id, v.title, v.status
		ORDER BY views DESC, v.created_at DESC
		LIMIT $%d
	`, where, len(args)), args...)
	if err != nil {
		return nil, fmt.Errorf("getting top videos: %w", err)
	}
	defer rows.Close()

	var results []repository.VideoRollup
	for rows.Next() {
		var v repository.VideoRollup
		var status string
		if err := rows.Scan(&v.VideoID, &v.Title, &status, &v.Views, &v.WatchedSeconds, &v.AvgPercentWatched, &v.FinishRate); err != nil {
			return nil, fmt.Errorf("scanning video rollup: %w", err)
		}
		v.Status = entity.VideoStatus(status)
		results = append(results, v)
	}
	return results, rows.Err()
}

func (r *AnalyticsRepository) GetRetention(ctx context.Context, videoID uuid.UUID, from, to time.Time) ([]float64, error) {
	db := connFromCtx(ctx, r.pool)

	rows, err := db.Query(ctx, `
		SELECT
			g.bucket,
			COALESCE(
				COUNT(*) FILTER (WHERE get_bit(ps.watched_buckets, g.bucket) = 1)::float
				/ NULLIF(COUNT(*), 0) * 100,
			0) AS pct
		FROM playback_sessions ps
		CROSS JOIN generate_series(0, 99) AS g(bucket)
		WHERE ps.video_id = $1 AND ps.is_view AND ps.started_at >= $2 AND ps.started_at < $3
		GROUP BY g.bucket
		ORDER BY g.bucket
	`, videoID, from, to)
	if err != nil {
		return nil, fmt.Errorf("getting retention curve: %w", err)
	}
	defer rows.Close()

	curve := make([]float64, entity.RetentionBuckets)
	for rows.Next() {
		var bucket int
		var pct float64
		if err := rows.Scan(&bucket, &pct); err != nil {
			return nil, fmt.Errorf("scanning retention bucket: %w", err)
		}
		if bucket >= 0 && bucket < entity.RetentionBuckets {
			curve[bucket] = pct
		}
	}
	return curve, rows.Err()
}

// breakdownColumns whitelists the categorical columns GetBreakdown may group by.
var breakdownColumns = map[repository.BreakdownDimension]string{
	repository.BreakdownSource:   "source",
	repository.BreakdownCountry:  "country_code",
	repository.BreakdownPlatform: "platform",
}

func (r *AnalyticsRepository) GetBreakdown(ctx context.Context, videoID uuid.UUID, dim repository.BreakdownDimension, from, to time.Time) ([]repository.BreakdownItem, error) {
	col, ok := breakdownColumns[dim]
	if !ok {
		return nil, fmt.Errorf("unknown breakdown dimension %q", dim)
	}

	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, fmt.Sprintf(`
		SELECT ps.%s, COUNT(*) AS views
		FROM playback_sessions ps
		WHERE ps.video_id = $1 AND ps.is_view AND ps.started_at >= $2 AND ps.started_at < $3
		GROUP BY 1
		ORDER BY 2 DESC
	`, col), videoID, from, to)
	if err != nil {
		return nil, fmt.Errorf("getting %s breakdown: %w", dim, err)
	}
	defer rows.Close()

	var items []repository.BreakdownItem
	for rows.Next() {
		var it repository.BreakdownItem
		if err := rows.Scan(&it.Key, &it.Views); err != nil {
			return nil, fmt.Errorf("scanning breakdown item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
