package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

func TestService_List(t *testing.T) {
	ctx := context.Background()
	defaultFS := filter.FilterSet{
		Pagination: valueobject.Pagination{Page: 1, PerPage: 20},
	}

	t.Run("success with results", func(t *testing.T) {
		users := []entity.User{*newActiveUser(), *newActiveUser()}
		svc := newTestService(&mockUserRepo{
			listFn: func(_ context.Context, fs *filter.FilterSet) ([]entity.User, int64, error) {
				assert.Equal(t, 1, fs.Pagination.Page)
				assert.Equal(t, 20, fs.Pagination.PerPage)
				return users, 2, nil
			},
		}, nil)

		result, total, appErr := svc.List(ctx, &defaultFS)

		require.Nil(t, appErr)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
	})

	t.Run("empty result", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			listFn: func(_ context.Context, _ *filter.FilterSet) ([]entity.User, int64, error) {
				return []entity.User{}, 0, nil
			},
		}, nil)

		result, total, appErr := svc.List(ctx, &defaultFS)

		require.Nil(t, appErr)
		assert.Empty(t, result)
		assert.Equal(t, int64(0), total)
	})

	t.Run("repository error", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			listFn: func(_ context.Context, _ *filter.FilterSet) ([]entity.User, int64, error) {
				return nil, 0, fmt.Errorf("db error")
			},
		}, nil)

		result, total, appErr := svc.List(ctx, &defaultFS)

		assert.Nil(t, result)
		assert.Equal(t, int64(0), total)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
