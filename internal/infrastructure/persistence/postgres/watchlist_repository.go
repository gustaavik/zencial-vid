package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// WatchlistRepository implements repository.WatchlistRepository using PostgreSQL.
type WatchlistRepository struct {
	pool      *pgxpool.Pool
	videoRepo *VideoRepository
}

// NewWatchlistRepository creates a new WatchlistRepository.
//
// The video repository is required to load genre IDs for each listed video,
// reusing the existing GetGenreIDs query.
func NewWatchlistRepository(pool *pgxpool.Pool, videoRepo *VideoRepository) *WatchlistRepository {
	return &WatchlistRepository{pool: pool, videoRepo: videoRepo}
}

// Add inserts an entry. Idempotent — repeated adds are no-ops.
func (r *WatchlistRepository) Add(ctx context.Context, userID, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO user_watchlist (user_id, video_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, video_id) DO NOTHING
	`, userID, videoID)
	if err != nil {
		return fmt.Errorf("adding watchlist entry: %w", err)
	}
	return nil
}

// Remove deletes the entry. Returns domain.ErrWatchlistEntryNotFound when no row was deleted.
func (r *WatchlistRepository) Remove(ctx context.Context, userID, videoID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	tag, err := db.Exec(ctx, `
		DELETE FROM user_watchlist WHERE user_id = $1 AND video_id = $2
	`, userID, videoID)
	if err != nil {
		return fmt.Errorf("removing watchlist entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWatchlistEntryNotFound
	}
	return nil
}

// Exists reports whether a watchlist entry is present for (userID, videoID).
func (r *WatchlistRepository) Exists(ctx context.Context, userID, videoID uuid.UUID) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_watchlist WHERE user_id = $1 AND video_id = $2)
	`, userID, videoID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking watchlist entry: %w", err)
	}
	return exists, nil
}

// ListVideos returns the user's watchlist as full Video entities, ordered by added_at DESC.
func (r *WatchlistRepository) ListVideos(ctx context.Context, userID uuid.UUID, page valueobject.Pagination) ([]entity.Video, int64, error) {
	db := connFromCtx(ctx, r.pool)

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_watchlist WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting watchlist: %w", err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	rows, err := db.Query(ctx, `
		SELECT v.id, v.title, v.slug, v.description, v.creator, v.duration, v.content_rating,
		       v.status, v.storage_key, v.content_type, v.file_size, v.thumbnail_key, v.uploaded_by,
		       v.minimum_plan_level, v.transcode_error, v.created_at, v.updated_at
		FROM videos v
		INNER JOIN user_watchlist w ON w.video_id = v.id
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC
		LIMIT $2 OFFSET $3
	`, userID, page.Limit(), page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("listing watchlist videos: %w", err)
	}
	defer rows.Close()

	var videos []entity.Video
	for rows.Next() {
		v, err := r.videoRepo.scanVideoRow(rows)
		if err != nil {
			return nil, 0, err
		}
		videos = append(videos, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating watchlist videos: %w", err)
	}

	for i := range videos {
		genreIDs, err := r.videoRepo.GetGenreIDs(ctx, videos[i].ID)
		if err != nil {
			return nil, 0, err
		}
		videos[i].GenreIDs = genreIDs
	}

	return videos, total, nil
}
