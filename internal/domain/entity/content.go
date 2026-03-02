package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// ContentType distinguishes films from series.
type ContentType string

const (
	ContentTypeFilm   ContentType = "film"
	ContentTypeSeries ContentType = "series"
	ContentTypeVideo  ContentType = "video"
)

// ContentStatus represents the publication status.
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
)

// Content is the aggregate root for all video content.
type Content struct {
	ID          uuid.UUID
	Type        ContentType
	Title       string
	Slug        valueobject.Slug
	Description string
	Synopsis    string
	Rating      valueobject.ContentRating
	ReleaseYear int
	PosterURL   string
	BackdropURL string
	TrailerURL  string
	Status      ContentStatus
	IsFeatured  bool
	Director    string

	// Relationships (loaded selectively)
	Genres []Genre
	Tags   []Tag
	Cast   []CastMember

	// Type-specific data (nil when not applicable)
	Film   *Film
	Series *Series
	Video  *Video

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Film holds film-specific data.
type Film struct {
	ContentID uuid.UUID
	Duration  valueobject.Duration
	Asset     VideoAsset
}

// Video holds video-specific data for viewer-submitted content.
type Video struct {
	ContentID   uuid.UUID
	Duration    valueobject.Duration
	CreatorName string
	IsFree      bool
	Asset       VideoAsset
}

// Series holds series-specific data.
type Series struct {
	ContentID    uuid.UUID
	Seasons      []Season
	TotalSeasons int
}

// Season represents a series season.
type Season struct {
	ID        uuid.UUID
	ContentID uuid.UUID
	Number    int
	Title     string
	Episodes  []Episode
	CreatedAt time.Time
}

// Episode represents a single episode.
type Episode struct {
	ID        uuid.UUID
	SeasonID  uuid.UUID
	Number    int
	Title     string
	Synopsis  string
	Duration  valueobject.Duration
	Asset     VideoAsset
	AirDate   *time.Time
	CreatedAt time.Time
}

// VideoAssetStatus represents the processing status of a video.
type VideoAssetStatus string

const (
	VideoAssetPending    VideoAssetStatus = "pending"
	VideoAssetProcessing VideoAssetStatus = "processing"
	VideoAssetReady      VideoAssetStatus = "ready"
	VideoAssetFailed     VideoAssetStatus = "failed"
)

// VideoAsset represents a video with multiple quality renditions.
type VideoAsset struct {
	ID         uuid.UUID
	StorageKey string
	Qualities  []VideoRendition
	Status     VideoAssetStatus
}

// VideoRendition represents a single quality version of a video.
type VideoRendition struct {
	Quality    valueobject.VideoQuality
	URL        string // HLS manifest URL
	Bitrate    int    // kbps
	Resolution string // e.g., "1920x1080"
}

// CastMember represents a person involved in content production.
type CastMember struct {
	Name      string
	Role      string // "actor", "director", "writer"
	Character string // Character name (for actors)
	ImageURL  string
}

// IsPlayable checks if this content can be streamed.
func (c *Content) IsPlayable() bool {
	if c.Status != ContentStatusPublished {
		return false
	}
	switch c.Type {
	case ContentTypeFilm:
		return c.Film != nil && c.Film.Asset.Status == VideoAssetReady
	case ContentTypeSeries:
		return c.Series != nil && c.Series.TotalSeasons > 0
	case ContentTypeVideo:
		return c.Video != nil && c.Video.Asset.Status == VideoAssetReady
	}
	return false
}

// IsPublished reports whether the content is published.
func (c *Content) IsPublished() bool {
	return c.Status == ContentStatusPublished
}

// Publish sets the content status to published.
func (c *Content) Publish() {
	c.Status = ContentStatusPublished
	c.UpdatedAt = time.Now()
}

// Archive sets the content status to archived.
func (c *Content) Archive() {
	c.Status = ContentStatusArchived
	c.UpdatedAt = time.Now()
}
