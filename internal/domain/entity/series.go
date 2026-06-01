package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// SeriesStatus represents the lifecycle state of a series.
type SeriesStatus string

const (
	SeriesStatusDraft     SeriesStatus = "draft"
	SeriesStatusPublished SeriesStatus = "published"
	SeriesStatusArchived  SeriesStatus = "archived"
)

// SeriesType describes the narrative structure of a series.
type SeriesType string

const (
	SeriesTypeOngoing     SeriesType = "ongoing"
	SeriesTypeLimited     SeriesType = "limited"
	SeriesTypeAnthology   SeriesType = "anthology"
	SeriesTypeDocumentary SeriesType = "documentary"
)

// VideoVisibility controls viewer-facing access to a video or episode.
type VideoVisibility string

const (
	VideoVisibilityPublic    VideoVisibility = "public"
	VideoVisibilityUnlisted  VideoVisibility = "unlisted"
	VideoVisibilityFollowers VideoVisibility = "followers_only"
	VideoVisibilityPrivate   VideoVisibility = "private"
)

// Series is the core series entity grouping ordered episodes.
type Series struct {
	ID                  uuid.UUID
	Title               string
	Slug                valueobject.Slug
	Description         string
	Creator             string
	Status              SeriesStatus
	SeriesType          SeriesType
	Logline             string
	PrimaryLanguage     string
	OriginCountry       string
	CoverImageKey       string
	PosterKey           string
	BannerKey           string
	TitleLogoKey        string
	UploadedBy          uuid.UUID
	GenreIDs            []uuid.UUID
	MinimumPlanLevel    *int
	AutoplayNext        bool
	BingeMode           bool
	HideEpisodeCount    bool
	DefaultVisibility   VideoVisibility
	DefaultMonetization []string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewSeries creates a new Series entity in draft status.
func NewSeries(title string, slug valueobject.Slug, description, creator string, uploadedBy uuid.UUID) *Series {
	now := time.Now().UTC()
	return &Series{
		ID:                  uuid.Must(uuid.NewV7()),
		Title:               title,
		Slug:                slug,
		Description:         description,
		Creator:             creator,
		Status:              SeriesStatusDraft,
		SeriesType:          SeriesTypeOngoing,
		PrimaryLanguage:     "en",
		UploadedBy:          uploadedBy,
		AutoplayNext:        true,
		BingeMode:           true,
		HideEpisodeCount:    false,
		DefaultVisibility:   VideoVisibilityPublic,
		DefaultMonetization: []string{},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// Publish transitions the series from draft to published.
func (s *Series) Publish() {
	s.Status = SeriesStatusPublished
	s.UpdatedAt = time.Now().UTC()
}

// Archive soft-deletes the series.
func (s *Series) Archive() {
	s.Status = SeriesStatusArchived
	s.UpdatedAt = time.Now().UTC()
}

// Unarchive restores an archived series back to draft status.
func (s *Series) Unarchive() {
	s.Status = SeriesStatusDraft
	s.UpdatedAt = time.Now().UTC()
}

// IsPublishable reports whether the series can be published.
func (s *Series) IsPublishable() bool {
	return s.Status == SeriesStatusDraft
}

// SetGenres replaces the genre associations.
func (s *Series) SetGenres(genreIDs []uuid.UUID) {
	s.GenreIDs = genreIDs
	s.UpdatedAt = time.Now().UTC()
}
