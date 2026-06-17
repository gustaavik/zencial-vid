package series

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func TestService_Update(t *testing.T) {
	ctx := context.Background()

	newOwnedSeries := func(ownerID uuid.UUID) *entity.Series {
		slug, _ := valueobject.NewSlug("bedford-street")
		return entity.NewSeries("Bedford Street", slug.WithRandomID(), "desc", "creator", ownerID)
	}

	t.Run("applies provided extended fields and leaves the rest untouched", func(t *testing.T) {
		ownerID := uuid.New()
		existing := newOwnedSeries(ownerID)

		var updated *entity.Series
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) {
				return existing, nil
			},
			updateFn: func(_ context.Context, s *entity.Series) error {
				updated = s
				return nil
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		seriesType := "documentary"
		logline := "Updated logline"
		rating := "TV-MA"
		hide := true
		monet := []string{"ads", "ppv"}
		_, appErr := svc.Update(ctx, &UpdateInput{
			ID:                  existing.ID,
			CallerID:            ownerID,
			SeriesType:          &seriesType,
			Logline:             &logline,
			ContentRating:       &rating,
			HideEpisodeCount:    &hide,
			DefaultMonetization: monet,
		})

		require.Nil(t, appErr)
		require.NotNil(t, updated)
		assert.Equal(t, entity.SeriesTypeDocumentary, updated.SeriesType)
		assert.Equal(t, "Updated logline", updated.Logline)
		assert.Equal(t, "TV-MA", updated.ContentRating)
		assert.True(t, updated.HideEpisodeCount)
		assert.Equal(t, monet, updated.DefaultMonetization)
		// Untouched fields keep their original values.
		assert.Equal(t, "Bedford Street", updated.Title)
		assert.Equal(t, "en", updated.PrimaryLanguage)
		assert.True(t, updated.BingeMode)
	})

	t.Run("rejects update from a non-owner non-admin", func(t *testing.T) {
		existing := newOwnedSeries(uuid.New())
		repo := &mockSeriesRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Series, error) {
				return existing, nil
			},
		}
		svc := newTestService(repo, nil, nil, nil)

		_, appErr := svc.Update(ctx, &UpdateInput{
			ID:       existing.ID,
			CallerID: uuid.New(), // different user
		})

		require.NotNil(t, appErr)
	})
}
