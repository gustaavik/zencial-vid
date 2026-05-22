package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// VideoCastRepository defines persistence operations for cast credits on a video.
type VideoCastRepository interface {
	Create(ctx context.Context, vc *entity.VideoCast) error
	// GetByID returns the credit with the given video_cast.id, or nil if not found.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.VideoCast, error)
	// GetByVideoAndCastAndRole returns the credit for a specific (video, cast, role)
	// triple, or nil if no such credit exists. Used for duplicate detection on create.
	GetByVideoAndCastAndRole(ctx context.Context, videoID, castID uuid.UUID, role string) (*entity.VideoCast, error)
	// ListByVideo returns all credits for a video ordered by sort_order then
	// created_at. Each VideoCast.Cast is populated via a JOIN.
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.VideoCast, error)
	Update(ctx context.Context, vc *entity.VideoCast) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// ListByCast returns a paginated list of published video credits for a cast
	// member ordered by v.created_at DESC. Each VideoCast.Video is populated via JOIN.
	ListByCast(ctx context.Context, castID uuid.UUID, offset, limit int) ([]entity.VideoCast, int, error)
}
