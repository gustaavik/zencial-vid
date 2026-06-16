package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// VideoStorageInfo holds the storage keys for a single video row, used for
// cross-referencing the database against object storage.
type VideoStorageInfo struct {
	ID           uuid.UUID
	StorageKey   string
	ThumbnailKey string
}

// CategoryCount is the number of videos associated with a single genre.
type CategoryCount struct {
	GenreID uuid.UUID
	Count   int64
}

// VideoStats holds platform-wide catalog aggregates for the admin dashboard.
type VideoStats struct {
	Total        int64
	ByStatus     map[string]int64
	BySubmission map[string]int64
	ByCategory   []CategoryCount
}

// VideoRepository defines persistence operations for videos.
type VideoRepository interface {
	Create(ctx context.Context, video *entity.Video) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error)
	GetBySlug(ctx context.Context, slug valueobject.Slug) (*entity.Video, error)
	Update(ctx context.Context, video *entity.Video) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	// ListAdmin lists videos in any status with per-row view counts populated,
	// optionally filtered to videos tagged with genreID (nil = no genre filter).
	ListAdmin(ctx context.Context, fs *filter.FilterSet, genreID *uuid.UUID) ([]entity.Video, int64, error)
	// Stats returns platform-wide catalog aggregates for the admin dashboard.
	Stats(ctx context.Context) (*VideoStats, error)
	ListPublished(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	ListByUploader(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error)
	ExistsBySlug(ctx context.Context, slug valueobject.Slug) (bool, error)
	SetGenres(ctx context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error
	GetGenreIDs(ctx context.Context, videoID uuid.UUID) ([]uuid.UUID, error)
	ListAllStorageKeys(ctx context.Context) ([]VideoStorageInfo, error)
	SetSeriesEpisode(ctx context.Context, videoID, seriesID uuid.UUID, season, episode int) error
	RemoveFromSeries(ctx context.Context, videoID uuid.UUID) error
	ListBySeries(ctx context.Context, seriesID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, error)
	ListFeatured(ctx context.Context, fs *filter.FilterSet) ([]entity.Video, int64, error)
	SetFeatured(ctx context.Context, videoID uuid.UUID, description string) error
	UnsetFeatured(ctx context.Context, videoID uuid.UUID) error
}
