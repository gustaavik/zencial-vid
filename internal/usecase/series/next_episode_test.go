package series

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

func TestService_GetNextEpisode(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	makeEpisodes := func(seriesID uuid.UUID, count int) []entity.Video {
		episodes := make([]entity.Video, count)
		for i := range episodes {
			slug, _ := valueobject.NewSlug("ep")
			v := entity.NewVideo("Episode", slug.WithRandomID(), "", "", "G", "k", "video/mp4", 0, uuid.New())
			season := 1
			epNum := i + 1
			v.SeriesID = &seriesID
			v.SeasonNumber = &season
			v.EpisodeNumber = &epNum
			episodes[i] = *v
		}
		return episodes
	}

	t.Run("returns first episode when no watch progress", func(t *testing.T) {
		series := newPublishedSeries()
		episodes := makeEpisodes(series.ID, 3)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			listBySeriesFn: func(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
				return episodes, int64(len(episodes)), nil
			},
		}
		svc := newTestService(repo, nil, vr, nil)

		result, appErr := svc.GetNextEpisode(ctx, userID, series.ID)

		require.Nil(t, appErr)
		assert.Equal(t, episodes[0].ID, result.ID)
	})

	t.Run("returns next episode after last watched", func(t *testing.T) {
		series := newPublishedSeries()
		episodes := makeEpisodes(series.ID, 3)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			listBySeriesFn: func(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
				return episodes, int64(len(episodes)), nil
			},
		}
		swpr := &mockSeriesWpRepo{
			getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.SeriesWatchProgress, error) {
				return &entity.SeriesWatchProgress{LastEpisodeID: episodes[0].ID}, nil
			},
		}
		svc := newTestService(repo, swpr, vr, nil)

		result, appErr := svc.GetNextEpisode(ctx, userID, series.ID)

		require.Nil(t, appErr)
		assert.Equal(t, episodes[1].ID, result.ID)
	})

	t.Run("returns not found when all episodes watched", func(t *testing.T) {
		series := newPublishedSeries()
		episodes := makeEpisodes(series.ID, 2)

		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			listBySeriesFn: func(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
				return episodes, int64(len(episodes)), nil
			},
		}
		swpr := &mockSeriesWpRepo{
			getFn: func(_ context.Context, _, _ uuid.UUID) (*entity.SeriesWatchProgress, error) {
				return &entity.SeriesWatchProgress{LastEpisodeID: episodes[1].ID}, nil
			},
		}
		svc := newTestService(repo, swpr, vr, nil)

		_, appErr := svc.GetNextEpisode(ctx, userID, series.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeNotFound, appErr.Code)
	})

	t.Run("returns not found when series has no episodes", func(t *testing.T) {
		series := newPublishedSeries()
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		vr := &mockVideoRepo{
			listBySeriesFn: func(_ context.Context, _ uuid.UUID, _ *filter.FilterSet) ([]entity.Video, int64, error) {
				return nil, 0, nil
			},
		}
		svc := newTestService(repo, nil, vr, nil)

		_, appErr := svc.GetNextEpisode(ctx, userID, series.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeNotFound, appErr.Code)
	})

	t.Run("returns 404 when series not found", func(t *testing.T) {
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return nil, nil },
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.GetNextEpisode(ctx, userID, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesNotFound, appErr.Code)
	})
}
