package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		result, appErr := svc.GetByID(ctx, user.ID)

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
	})

	t.Run("returns deleted user (admin can see all)", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		result, appErr := svc.GetByID(ctx, user.ID)

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, entity.UserStatusDeleted, result.Status)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
		}, nil)

		result, appErr := svc.GetByID(ctx, uuid.New())

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, fmt.Errorf("db error")
			},
		}, nil)

		result, appErr := svc.GetByID(ctx, uuid.New())

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
