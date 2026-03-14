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

func TestService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("suspend active user", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, dispatcher)

		result, appErr := svc.UpdateStatus(ctx, UpdateStatusInput{
			UserID: user.ID,
			Status: entity.UserStatusSuspended,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, entity.UserStatusSuspended, result.Status)

		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.UserStatusChanged)
		require.True(t, ok)
		assert.Equal(t, user.ID, evt.UserID)
		assert.Equal(t, "suspended", evt.NewStatus)
	})

	t.Run("activate suspended user", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusSuspended
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		result, appErr := svc.UpdateStatus(ctx, UpdateStatusInput{
			UserID: user.ID,
			Status: entity.UserStatusActive,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, entity.UserStatusActive, result.Status)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
		}, nil)

		result, appErr := svc.UpdateStatus(ctx, UpdateStatusInput{
			UserID: uuid.New(),
			Status: entity.UserStatusSuspended,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("repository update status error", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
			updateStatusFn: func(_ context.Context, _ uuid.UUID, _ entity.UserStatus) error {
				return fmt.Errorf("db error")
			},
		}, nil)

		result, appErr := svc.UpdateStatus(ctx, UpdateStatusInput{
			UserID: user.ID,
			Status: entity.UserStatusSuspended,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
