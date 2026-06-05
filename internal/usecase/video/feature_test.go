package video

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func newPublishedVideo(t *testing.T) *entity.Video {
	t.Helper()
	slug, err := valueobject.NewSlug("featured-clip")
	require.NoError(t, err)
	v := entity.NewVideo("Featured Clip", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())
	v.Publish()        // processing
	v.MarkTranscoded() // published
	return v
}

func TestService_SetFeatured(t *testing.T) {
	ctx := context.Background()

	t.Run("marks a published video as featured", func(t *testing.T) {
		v := newPublishedVideo(t)
		var featuredID uuid.UUID
		var featuredDesc string

		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Video, error) { return v, nil },
			setFeaturedFn: func(_ context.Context, id uuid.UUID, desc string) error {
				featuredID = id
				featuredDesc = desc
				return nil
			},
		}, nil)

		appErr := svc.SetFeatured(ctx, v.ID, "A great feature description")

		require.Nil(t, appErr)
		assert.Equal(t, v.ID, featuredID)
		assert.Equal(t, "A great feature description", featuredDesc)
	})

	t.Run("returns not-found when video does not exist", func(t *testing.T) {
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		}, nil)

		appErr := svc.SetFeatured(ctx, uuid.New(), "")

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})

	t.Run("returns conflict when video is not published", func(t *testing.T) {
		slug, _ := valueobject.NewSlug("draft-clip")
		draft := entity.NewVideo("Draft", slug, "", "", "G", "videos/x.mp4", "video/mp4", 0, uuid.New())

		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return draft, nil },
		}, nil)

		appErr := svc.SetFeatured(ctx, draft.ID, "")

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotPublished, appErr.Code)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		v := newPublishedVideo(t)
		repoErr := errors.New("db down")

		svc := newTestService(&mockVideoRepo{
			getByIDFn:     func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			setFeaturedFn: func(_ context.Context, _ uuid.UUID, _ string) error { return repoErr },
		}, nil)

		appErr := svc.SetFeatured(ctx, v.ID, "")

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

func TestService_UnsetFeatured(t *testing.T) {
	ctx := context.Background()

	t.Run("removes featured flag from any existing video", func(t *testing.T) {
		v := newPublishedVideo(t)
		v.IsFeatured = true
		var unsetID uuid.UUID

		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Video, error) { return v, nil },
			unsetFeaturedFn: func(_ context.Context, id uuid.UUID) error {
				unsetID = id
				return nil
			},
		}, nil)

		appErr := svc.UnsetFeatured(ctx, v.ID)

		require.Nil(t, appErr)
		assert.Equal(t, v.ID, unsetID)
	})

	t.Run("returns not-found when video does not exist", func(t *testing.T) {
		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return nil, nil },
		}, nil)

		appErr := svc.UnsetFeatured(ctx, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeVideoNotFound, appErr.Code)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		v := newPublishedVideo(t)
		repoErr := errors.New("db down")

		svc := newTestService(&mockVideoRepo{
			getByIDFn:       func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return v, nil },
			unsetFeaturedFn: func(_ context.Context, _ uuid.UUID) error { return repoErr },
		}, nil)

		appErr := svc.UnsetFeatured(ctx, v.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}

// ensure domain sentinel is reachable from this package
var _ = domain.ErrVideoNotPublished
