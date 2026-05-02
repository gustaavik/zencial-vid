package user

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

func TestService_AdminDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("success soft-deletes and dispatches event", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		var updated *entity.User
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
			updateFn: func(_ context.Context, u *entity.User) error {
				updated = u
				return nil
			},
		}, dispatcher, nil)

		appErr := svc.AdminDelete(ctx, user.ID)
		require.Nil(t, appErr)
		require.NotNil(t, updated)
		assert.Equal(t, entity.UserStatusDeleted, updated.Status)

		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.UserAccountDeleted)
		assert.True(t, ok)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return nil, nil },
		}, nil, nil)

		appErr := svc.AdminDelete(ctx, uuid.New())
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("already deleted returns not found", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, nil, nil)

		appErr := svc.AdminDelete(ctx, user.ID)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})
}
