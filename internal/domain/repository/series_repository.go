package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// SeriesRepository defines persistence operations for series.
type SeriesRepository interface {
	Create(ctx context.Context, series *entity.Series) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Series, error)
	GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Series, error)
	Update(ctx context.Context, series *entity.Series) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error)
	ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Series, int64, error)
	ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Series, int64, error)
	ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error)
	SetGenres(ctx context.Context, seriesID uuid.UUID, genreIDs []uuid.UUID) error
	GetGenreIDs(ctx context.Context, seriesID uuid.UUID) ([]uuid.UUID, error)
}
