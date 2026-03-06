package entity

import (
	"time"

	"github.com/google/uuid"
)

// StreamSession represents an active streaming session.
type StreamSession struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ContentID    uuid.UUID
	EpisodeID    *uuid.UUID
	StartedAt    time.Time
	LastActiveAt time.Time
	DeviceInfo   string
	IPAddress    string
}

// NewStreamSession creates a new streaming session.
func NewStreamSession(userID, contentID uuid.UUID, episodeID *uuid.UUID, deviceInfo, ipAddress string) *StreamSession {
	now := time.Now()
	return &StreamSession{
		ID:           uuid.New(),
		UserID:       userID,
		ContentID:    contentID,
		EpisodeID:    episodeID,
		StartedAt:    now,
		LastActiveAt: now,
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
	}
}

// PlaybackProgress tracks a user's progress through content.
type PlaybackProgress struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ContentID uuid.UUID
	EpisodeID *uuid.UUID
	Position  int64 // Seconds from start
	Duration  int64 // Total duration in seconds
	Completed bool  // True if watched >= 90%
	UpdatedAt time.Time
}

// Percentage returns the playback progress as a percentage.
func (p *PlaybackProgress) Percentage() float64 {
	if p.Duration == 0 {
		return 0
	}
	return float64(p.Position) / float64(p.Duration) * 100
}

// MarkCompleted checks if the progress warrants marking as completed.
func (p *PlaybackProgress) MarkCompleted() {
	if p.Percentage() >= 90 {
		p.Completed = true
	}
}

// ContinueWatching represents an item in the "Continue Watching" row.
type ContinueWatching struct {
	Content  Content
	Episode  *Episode
	Progress PlaybackProgress
}
