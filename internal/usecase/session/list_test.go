package session

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

func TestListMine(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := &mockSessionRepo{
		listByUserIDFn: func(_ context.Context, id uuid.UUID, _ *filter.FilterSet) ([]entity.Session, int64, error) {
			assert.Equal(t, userID, id)
			return []entity.Session{{ID: uuid.New(), UserID: userID}}, 1, nil
		},
	}
	svc := newTestService(repo, nil)

	out, appErr := svc.ListMine(ctx, userID, &filter.FilterSet{})
	require.Nil(t, appErr)
	require.NotNil(t, out)
	assert.EqualValues(t, 1, out.Total)
	assert.Len(t, out.Sessions, 1)
}
