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

func TestService_GetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
				assert.Equal(t, user.ID, id)
				return user, nil
			},
		}, nil)

		result, appErr := svc.GetProfile(ctx, user.ID)

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, "Test User", result.Profile.DisplayName)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
		}, nil)

		result, appErr := svc.GetProfile(ctx, uuid.New())

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("deleted user returns not found", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		result, appErr := svc.GetProfile(ctx, user.ID)

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("repository error returns internal error", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, fmt.Errorf("db error")
			},
		}, nil)

		result, appErr := svc.GetProfile(ctx, uuid.New())

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
