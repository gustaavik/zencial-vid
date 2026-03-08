package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// ContentRepository defines persistence operations for all content types.
type ContentRepository interface {
	// --- Film ---
	CreateFilm(ctx context.Context, film *entity.Film) error
	GetFilmByID(ctx context.Context, id uuid.UUID) (*entity.Film, error)
	GetFilmBySlug(ctx context.Context, slug string) (*entity.Film, error)
	UpdateFilm(ctx context.Context, film *entity.Film) error

	// --- Video ---
	CreateVideo(ctx context.Context, video *entity.Video) error
	GetVideoByID(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	GetVideoBySlug(ctx context.Context, slug string) (*entity.Video, error)
	UpdateVideo(ctx context.Context, video *entity.Video) error

	// --- Series ---
	CreateSeries(ctx context.Context, series *entity.Series) error
	GetSeriesByID(ctx context.Context, id uuid.UUID) (*entity.Series, error)
	GetSeriesBySlug(ctx context.Context, slug string) (*entity.Series, error)
	UpdateSeries(ctx context.Context, series *entity.Series) error

	// --- Shared ---
	// ExistsBySlug checks whether a slug is already taken across all content types.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	// GetTypeByID returns the ContentType for a given ID without loading the full entity.
	GetTypeByID(ctx context.Context, id uuid.UUID) (entity.ContentType, error)
	// Delete removes a content row (cascades to type-specific tables).
	Delete(ctx context.Context, id uuid.UUID) error
	// SetStatus updates status and updated_at for any content row.
	SetStatus(ctx context.Context, id uuid.UUID, status entity.ContentStatus) error

	// --- List / search (returns lightweight summaries) ---
	Search(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.ContentSummary, int64, error)
	AdminSearch(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.ContentSummary, int64, error)
	GetFeatured(ctx context.Context, limit int) ([]entity.ContentSummary, error)
	GetByGenre(ctx context.Context, genreID uuid.UUID, page, perPage int) ([]entity.ContentSummary, int64, error)

	// --- Seasons and episodes (Series sub-entities) ---
	GetSeasonsForSeries(ctx context.Context, seriesID uuid.UUID) ([]entity.Season, error)
	GetEpisodesForSeason(ctx context.Context, seasonID uuid.UUID) ([]entity.Episode, error)
	GetEpisodeByID(ctx context.Context, id uuid.UUID) (*entity.Episode, error)

	// --- Video assets ---
	CreateVideoAsset(ctx context.Context, asset *entity.VideoAsset, contentID uuid.UUID) error
	GetVideoAssetForContent(ctx context.Context, contentID uuid.UUID) (*entity.VideoAsset, error)
	UpdateVideoAssetStatus(ctx context.Context, assetID uuid.UUID, status entity.VideoAssetStatus) error
}
