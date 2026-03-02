package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// ContentRepository defines persistence operations for content.
type ContentRepository interface {
	Create(ctx context.Context, content *entity.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Content, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Content, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Update(ctx context.Context, content *entity.Content) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, error)
	AdminSearch(ctx context.Context, fs filter.FilterSet, searchQuery string) ([]entity.Content, int64, error)
	GetFeatured(ctx context.Context, limit int) ([]entity.Content, error)
	GetByGenre(ctx context.Context, genreID uuid.UUID, page, perPage int) ([]entity.Content, int64, error)

	// Series-specific
	GetSeasonsForContent(ctx context.Context, contentID uuid.UUID) ([]entity.Season, error)
	GetEpisodesForSeason(ctx context.Context, seasonID uuid.UUID) ([]entity.Episode, error)
	GetEpisodeByID(ctx context.Context, id uuid.UUID) (*entity.Episode, error)

	// Video-specific
	CreateVideo(ctx context.Context, video *entity.Video) error
	UpdateVideo(ctx context.Context, video *entity.Video) error
	GetVideoForContent(ctx context.Context, contentID uuid.UUID) (*entity.Video, error)

	// Video assets
	CreateVideoAsset(ctx context.Context, asset *entity.VideoAsset, contentID uuid.UUID) error
	GetVideoAssetForContent(ctx context.Context, contentID uuid.UUID) (*entity.VideoAsset, error)
	UpdateVideoAssetStatus(ctx context.Context, assetID uuid.UUID, status entity.VideoAssetStatus) error
}
