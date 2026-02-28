package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// CatalogRepository defines persistence operations for catalog items.
type CatalogRepository interface {
	// Genres
	CreateGenre(ctx context.Context, genre *entity.Genre) error
	GetGenreByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error)
	GetGenreBySlug(ctx context.Context, slug string) (*entity.Genre, error)
	ListGenres(ctx context.Context) ([]entity.Genre, error)
	UpdateGenre(ctx context.Context, genre *entity.Genre) error
	DeleteGenre(ctx context.Context, id uuid.UUID) error

	// Categories
	ListCategories(ctx context.Context) ([]entity.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error)
}
