package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// ContentRepository defines persistence operations for content.
type ContentRepository interface {
	Create(ctx context.Context, content *entity.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Content, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Content, error)
	Update(ctx context.Context, content *entity.Content) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, criteria entity.SearchCriteria) ([]entity.Content, int64, error)
	GetFeatured(ctx context.Context, limit int) ([]entity.Content, error)
	GetByGenre(ctx context.Context, genreID uuid.UUID, page, perPage int) ([]entity.Content, int64, error)

	// Series-specific
	GetSeasonsForContent(ctx context.Context, contentID uuid.UUID) ([]entity.Season, error)
	GetEpisodesForSeason(ctx context.Context, seasonID uuid.UUID) ([]entity.Episode, error)
	GetEpisodeByID(ctx context.Context, id uuid.UUID) (*entity.Episode, error)
}
