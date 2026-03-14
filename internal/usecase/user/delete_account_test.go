package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_DeleteAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := newActiveUser()
		var updatedUser *entity.User
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
			updateFn: func(_ context.Context, u *entity.User) error {
				updatedUser = u
				return nil
			},
		}, dispatcher)

		appErr := svc.DeleteAccount(ctx, user.ID)

		require.Nil(t, appErr)
		require.NotNil(t, updatedUser)
		assert.Equal(t, entity.UserStatusDeleted, updatedUser.Status)

		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.UserAccountDeleted)
		require.True(t, ok)
		assert.Equal(t, user.ID, evt.UserID)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
		}, nil)

		appErr := svc.DeleteAccount(ctx, uuid.New())

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("already deleted returns not found", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		appErr := svc.DeleteAccount(ctx, user.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
			updateFn: func(_ context.Context, _ *entity.User) error {
				return fmt.Errorf("db error")
			},
		}, nil)

		appErr := svc.DeleteAccount(ctx, user.ID)

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
