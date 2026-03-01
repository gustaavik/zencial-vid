package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestContent(t *testing.T, ct ContentType, status ContentStatus) *Content {
	t.Helper()
	slug, err := valueobject.NewSlug("test-content")
	if err != nil {
		t.Fatalf("creating test slug: %v", err)
	}
	return &Content{
		ID:          uuid.New(),
		Type:        ct,
		Title:       "Test Content",
		Slug:        slug,
		Description: "A test",
		Rating:      valueobject.RatingPG,
		ReleaseYear: 2024,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestContent_IsPlayable(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Content
		expect bool
	}{
		{
			name: "published film with ready asset is playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeFilm, ContentStatusPublished)
				c.Film = &Film{
					ContentID: c.ID,
					Duration:  valueobject.NewDuration(7200),
					Asset: VideoAsset{
						ID:     uuid.New(),
						Status: VideoAssetReady,
					},
				}
				return c
			},
			expect: true,
		},
		{
			name: "published film with pending asset is not playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeFilm, ContentStatusPublished)
				c.Film = &Film{
					ContentID: c.ID,
					Asset: VideoAsset{
						ID:     uuid.New(),
						Status: VideoAssetPending,
					},
				}
				return c
			},
			expect: false,
		},
		{
			name: "draft film is not playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeFilm, ContentStatusDraft)
				c.Film = &Film{
					ContentID: c.ID,
					Asset: VideoAsset{
						ID:     uuid.New(),
						Status: VideoAssetReady,
					},
				}
				return c
			},
			expect: false,
		},
		{
			name: "published film with nil Film data is not playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeFilm, ContentStatusPublished)
				return c
			},
			expect: false,
		},
		{
			name: "published series with seasons is playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeSeries, ContentStatusPublished)
				c.Series = &Series{
					ContentID:    c.ID,
					TotalSeasons: 3,
				}
				return c
			},
			expect: true,
		},
		{
			name: "published series with zero seasons is not playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeSeries, ContentStatusPublished)
				c.Series = &Series{
					ContentID:    c.ID,
					TotalSeasons: 0,
				}
				return c
			},
			expect: false,
		},
		{
			name: "archived content is not playable",
			setup: func() *Content {
				c := newTestContent(t, ContentTypeFilm, ContentStatusArchived)
				c.Film = &Film{
					ContentID: c.ID,
					Asset: VideoAsset{
						ID:     uuid.New(),
						Status: VideoAssetReady,
					},
				}
				return c
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			assert.Equal(t, tt.expect, c.IsPlayable())
		})
	}
}

func TestContent_IsPublished(t *testing.T) {
	tests := []struct {
		name   string
		status ContentStatus
		want   bool
	}{
		{"published", ContentStatusPublished, true},
		{"draft", ContentStatusDraft, false},
		{"archived", ContentStatusArchived, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContent(t, ContentTypeFilm, tt.status)
			assert.Equal(t, tt.want, c.IsPublished())
		})
	}
}

func TestContent_Publish(t *testing.T) {
	c := newTestContent(t, ContentTypeFilm, ContentStatusDraft)
	before := c.UpdatedAt

	time.Sleep(time.Millisecond)
	c.Publish()

	assert.Equal(t, ContentStatusPublished, c.Status)
	assert.True(t, c.UpdatedAt.After(before), "UpdatedAt should be updated")
}

func TestContent_Archive(t *testing.T) {
	c := newTestContent(t, ContentTypeFilm, ContentStatusPublished)
	before := c.UpdatedAt

	time.Sleep(time.Millisecond)
	c.Archive()

	assert.Equal(t, ContentStatusArchived, c.Status)
	assert.True(t, c.UpdatedAt.After(before), "UpdatedAt should be updated")
}
