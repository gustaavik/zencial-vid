package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// MusicCueRepository defines persistence operations for video music cues.
type MusicCueRepository interface {
	Create(ctx context.Context, cue *entity.MusicCue) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.MusicCue, error)
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.MusicCue, error)
	Update(ctx context.Context, cue *entity.MusicCue) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// HasBlockingCues returns true if any cue for this video has pending_clearance status.
	HasBlockingCues(ctx context.Context, videoID uuid.UUID) (bool, error)
}
