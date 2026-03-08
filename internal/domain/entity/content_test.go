package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestFilm(t *testing.T, status ContentStatus) *Film {
	t.Helper()
	slug, err := valueobject.NewSlug("test-film")
	if err != nil {
		t.Fatalf("creating test slug: %v", err)
	}
	return &Film{
		BaseContent: BaseContent{
			ID:        uuid.New(),
			Type:      ContentTypeFilm,
			Title:     "Test Film",
			Slug:      slug,
			Rating:    valueobject.RatingPG,
			Status:    status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func newTestVideo(t *testing.T, status ContentStatus) *Video {
	t.Helper()
	slug, err := valueobject.NewSlug("test-video")
	if err != nil {
		t.Fatalf("creating test slug: %v", err)
	}
	return &Video{
		BaseContent: BaseContent{
			ID:        uuid.New(),
			Type:      ContentTypeVideo,
			Title:     "Test Video",
			Slug:      slug,
			Rating:    valueobject.RatingG,
			Status:    status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CreatorName: "Test Creator",
	}
}

func TestFilm_IsPlayable(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Film
		expect bool
	}{
		{
			name: "published film with ready asset is playable",
			setup: func() *Film {
				f := newTestFilm(t, ContentStatusPublished)
				f.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetReady}
				return f
			},
			expect: true,
		},
		{
			name: "published film with pending asset is not playable",
			setup: func() *Film {
				f := newTestFilm(t, ContentStatusPublished)
				f.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetPending}
				return f
			},
			expect: false,
		},
		{
			name: "draft film is not playable",
			setup: func() *Film {
				f := newTestFilm(t, ContentStatusDraft)
				f.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetReady}
				return f
			},
			expect: false,
		},
		{
			name: "published film with nil asset is not playable",
			setup: func() *Film {
				return newTestFilm(t, ContentStatusPublished)
			},
			expect: false,
		},
		{
			name: "archived film is not playable",
			setup: func() *Film {
				f := newTestFilm(t, ContentStatusArchived)
				f.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetReady}
				return f
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.setup().IsPlayable())
		})
	}
}

func TestVideo_IsPlayable(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Video
		expect bool
	}{
		{
			name: "published video with ready asset is playable",
			setup: func() *Video {
				v := newTestVideo(t, ContentStatusPublished)
				v.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetReady}
				return v
			},
			expect: true,
		},
		{
			name: "published video with pending asset is not playable",
			setup: func() *Video {
				v := newTestVideo(t, ContentStatusPublished)
				v.Asset = &VideoAsset{ID: uuid.New(), Status: VideoAssetPending}
				return v
			},
			expect: false,
		},
		{
			name: "published video with nil asset is not playable",
			setup: func() *Video {
				return newTestVideo(t, ContentStatusPublished)
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.setup().IsPlayable())
		})
	}
}

func TestBaseContent_IsPublished(t *testing.T) {
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
			assert.Equal(t, tt.want, newTestFilm(t, tt.status).IsPublished())
		})
	}
}

func TestBaseContent_IsFree(t *testing.T) {
	f := newTestFilm(t, ContentStatusPublished)
	assert.True(t, f.IsFree(), "film with no plan should be free")

	f.Plan = &Plan{ID: uuid.New(), Name: "Premium", Tier: PlanPremium}
	assert.False(t, f.IsFree(), "film with plan should not be free")
}

func TestBaseContent_Publish(t *testing.T) {
	f := newTestFilm(t, ContentStatusDraft)
	before := f.UpdatedAt
	time.Sleep(time.Millisecond)
	f.Publish()
	assert.Equal(t, ContentStatusPublished, f.Status)
	assert.True(t, f.UpdatedAt.After(before))
}

func TestBaseContent_Archive(t *testing.T) {
	f := newTestFilm(t, ContentStatusPublished)
	before := f.UpdatedAt
	time.Sleep(time.Millisecond)
	f.Archive()
	assert.Equal(t, ContentStatusArchived, f.Status)
	assert.True(t, f.UpdatedAt.After(before))
}

func TestSeries_IsPlayable(t *testing.T) {
	slug, _ := valueobject.NewSlug("test-series")
	s := &Series{
		ID:           uuid.New(),
		Slug:         slug,
		Status:       ContentStatusPublished,
		TotalSeasons: 3,
	}
	assert.True(t, s.IsPlayable())

	s.TotalSeasons = 0
	assert.False(t, s.IsPlayable(), "series with no seasons should not be playable")

	s.TotalSeasons = 2
	s.Status = ContentStatusDraft
	assert.False(t, s.IsPlayable(), "unpublished series should not be playable")
}
