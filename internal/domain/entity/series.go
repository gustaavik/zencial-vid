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

// Series is the core series entity grouping ordered episodes.
type Series struct {
	ID               uuid.UUID
	Title            string
	Slug             valueobject.Slug
	Description      string
	Creator          string
	Status           SeriesStatus
	CoverImageKey    string
	UploadedBy       uuid.UUID
	GenreIDs         []uuid.UUID
	MinimumPlanLevel *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewSeries creates a new Series entity in draft status.
func NewSeries(title string, slug valueobject.Slug, description, creator string, uploadedBy uuid.UUID) *Series {
	now := time.Now().UTC()
	return &Series{
		ID:          uuid.New(),
		Title:       title,
		Slug:        slug,
		Description: description,
		Creator:     creator,
		Status:      SeriesStatusDraft,
		UploadedBy:  uploadedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
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
