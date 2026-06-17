package series

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	uploaderID := uuid.New()

	t.Run("creates series in draft status with slug suffix", func(t *testing.T) {
		svc := newTestService(nil, nil, nil, nil)

		out, appErr := svc.Create(ctx, &CreateInput{
			Title:      "My Great Series",
			UploadedBy: uploaderID,
		})

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.Equal(t, "My Great Series", out.Series.Title)
		assert.Contains(t, out.Series.Slug.String(), "my-great-series")
		assert.NotEqual(t, "my-great-series", out.Series.Slug.String(), "slug should have random suffix")
		assert.Equal(t, entity.SeriesStatusDraft, out.Series.Status)
	})

	t.Run("dispatches SeriesCreated event on success", func(t *testing.T) {
		dispatcher := &mockDispatcher{}
		svc := newTestService(nil, nil, nil, dispatcher)

		out, appErr := svc.Create(ctx, &CreateInput{Title: "Event Series", UploadedBy: uploaderID})

		require.Nil(t, appErr)
		require.Len(t, dispatcher.dispatched, 1)
		ev, ok := dispatcher.dispatched[0].(event.SeriesCreated)
		require.True(t, ok)
		assert.Equal(t, out.Series.ID, ev.SeriesID)
		assert.Equal(t, "Event Series", ev.Title)
	})

	t.Run("returns validation error for empty title", func(t *testing.T) {
		svc := newTestService(nil, nil, nil, nil)

		_, appErr := svc.Create(ctx, &CreateInput{Title: "", UploadedBy: uploaderID})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("returns 500 when repo create fails", func(t *testing.T) {
		repo := &mockSeriesRepo{
			createFn: func(_ context.Context, _ *entity.Series) error {
				return errors.New("db down")
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.Create(ctx, &CreateInput{Title: "Title", UploadedBy: uploaderID})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("applies extended metadata fields to the entity", func(t *testing.T) {
		var created *entity.Series
		repo := &mockSeriesRepo{
			createFn: func(_ context.Context, s *entity.Series) error {
				created = s
				return nil
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		hide := true
		autoplay := false
		out, appErr := svc.Create(ctx, &CreateInput{
			Title:               "Bedford Street",
			UploadedBy:          uploaderID,
			SeriesType:          "limited",
			Logline:             "A neighborhood comes alive after dark.",
			PrimaryLanguage:     "en",
			OriginCountry:       "United States",
			ContentRating:       "TV-14",
			PosterKey:           "series/x/poster.jpg",
			AutoplayNext:        &autoplay,
			HideEpisodeCount:    &hide,
			DefaultVisibility:   "followers_only",
			DefaultMonetization: []string{"premium"},
		})

		require.Nil(t, appErr)
		require.NotNil(t, created)
		assert.Equal(t, entity.SeriesTypeLimited, created.SeriesType)
		assert.Equal(t, "A neighborhood comes alive after dark.", created.Logline)
		assert.Equal(t, "United States", created.OriginCountry)
		assert.Equal(t, "TV-14", created.ContentRating)
		assert.Equal(t, "series/x/poster.jpg", created.PosterKey)
		assert.False(t, created.AutoplayNext)
		assert.True(t, created.HideEpisodeCount)
		assert.Equal(t, entity.VideoVisibilityFollowers, created.DefaultVisibility)
		assert.Equal(t, []string{"premium"}, created.DefaultMonetization)
		assert.Equal(t, created.ID, out.Series.ID)
	})

	t.Run("keeps entity defaults when extended fields are omitted", func(t *testing.T) {
		var created *entity.Series
		repo := &mockSeriesRepo{
			createFn: func(_ context.Context, s *entity.Series) error {
				created = s
				return nil
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.Create(ctx, &CreateInput{Title: "Defaults", UploadedBy: uploaderID})

		require.Nil(t, appErr)
		require.NotNil(t, created)
		assert.Equal(t, entity.SeriesTypeOngoing, created.SeriesType)
		assert.Equal(t, "en", created.PrimaryLanguage)
		assert.True(t, created.AutoplayNext)
		assert.True(t, created.BingeMode)
		assert.False(t, created.HideEpisodeCount)
		assert.Equal(t, entity.VideoVisibilityPublic, created.DefaultVisibility)
	})

	t.Run("sets GenreIDs when provided", func(t *testing.T) {
		genreID := uuid.New()
		repo := &mockSeriesRepo{
			getGenreIDsFn: func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{genreID}, nil
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		out, appErr := svc.Create(ctx, &CreateInput{
			Title:      "Genre Series",
			UploadedBy: uploaderID,
			GenreIDs:   []uuid.UUID{genreID},
		})

		require.Nil(t, appErr)
		assert.Equal(t, []uuid.UUID{genreID}, out.Series.GenreIDs)
	})
}
