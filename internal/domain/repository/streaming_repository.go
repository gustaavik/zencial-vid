package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// StreamingRepository defines persistence operations for streaming.
type StreamingRepository interface {
	// Sessions
	CreateSession(ctx context.Context, session *entity.StreamSession) error
	GetActiveSessionsByUser(ctx context.Context, userID uuid.UUID) ([]entity.StreamSession, error)
	EndSession(ctx context.Context, sessionID uuid.UUID) error
	EndSessionsForContent(ctx context.Context, userID, contentID uuid.UUID) error

	// Progress
	SaveProgress(ctx context.Context, progress *entity.PlaybackProgress) error
	GetProgress(ctx context.Context, userID, contentID uuid.UUID, episodeID *uuid.UUID) (*entity.PlaybackProgress, error)
	GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]entity.ContinueWatching, error)
}
