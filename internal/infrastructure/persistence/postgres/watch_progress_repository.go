package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// WatchProgressRepository implements repository.WatchProgressRepository using PostgreSQL.
type WatchProgressRepository struct {
	pool      *pgxpool.Pool
	videoRepo *VideoRepository
}

// NewWatchProgressRepository creates a new WatchProgressRepository.
//
// The video repository is required for ListInProgress, which reuses GetGenreIDs
// to hydrate genre associations on the joined Video rows.
func NewWatchProgressRepository(pool *pgxpool.Pool, videoRepo *VideoRepository) *WatchProgressRepository {
	return &WatchProgressRepository{pool: pool, videoRepo: videoRepo}
}

// Upsert inserts or updates progress for (userID, videoID).
func (r *WatchProgressRepository) Upsert(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int64) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO user_watch_progress (user_id, video_id, position_seconds, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, video_id) DO UPDATE
		SET position_seconds = EXCLUDED.position_seconds,
		    updated_at       = NOW()
	`, userID, videoID, positionSeconds)
	if err != nil {
		return fmt.Errorf("upserting watch progress: %w", err)
	}
	return nil
}

// Get returns progress for (userID, videoID), or (nil, nil) if no row exists.
func (r *WatchProgressRepository) Get(ctx context.Context, userID, videoID uuid.UUID) (*entity.WatchProgress, error) {
	db := connFromCtx(ctx, r.pool)
	var p entity.WatchProgress
	err := db.QueryRow(ctx, `
		SELECT user_id, video_id, position_seconds, updated_at
		FROM user_watch_progress
		WHERE user_id = $1 AND video_id = $2
	`, userID, videoID).Scan(&p.UserID, &p.VideoID, &p.PositionSeconds, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting watch progress: %w", err)
	}
	return &p, nil
}

// Delete removes the progress entry. Returns domain.ErrWatchProgressNotFound when no row was deleted.
func (r *WatchProgressRepository) Delete(ctx context.Context, userID, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	tag, err := db.Exec(ctx, `
		DELETE FROM user_watch_progress WHERE user_id = $1 AND video_id = $2
	`, userID, videoID)
	if err != nil {
		return fmt.Errorf("deleting watch progress: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWatchProgressNotFound
	}
	return nil
}

// ListInProgress returns published videos the user has started but not finished
// (position * 100 < duration * 95), ordered by most recently updated first.
//
// Integer arithmetic is used to avoid floating-point comparisons in the WHERE clause.
func (r *WatchProgressRepository) ListInProgress(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.VideoWithProgress, int64, error) {
	db := connFromCtx(ctx, r.pool)

	var total int64
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM user_watch_progress wp
		INNER JOIN videos v ON v.id = wp.video_id
		WHERE wp.user_id = $1
		  AND v.status = 'published'
		  AND wp.position_seconds * 100 < v.duration * 95
	`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting watch progress: %w", err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	rows, err := db.Query(ctx, `
		SELECT v.id, v.title, v.slug, v.description, v.creator, v.duration, v.content_rating,
		       v.status, v.storage_key, v.content_type, v.file_size, v.thumbnail_key, v.uploaded_by,
		       v.minimum_plan_level, v.transcode_error, v.created_at, v.updated_at,
		       wp.position_seconds, wp.updated_at
		FROM user_watch_progress wp
		INNER JOIN videos v ON v.id = wp.video_id
		WHERE wp.user_id = $1
		  AND v.status = 'published'
		  AND wp.position_seconds * 100 < v.duration * 95
		ORDER BY wp.updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, page.Limit(), page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("listing watch progress: %w", err)
	}
	defer rows.Close()

	var items []entity.VideoWithProgress
	for rows.Next() {
		var v entity.Video
		var slug, contentRating, status, transcodeError string
		var duration int64
		var positionSeconds int64
		var progressUpdatedAt time.Time

		if err := rows.Scan(
			&v.ID, &v.Title, &slug, &v.Description, &v.Creator,
			&duration, &contentRating, &status,
			&v.StorageKey, &v.ContentType, &v.FileSize, &v.ThumbnailKey,
			&v.UploadedBy, &v.MinimumPlanLevel, &transcodeError, &v.CreatedAt, &v.UpdatedAt,
			&positionSeconds, &progressUpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning watch progress row: %w", err)
		}

		v.Slug = valueobject.SlugFromTrusted(slug)
		v.Duration = valueobject.NewDuration(duration)
		v.ContentRating = contentRating
		v.Status = entity.VideoStatus(status)
		v.TranscodeError = transcodeError

		items = append(items, entity.VideoWithProgress{
			Video: v,
			Progress: entity.WatchProgress{
				UserID:          userID,
				VideoID:         v.ID,
				PositionSeconds: positionSeconds,
				UpdatedAt:       progressUpdatedAt,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating watch progress rows: %w", err)
	}

	for i := range items {
		genreIDs, err := r.videoRepo.GetGenreIDs(ctx, items[i].Video.ID)
		if err != nil {
			return nil, 0, err
		}
		items[i].Video.GenreIDs = genreIDs
	}

	return items, total, nil
}
