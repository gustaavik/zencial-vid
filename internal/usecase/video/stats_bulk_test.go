package video

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

func TestService_Stats(t *testing.T) {
	ctx := context.Background()

	want := &repository.VideoStats{
		Total:        3,
		ByStatus:     map[string]int64{"published": 2, "draft": 1},
		BySubmission: map[string]int64{"approved": 2, "draft": 1},
		ByCategory:   []repository.CategoryCount{{GenreID: uuid.New(), Count: 2}},
	}

	svc := newTestService(&mockVideoRepo{
		statsFn: func(_ context.Context) (*repository.VideoStats, error) { return want, nil },
	}, nil)

	got, appErr := svc.Stats(ctx)
	require.Nil(t, appErr)
	assert.Equal(t, want, got)
}

func TestService_ListAdmin(t *testing.T) {
	ctx := context.Background()
	genreID := uuid.New()

	var gotGenre *uuid.UUID
	svc := newTestService(&mockVideoRepo{
		listAdminFn: func(_ context.Context, _ *filter.FilterSet, g *uuid.UUID) ([]entity.Video, int64, error) {
			gotGenre = g
			return []entity.Video{{ID: uuid.New(), Views: 42}}, 1, nil
		},
	}, nil)

	videos, total, appErr := svc.ListAdmin(ctx, &filter.FilterSet{}, &genreID)
	require.Nil(t, appErr)
	assert.Equal(t, int64(1), total)
	require.Len(t, videos, 1)
	assert.Equal(t, int64(42), videos[0].Views)
	require.NotNil(t, gotGenre)
	assert.Equal(t, genreID, *gotGenre)
}

func TestService_BulkUpdate(t *testing.T) {
	ctx := context.Background()
	adminRoles := []entity.UserRole{entity.RoleAdmin}
	pg13 := "PG13"

	t.Run("requires at least one ID", func(t *testing.T) {
		svc := newTestService(&mockVideoRepo{}, nil)
		_, appErr := svc.BulkUpdate(ctx, nil, BulkUpdateInput{ContentRating: &pg13})
		require.NotNil(t, appErr)
	})

	t.Run("requires at least one field to update", func(t *testing.T) {
		svc := newTestService(&mockVideoRepo{}, nil)
		_, appErr := svc.BulkUpdate(ctx, []uuid.UUID{uuid.New()}, BulkUpdateInput{})
		require.NotNil(t, appErr)
	})

	t.Run("reassigns category across multiple videos", func(t *testing.T) {
		id1, id2 := uuid.New(), uuid.New()
		newGenre := uuid.New()
		setGenresCalls := map[uuid.UUID][]uuid.UUID{}

		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Video, error) {
				return &entity.Video{ID: id}, nil
			},
			setGenresFn: func(_ context.Context, videoID uuid.UUID, genreIDs []uuid.UUID) error {
				setGenresCalls[videoID] = genreIDs
				return nil
			},
		}, nil)

		result, appErr := svc.BulkUpdate(ctx, []uuid.UUID{id1, id2}, BulkUpdateInput{
			GenreIDs:    []uuid.UUID{newGenre},
			CallerRoles: adminRoles,
		})
		require.Nil(t, appErr)
		assert.ElementsMatch(t, []uuid.UUID{id1, id2}, result.Succeeded)
		assert.Empty(t, result.Failed)
		assert.Equal(t, []uuid.UUID{newGenre}, setGenresCalls[id1])
		assert.Equal(t, []uuid.UUID{newGenre}, setGenresCalls[id2])
	})

	t.Run("records per-video failures", func(t *testing.T) {
		ok, missing := uuid.New(), uuid.New()

		svc := newTestService(&mockVideoRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Video, error) {
				if id == missing {
					return nil, nil // not found
				}
				return &entity.Video{ID: id}, nil
			},
		}, nil)

		result, appErr := svc.BulkUpdate(ctx, []uuid.UUID{ok, missing}, BulkUpdateInput{
			ContentRating: &pg13,
			CallerRoles:   adminRoles,
		})
		require.Nil(t, appErr)
		assert.Equal(t, []uuid.UUID{ok}, result.Succeeded)
		require.Len(t, result.Failed, 1)
		assert.Equal(t, missing, result.Failed[0].ID)
	})
}
