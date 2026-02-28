package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// WatchlistRepository implements repository.WatchlistRepository.
type WatchlistRepository struct {
	pool *pgxpool.Pool
}

// NewWatchlistRepository creates a new WatchlistRepository.
func NewWatchlistRepository(pool *pgxpool.Pool) *WatchlistRepository {
	return &WatchlistRepository{pool: pool}
}

func (r *WatchlistRepository) Add(ctx context.Context, item *entity.WatchlistItem) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO watchlist_items (id, user_id, content_id, added_at)
		VALUES ($1, $2, $3, $4)
	`, item.ID, item.UserID, item.ContentID, item.AddedAt)
	if err != nil {
		return fmt.Errorf("adding to watchlist: %w", err)
	}
	return nil
}

func (r *WatchlistRepository) Remove(ctx context.Context, userID, contentID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM watchlist_items WHERE user_id = $1 AND content_id = $2`, userID, contentID)
	if err != nil {
		return fmt.Errorf("removing from watchlist: %w", err)
	}
	return nil
}

func (r *WatchlistRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]entity.WatchlistItem, int64, error) {
	db := connFromCtx(ctx, r.pool)
	offset := (page - 1) * perPage

	var total int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM watchlist_items WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting watchlist: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT id, user_id, content_id, added_at
		FROM watchlist_items WHERE user_id = $1
		ORDER BY added_at DESC LIMIT $2 OFFSET $3
	`, userID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing watchlist: %w", err)
	}
	defer rows.Close()

	var items []entity.WatchlistItem
	for rows.Next() {
		var item entity.WatchlistItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ContentID, &item.AddedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning watchlist item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *WatchlistRepository) Exists(ctx context.Context, userID, contentID uuid.UUID) (bool, error) {
	db := connFromCtx(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM watchlist_items WHERE user_id = $1 AND content_id = $2)`,
		userID, contentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking watchlist: %w", err)
	}
	return exists, nil
}
