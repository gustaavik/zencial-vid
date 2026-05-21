package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// VideoCastRepository defines persistence operations for cast credits on a video.
type VideoCastRepository interface {
	Create(ctx context.Context, vc *entity.VideoCast) error
	// GetByVideoAndCast returns the credit for a specific (video, cast) pair,
	// or nil if no such credit exists.
	GetByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) (*entity.VideoCast, error)
	// ListByVideo returns all credits for a video ordered by sort_order then
	// created_at. Each VideoCast.Cast is populated via a JOIN.
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error)
	Update(ctx context.Context, vc *entity.VideoCast) error
	DeleteByVideoAndCast(ctx context.Context, videoID, castID uuid.UUID) error
}
