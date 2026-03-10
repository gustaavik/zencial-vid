package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// VideoRepository defines persistence operations for videos.
type VideoRepository interface {
	Create(ctx context.Context, video *entity.Video) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Video, error)
	Update(ctx context.Context, video *entity.Video) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, fs filter.FilterSet) ([]entity.Video, int64, error)
	ListPublished(ctx context.Context, fs filter.FilterSet) ([]entity.Video, int64, error)
	ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error)
	SetGenres(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error
	GetGenreIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error)
}
