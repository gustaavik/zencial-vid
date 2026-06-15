package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// PlaybackHeartbeat is one cumulative playback report for a session. All
// numeric fields and WatchedBuckets are cumulative for the session, so
// applying the same heartbeat twice is a no-op.
type PlaybackHeartbeat struct {
	SessionID            uuid.UUID
	VideoID              uuid.UUID
	UserID               *uuid.UUID
	Source               entity.PlaybackSource
	Platform             entity.PlaybackPlatform
	CountryCode          string
	PositionSeconds      int64
	WatchedSeconds       int64
	ViewThresholdSeconds int64
	WatchedBuckets       []int
	Completed            bool
}

// PlaybackSessionRepository persists playback sessions accumulated from heartbeats.
type PlaybackSessionRepository interface {
	UpsertHeartbeat(ctx context.Context, hb *PlaybackHeartbeat) error
}
