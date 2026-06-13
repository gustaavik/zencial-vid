package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// PlaybackSessionRepository implements repository.PlaybackSessionRepository using PostgreSQL.
type PlaybackSessionRepository struct {
	pool *pgxpool.Pool
}

// NewPlaybackSessionRepository creates a new PlaybackSessionRepository.
func NewPlaybackSessionRepository(pool *pgxpool.Pool) *PlaybackSessionRepository {
	return &PlaybackSessionRepository{pool: pool}
}

// UpsertHeartbeat inserts or merges a cumulative playback heartbeat into its
// session row. Merging uses GREATEST/bit-OR semantics so duplicated or
// reordered heartbeats are harmless. watched_seconds is additionally capped by
// the session's wall-clock age (+60s slack) so a client cannot fabricate
// watch time. If the session ID exists but belongs to another video or user,
// the update is a guarded no-op and domain.ErrPlaybackSessionConflict is returned.
func (r *PlaybackSessionRepository) UpsertHeartbeat(ctx context.Context, hb *repository.PlaybackHeartbeat) error {
	db := connFromCtx(ctx, r.pool)

	tag, err := db.Exec(ctx, `
		INSERT INTO playback_sessions (
			id, video_id, user_id, source, platform, country_code,
			last_position, max_position, watched_seconds, watched_buckets, is_view, completed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, $9::bit(100), $8 >= $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			last_event_at   = NOW(),
			last_position   = EXCLUDED.last_position,
			max_position    = GREATEST(playback_sessions.max_position, EXCLUDED.max_position),
			watched_seconds = LEAST(
				GREATEST(playback_sessions.watched_seconds, EXCLUDED.watched_seconds),
				CEIL(EXTRACT(EPOCH FROM (NOW() - playback_sessions.started_at)))::bigint + 60
			),
			watched_buckets = playback_sessions.watched_buckets | EXCLUDED.watched_buckets,
			is_view         = playback_sessions.is_view OR EXCLUDED.is_view
				OR GREATEST(playback_sessions.watched_seconds, EXCLUDED.watched_seconds) >= $10,
			completed       = playback_sessions.completed OR EXCLUDED.completed
		WHERE playback_sessions.video_id = EXCLUDED.video_id
		  AND playback_sessions.user_id IS NOT DISTINCT FROM EXCLUDED.user_id
	`,
		hb.SessionID, hb.VideoID, hb.UserID, string(hb.Source), string(hb.Platform), hb.CountryCode,
		hb.PositionSeconds, hb.WatchedSeconds, bucketsToBitString(hb.WatchedBuckets),
		hb.ViewThresholdSeconds, hb.Completed,
	)
	if err != nil {
		return fmt.Errorf("upserting playback session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPlaybackSessionConflict
	}
	return nil
}

// bucketsToBitString renders a cumulative bucket set as a fixed-width bit
// string ('0'/'1' characters), bit 0 leftmost, matching PostgreSQL's bit(n)
// ordering and get_bit indexing. Out-of-range buckets are ignored.
func bucketsToBitString(buckets []int) string {
	bits := []byte(strings.Repeat("0", entity.RetentionBuckets))
	for _, b := range buckets {
		if b >= 0 && b < entity.RetentionBuckets {
			bits[b] = '1'
		}
	}
	return string(bits)
}
