package mapper

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newPublishedVideo(t *testing.T, minimumPlanLevel *int) *entity.Video {
	t.Helper()
	slug, err := valueobject.NewSlug("test-video")
	require.NoError(t, err)
	v := entity.NewVideo("Test", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())
	v.Status = entity.VideoStatusPublished
	v.MinimumPlanLevel = minimumPlanLevel
	return v
}

func TestVideoToResponseWithAccess(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		minimumPlanLevel *int
		userPlanLevel    *int
		wantAccessible   bool
	}{
		{
			name:             "free video (nil minimum) — accessible without plan",
			minimumPlanLevel: nil,
			userPlanLevel:    nil,
			wantAccessible:   true,
		},
		{
			name:             "free video (nil minimum) — accessible with plan",
			minimumPlanLevel: nil,
			userPlanLevel:    new(1),
			wantAccessible:   true,
		},
		{
			name:             "free video (zero minimum) — accessible without plan",
			minimumPlanLevel: new(0),
			userPlanLevel:    nil,
			wantAccessible:   true,
		},
		{
			name:             "gated video — no subscription locks access",
			minimumPlanLevel: new(2),
			userPlanLevel:    nil,
			wantAccessible:   false,
		},
		{
			name:             "gated video — lower plan locks access",
			minimumPlanLevel: new(2),
			userPlanLevel:    new(1),
			wantAccessible:   false,
		},
		{
			name:             "gated video — equal plan grants access",
			minimumPlanLevel: new(2),
			userPlanLevel:    new(2),
			wantAccessible:   true,
		},
		{
			name:             "gated video — higher plan grants access",
			minimumPlanLevel: new(2),
			userPlanLevel:    new(5),
			wantAccessible:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newPublishedVideo(t, tt.minimumPlanLevel)

			resp := VideoToResponseWithAccess(ctx, v, nil, tt.userPlanLevel)

			require.NotNil(t, resp.IsAccessible, "is_accessible must always be populated by WithAccess mapper")
			assert.Equal(t, tt.wantAccessible, *resp.IsAccessible)
		})
	}
}

// stubURLs implements ThumbnailURLBuilder for tests.
type stubURLs struct{ base string }

func (s stubURLs) ThumbnailURL(videoID string) string {
	return s.base + "/api/v1/thumbnails/" + videoID
}

func TestVideoToResponse_EmitsCDNThumbnailURL(t *testing.T) {
	ctx := context.Background()
	v := newPublishedVideo(t, nil)
	v.ThumbnailKey = "videos/" + v.ID.String() + "/thumbnail.jpg"

	resp := VideoToResponse(ctx, v, stubURLs{base: "https://cdn.example"})

	want := "https://cdn.example/api/v1/thumbnails/" + v.ID.String()
	assert.Equal(t, want, resp.ThumbnailURL)
	assert.NotContains(t, resp.ThumbnailURL, "s3", "frontend must never see an S3 host")
	assert.NotContains(t, resp.ThumbnailURL, "minio", "frontend must never see a MinIO host")
}

func TestVideoToResponse_NoThumbnailKey_LeavesURLEmpty(t *testing.T) {
	ctx := context.Background()
	v := newPublishedVideo(t, nil)
	v.ThumbnailKey = ""

	resp := VideoToResponse(ctx, v, stubURLs{base: "https://cdn.example"})

	assert.Empty(t, resp.ThumbnailURL)
}

func TestVideosToResponseWithAccess_AppliesPlanLevelToEachItem(t *testing.T) {
	ctx := context.Background()

	free := *newPublishedVideo(t, nil)
	gatedLow := *newPublishedVideo(t, new(1))
	gatedHigh := *newPublishedVideo(t, new(5))

	responses := VideosToResponseWithAccess(ctx, []entity.Video{free, gatedLow, gatedHigh}, nil, new(2))

	require.Len(t, responses, 3)
	require.NotNil(t, responses[0].IsAccessible)
	assert.True(t, *responses[0].IsAccessible, "free video accessible")
	require.NotNil(t, responses[1].IsAccessible)
	assert.True(t, *responses[1].IsAccessible, "plan level 2 covers gated level 1")
	require.NotNil(t, responses[2].IsAccessible)
	assert.False(t, *responses[2].IsAccessible, "plan level 2 below gated level 5")
}
