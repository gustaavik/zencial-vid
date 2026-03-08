package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// ContentType distinguishes content types.
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

// BaseContent holds fields shared by Film and Video — all playable content types.
// It is embedded in Film and Video; it is not stored as a standalone entity.
type BaseContent struct {
	ID          uuid.UUID
	Type        ContentType
	Title       string
	Slug        valueobject.Slug
	Description string
	Synopsis    string
	Rating      valueobject.ContentRating
	Status      ContentStatus
	IsFeatured  bool
	PosterURL   string

	// Genre is the single assigned genre (nil if unset).
	Genre *Genre

	// Asset is the playable video asset (nil until attached).
	Asset *VideoAsset

	// Duration of the playable content.
	Duration valueobject.Duration

	// Plan is the minimum subscription plan required to watch.
	// nil means the content is free for all subscribers.
	Plan *Plan

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsPlayable reports whether the content can be streamed.
func (b *BaseContent) IsPlayable() bool {
	return b.Status == ContentStatusPublished &&
		b.Asset != nil &&
		b.Asset.Status == VideoAssetReady
}

// IsPublished reports whether the content is published.
func (b *BaseContent) IsPublished() bool {
	return b.Status == ContentStatusPublished
}

// IsFree reports whether the content requires no specific subscription plan.
func (b *BaseContent) IsFree() bool {
	return b.Plan == nil
}

// Publish sets the status to published.
func (b *BaseContent) Publish() {
	b.Status = ContentStatusPublished
	b.UpdatedAt = time.Now()
}

// Archive sets the status to archived.
func (b *BaseContent) Archive() {
	b.Status = ContentStatusArchived
	b.UpdatedAt = time.Now()
}

// Film represents a professionally produced film.
type Film struct {
	BaseContent
	BackdropURL string
	TrailerURL  string
	ReleaseYear int
	Director    string
	CastMembers []CastMember
}

// Video represents viewer-submitted or creator content.
type Video struct {
	BaseContent
	CreatorName string
	UploadedAt  time.Time
}

// ContentSummary is a lightweight projection used for list and search results.
// It covers Film, Video, and Series rows from the content table.
type ContentSummary struct {
	ID          uuid.UUID
	Type        ContentType
	Title       string
	Slug        valueobject.Slug
	Description string
	Status      ContentStatus
	Rating      valueobject.ContentRating
	PosterURL   string
	IsFeatured  bool

	// Genre is the single assigned genre (nil if unset).
	Genre *Genre

	// Plan is nil when the content is free.
	Plan *Plan

	// CreatorName is populated for Video type only.
	CreatorName string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsFree reports whether the content requires no specific subscription plan.
func (c *ContentSummary) IsFree() bool {
	return c.Plan == nil
}

// Series is a container of seasons. It is stored as a content row (type="series")
// so that watchlist and streaming references can use the same content ID.
type Series struct {
	ID           uuid.UUID
	Title        string
	Slug         valueobject.Slug
	Description  string
	Synopsis     string
	PosterURL    string
	BackdropURL  string
	TrailerURL   string
	Status       ContentStatus
	IsFeatured   bool
	TotalSeasons int
	Seasons      []Season
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsPublished reports whether the series is published.
func (s *Series) IsPublished() bool {
	return s.Status == ContentStatusPublished
}

// IsPlayable reports whether the series has seasons available.
func (s *Series) IsPlayable() bool {
	return s.Status == ContentStatusPublished && s.TotalSeasons > 0
}

// Publish sets the status to published.
func (s *Series) Publish() {
	s.Status = ContentStatusPublished
	s.UpdatedAt = time.Now()
}

// Archive sets the status to archived.
func (s *Series) Archive() {
	s.Status = ContentStatusArchived
	s.UpdatedAt = time.Now()
}

// Season represents a series season.
type Season struct {
	ID          uuid.UUID
	SeriesID    uuid.UUID
	Number      int
	Title       string
	TrailerURL  string
	BackdropURL string
	Episodes    []Episode
	CreatedAt   time.Time
}

// Episode represents a single episode within a season.
type Episode struct {
	ID          uuid.UUID
	SeasonID    uuid.UUID
	SeriesID    uuid.UUID
	Number      int
	Title       string
	Synopsis    string
	Duration    valueobject.Duration
	Asset       VideoAsset
	AirDate     *time.Time
	Director    string
	CastMembers []CastMember
	CreatedAt   time.Time
}

// VideoAssetStatus represents the processing status of a video asset.
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
