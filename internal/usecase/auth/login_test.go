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
		sessionRepo := &mockSessionRepo{}
		svc := newTestService(
			&mockUserRepo{
				getByEmailFn: func(_ context.Context, _ valueobject.Email) (*entity.User, error) {
					return user, nil
				},
			},
			sessionRepo, nil, nil, dispatcher,
		)

		output, appErr := svc.Login(ctx, &LoginInput{
			Email:    "user@example.com",
			Password: "password123",
			Session: SessionContext{
				DeviceName: "MacBook",
				UserAgent:  "Mozilla/5.0",
				IPAddress:  "10.0.0.1",
			},
		})

		require.Nil(t, appErr)
		require.NotNil(t, output)
		assert.Equal(t, user.ID, output.User.ID)
		require.NotNil(t, output.Session)
		assert.NotEmpty(t, output.Token)
		require.Len(t, sessionRepo.created, 1)
		created := sessionRepo.created[0]
		assert.Equal(t, "MacBook", created.DeviceName)
		assert.Equal(t, "Mozilla/5.0", created.UserAgent)
		assert.Equal(t, "10.0.0.1", created.IPAddress)
		assert.Equal(t, fixedNow, created.CreatedAt)

		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.UserLoggedIn)
		require.True(t, ok)
		assert.Equal(t, output.Session.ID, evt.SessionID)
	})

	t.Run("invalid email format returns bad request", func(t *testing.T) {
		svc := newTestService(nil, nil, nil, nil, nil)

		output, appErr := svc.Login(ctx, &LoginInput{
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
					return nil, nil
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Login(ctx, &LoginInput{
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

		output, appErr := svc.Login(ctx, &LoginInput{
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

		output, appErr := svc.Login(ctx, &LoginInput{
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
			nil, nil,
			&mockHasher{
				compareFn: func(_, _ string) error {
					return fmt.Errorf("password mismatch")
				},
			},
			nil,
		)

		output, appErr := svc.Login(ctx, &LoginInput{
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

		output, appErr := svc.Login(ctx, &LoginInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
