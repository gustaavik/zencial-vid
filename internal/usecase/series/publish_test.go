package series

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_Publish(t *testing.T) {
	ctx := context.Background()

	t.Run("publishes draft series and dispatches event", func(t *testing.T) {
		series := newDraftSeries()
		dispatcher := &mockDispatcher{}
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		svc := newTestService(repo, nil, nil, dispatcher)

		result, appErr := svc.Publish(ctx, series.ID)

		require.Nil(t, appErr)
		assert.Equal(t, entity.SeriesStatusPublished, result.Status)
		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.SeriesPublished)
		assert.True(t, ok)
	})

	t.Run("is idempotent when already published", func(t *testing.T) {
		series := newPublishedSeries()
		dispatcher := &mockDispatcher{}
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return series, nil },
		}
		svc := newTestService(repo, nil, nil, dispatcher)

		_, appErr := svc.Publish(ctx, series.ID)

		require.Nil(t, appErr)
		assert.Empty(t, dispatcher.dispatched, "no event should be dispatched for already-published series")
	})

	t.Run("returns 404 when series not found", func(t *testing.T) {
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) { return nil, nil },
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.Publish(ctx, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSeriesNotFound, appErr.Code)
	})
}
