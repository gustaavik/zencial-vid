package series

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_AddEpisode(t *testing.T) {
	ctx := context.Background()

	t.Run("links video to series as episode", func(t *testing.T) {
		series := newDraftSeries()
		video := newPublishedVideo(uuid.Nil, 0, 0)
		video.SeriesID = nil
		video.SeasonNumber = nil
		video.EpisodeNumber = nil

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
		}
		svc := newTestService(repo, nil, vr, nil)

		result, appErr := svc.AddEpisode(ctx, &AddEpisodeInput{
			SeriesID:      series.ID,
			VideoID:       video.ID,
			SeasonNumber:  1,
			EpisodeNumber: 1,
			CallerID:      series.UploadedBy,
			CallerRoles:   []entity.UserRole{entity.RolePublisher},
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, &series.ID, result.SeriesID)
		assert.Equal(t, 1, *result.SeasonNumber)
		assert.Equal(t, 1, *result.EpisodeNumber)
	})

	t.Run("returns 403 when caller does not own series", func(t *testing.T) {
		series := newDraftSeries()
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.AddEpisode(ctx, &AddEpisodeInput{
			SeriesID:    series.ID,
			VideoID:     uuid.New(),
			CallerID:    uuid.New(), // different owner
			CallerRoles: []entity.UserRole{entity.RolePublisher},
		})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesOwnershipRequired, appErr.Code)
	})

	t.Run("returns 409 when video already in another series", func(t *testing.T) {
		series := newDraftSeries()
		otherSeriesID := uuid.New()
		video := newPublishedVideo(otherSeriesID, 1, 1)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
		}
		svc := newTestService(repo, nil, vr, nil)

		_, appErr := svc.AddEpisode(ctx, &AddEpisodeInput{
			SeriesID:    series.ID,
			VideoID:     video.ID,
			CallerID:    series.UploadedBy,
			CallerRoles: []entity.UserRole{entity.RolePublisher},
		})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeEpisodeAlreadyExists, appErr.Code)
	})

	t.Run("returns 404 when series not found", func(t *testing.T) {
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return nil, nil },
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.AddEpisode(ctx, &AddEpisodeInput{SeriesID: uuid.New()})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesNotFound, appErr.Code)
	})
}

func TestService_RemoveEpisode(t *testing.T) {
	ctx := context.Background()

	t.Run("removes video from series", func(t *testing.T) {
		series := newDraftSeries()
		video := newPublishedVideo(series.ID, 1, 1)

		removed := false
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
			removeFromSeriesFn: func(_ context.Context, _ uuid.UUID) error {
				removed = true
				return nil
			},
		}
		svc := newTestService(repo, nil, vr, nil)

		appErr := svc.RemoveEpisode(ctx, &RemoveEpisodeInput{
			SeriesID:    series.ID,
			VideoID:     video.ID,
			CallerID:    series.UploadedBy,
			CallerRoles: []entity.UserRole{entity.RolePublisher},
		})

		require.Nil(t, appErr)
		assert.True(t, removed)
	})

	t.Run("returns 400 when video does not belong to series", func(t *testing.T) {
		series := newDraftSeries()
		otherSeriesID := uuid.New()
		video := newPublishedVideo(otherSeriesID, 1, 1)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
		}
		svc := newTestService(repo, nil, vr, nil)

		appErr := svc.RemoveEpisode(ctx, &RemoveEpisodeInput{
			SeriesID:    series.ID,
			VideoID:     video.ID,
			CallerID:    series.UploadedBy,
			CallerRoles: []entity.UserRole{entity.RolePublisher},
		})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
	})
}
