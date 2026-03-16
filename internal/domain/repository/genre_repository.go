package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// GenreRepository defines persistence operations for genres.
type GenreRepository interface {
	Create(ctx context.Context, genre *entity.Genre) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error)
	GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Genre, error)
	Update(ctx context.Context, genre *entity.Genre) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, fs *filter.FilterSet) ([]entity.Genre, int64, error)
	ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error)
}
