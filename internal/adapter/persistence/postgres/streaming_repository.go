package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// StreamingRepository implements repository.StreamingRepository.
type StreamingRepository struct {
	pool *pgxpool.Pool
}

// NewStreamingRepository creates a new StreamingRepository.
func NewStreamingRepository(pool *pgxpool.Pool) *StreamingRepository {
	return &StreamingRepository{pool: pool}
}

func (r *StreamingRepository) CreateSession(ctx context.Context, session *entity.StreamSession) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO stream_sessions (id, user_id, content_id, episode_id, started_at, last_active_at, device_info, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, session.ID, session.UserID, session.ContentID, session.EpisodeID,
		session.StartedAt, session.LastActiveAt, session.DeviceInfo, session.IPAddress)
	if err != nil {
		return fmt.Errorf("creating stream session: %w", err)
	}
	return nil
}

func (r *StreamingRepository) GetActiveSessionsByUser(ctx context.Context, userID uuid.UUID) ([]entity.StreamSession, error) {
	db := connFromCtx(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, user_id, content_id, episode_id, started_at, last_active_at, device_info, ip_address
		FROM stream_sessions
		WHERE user_id = $1 AND last_active_at > $2
		ORDER BY started_at DESC
	`, userID, time.Now().Add(-30*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("getting active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []entity.StreamSession
	for rows.Next() {
		var s entity.StreamSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.ContentID, &s.EpisodeID,
			&s.StartedAt, &s.LastActiveAt, &s.DeviceInfo, &s.IPAddress); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *StreamingRepository) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM stream_sessions WHERE id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("ending session: %w", err)
	}
	return nil
}

func (r *StreamingRepository) SaveProgress(ctx context.Context, progress *entity.PlaybackProgress) error {
	db := connFromCtx(ctx, r.pool)
	_, err := db.Exec(ctx, `
		INSERT INTO playback_progress (id, user_id, content_id, episode_id, position, duration, completed, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, content_id, COALESCE(episode_id, '00000000-0000-0000-0000-000000000000'))
		DO UPDATE SET position = $5, duration = $6, completed = $7, updated_at = $8
	`, progress.ID, progress.UserID, progress.ContentID, progress.EpisodeID,
		progress.Position, progress.Duration, progress.Completed, time.Now())
	if err != nil {
		return fmt.Errorf("saving progress: %w", err)
	}
	return nil
}

func (r *StreamingRepository) GetProgress(ctx context.Context, userID, contentID uuid.UUID, episodeID *uuid.UUID) (*entity.PlaybackProgress, error) {
	db := connFromCtx(ctx, r.pool)
	p := &entity.PlaybackProgress{}
	var query string
	var args []interface{}

	if episodeID != nil {
		query = `SELECT id, user_id, content_id, episode_id, position, duration, completed, updated_at
		         FROM playback_progress WHERE user_id = $1 AND content_id = $2 AND episode_id = $3`
		args = []interface{}{userID, contentID, *episodeID}
	} else {
		query = `SELECT id, user_id, content_id, episode_id, position, duration, completed, updated_at
		         FROM playback_progress WHERE user_id = $1 AND content_id = $2 AND episode_id IS NULL`
		args = []interface{}{userID, contentID}
	}

	err := db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.UserID, &p.ContentID, &p.EpisodeID, &p.Position, &p.Duration, &p.Completed, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting progress: %w", err)
	}
	return p, nil
}

func (r *StreamingRepository) GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]entity.ContinueWatching, error) {
	// Simplified: returns empty for now. Full implementation would join with content.
	_ = ctx
	_ = userID
	_ = limit
	return nil, nil
}
