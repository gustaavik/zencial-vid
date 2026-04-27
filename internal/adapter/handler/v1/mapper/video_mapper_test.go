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

func intPtr(v int) *int { return &v }

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
			userPlanLevel:    intPtr(1),
			wantAccessible:   true,
		},
		{
			name:             "free video (zero minimum) — accessible without plan",
			minimumPlanLevel: intPtr(0),
			userPlanLevel:    nil,
			wantAccessible:   true,
		},
		{
			name:             "gated video — no subscription locks access",
			minimumPlanLevel: intPtr(2),
			userPlanLevel:    nil,
			wantAccessible:   false,
		},
		{
			name:             "gated video — lower plan locks access",
			minimumPlanLevel: intPtr(2),
			userPlanLevel:    intPtr(1),
			wantAccessible:   false,
		},
		{
			name:             "gated video — equal plan grants access",
			minimumPlanLevel: intPtr(2),
			userPlanLevel:    intPtr(2),
			wantAccessible:   true,
		},
		{
			name:             "gated video — higher plan grants access",
			minimumPlanLevel: intPtr(2),
			userPlanLevel:    intPtr(5),
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

func TestVideosToResponseWithAccess_AppliesPlanLevelToEachItem(t *testing.T) {
	ctx := context.Background()

	free := *newPublishedVideo(t, nil)
	gatedLow := *newPublishedVideo(t, intPtr(1))
	gatedHigh := *newPublishedVideo(t, intPtr(5))

	responses := VideosToResponseWithAccess(ctx, []entity.Video{free, gatedLow, gatedHigh}, nil, intPtr(2))

	require.Len(t, responses, 3)
	require.NotNil(t, responses[0].IsAccessible)
	assert.True(t, *responses[0].IsAccessible, "free video accessible")
	require.NotNil(t, responses[1].IsAccessible)
	assert.True(t, *responses[1].IsAccessible, "plan level 2 covers gated level 1")
	require.NotNil(t, responses[2].IsAccessible)
	assert.False(t, *responses[2].IsAccessible, "plan level 2 below gated level 5")
}
