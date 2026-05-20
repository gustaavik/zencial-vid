package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// CastRepository defines persistence operations for video cast credits.
type CastRepository interface {
	Create(ctx context.Context, cast *entity.Cast) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Cast, error)
	ListByVideo(ctx context.Context, videoID uuid.UUID) ([]entity.Cast, error)
	Update(ctx context.Context, cast *entity.Cast) error
	Delete(ctx context.Context, id uuid.UUID) error
}
