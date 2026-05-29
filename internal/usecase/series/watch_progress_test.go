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

func TestService_UpdateWatchProgress(t *testing.T) {
	ctx := context.Background()

	t.Run("upserts progress for valid episode in series", func(t *testing.T) {
		series := newPublishedSeries()
		video := newPublishedVideo(series.ID, 1, 1)

		upserted := false
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
		}
		swpr := &mockSeriesWpRepo{
			upsertFn: func(_ context.Context, _, _, _ uuid.UUID) error {
				upserted = true
				return nil
			},
		}
		svc := newTestService(repo, swpr, vr, nil)

		appErr := svc.UpdateWatchProgress(ctx, uuid.New(), series.ID, video.ID)

		require.Nil(t, appErr)
		assert.True(t, upserted)
	})

	t.Run("returns 400 when video does not belong to series", func(t *testing.T) {
		series := newPublishedSeries()
		otherSeriesID := uuid.New()
		video := newPublishedVideo(otherSeriesID, 1, 1)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Video, error) { return video, nil },
		}
		svc := newTestService(repo, nil, vr, nil)

		appErr := svc.UpdateWatchProgress(ctx, uuid.New(), series.ID, video.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
	})

	t.Run("returns 404 when series not found", func(t *testing.T) {
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return nil, nil },
		}
		svc := newTestService(repo, nil, nil, nil)

		appErr := svc.UpdateWatchProgress(ctx, uuid.New(), uuid.New(), uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesNotFound, appErr.Code)
	})
}

func TestService_GetWatchProgress(t *testing.T) {
	ctx := context.Background()

	t.Run("returns existing watch progress", func(t *testing.T) {
		series := newPublishedSeries()
		episodeID := uuid.New()

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		swpr := &mockSeriesWpRepo{
			getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.SeriesWatchProgress, error) {
				return &entity.SeriesWatchProgress{
					UserID:        uuid.New(),
					SeriesID:      series.ID,
					LastEpisodeID: episodeID,
				}, nil
			},
		}
		svc := newTestService(repo, swpr, nil, nil)

		progress, appErr := svc.GetWatchProgress(ctx, uuid.New(), series.ID)

		require.Nil(t, appErr)
		assert.Equal(t, episodeID, progress.LastEpisodeID)
	})

	t.Run("returns 404 when no progress exists", func(t *testing.T) {
		series := newPublishedSeries()
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		swpr := &mockSeriesWpRepo{
			getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.SeriesWatchProgress, error) {
				return nil, nil
			},
		}
		svc := newTestService(repo, swpr, nil, nil)

		_, appErr := svc.GetWatchProgress(ctx, uuid.New(), series.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesWatchProgressNotFound, appErr.Code)
	})
}
