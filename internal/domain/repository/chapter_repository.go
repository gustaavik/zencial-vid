package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// ChapterRepository defines persistence operations for video chapters.
type ChapterRepository interface {
	// BulkReplace atomically replaces all chapters for a video.
	BulkReplace(ctx context.Context, videoID uuid.UUID, chapters []entity.Chapter) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Chapter, error)
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Chapter, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByVideo(ctx context.Context, videoID uuid.UUID) error
}
