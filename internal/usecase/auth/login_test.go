package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func newActiveUser() *entity.User {
	email := valueobject.EmailFromTrusted("user@example.com")
	return entity.NewUser(email, valueobject.NewHashedPassword("hashed-password"))
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return user, nil
				},
			},
			nil, nil, nil, dispatcher,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "user@example.com",
			Password: "password123",
		})

		require.Nil(t, appErr)
		require.NotNil(t, output)
		assert.Equal(t, user.ID, output.User.ID)
		require.NotNil(t, output.TokenPair)

		// Verify login event dispatched
		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.UserLoggedIn)
		assert.True(t, ok)
	})

	t.Run("invalid email format returns bad request", func(t *testing.T) {
		svc := newTestService(nil, nil, nil, nil, nil)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "invalid",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("user not found returns unauthorized", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return nil, nil // Not found
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "unknown@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("suspended user returns forbidden", func(t *testing.T) {
		user := newActiveUser()
		user.Suspend()
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return user, nil
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserSuspended, appErr.Code)
	})

	t.Run("deleted user returns unauthorized", func(t *testing.T) {
		user := newActiveUser()
		user.SoftDelete()
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return user, nil
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("wrong password returns unauthorized", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return user, nil
				},
			},
			nil,
			&mockHasher{
				compareFn: func(_, _ string) error {
					return fmt.Errorf("password mismatch")
				},
			},
			nil, nil,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "user@example.com",
			Password: "wrong-password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("database error returns internal error", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return nil, fmt.Errorf("database error")
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Login(ctx, LoginInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
