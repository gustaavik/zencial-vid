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

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		dispatcher := &mockDispatcher{}
		svc := newTestService(
			&mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
					return false, nil
				},
			},
			nil, nil, nil, dispatcher,
		)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "user@example.com",
			Password: "securePassword123",
			Name:     "John Doe",
		})

		require.Nil(t, appErr)
		require.NotNil(t, output)
		require.NotNil(t, output.User)
		require.NotNil(t, output.TokenPair)
		assert.Equal(t, "user@example.com", output.User.Email.String())
		assert.Equal(t, "John Doe", output.User.Profile.DisplayName)
		assert.Equal(t, entity.RoleUser, output.User.Role)
		assert.Equal(t, entity.UserStatusActive, output.User.Status)

		// Verify event was dispatched
		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.UserRegistered)
		require.True(t, ok)
		assert.Equal(t, output.User.ID, evt.UserID)
		assert.Equal(t, "user@example.com", evt.Email)
	})

	t.Run("invalid email returns validation error", func(t *testing.T) {
		svc := newTestService(nil, nil, nil, nil, nil)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "not-an-email",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("duplicate email returns conflict error", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
					return true, nil
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "existing@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeEmailAlreadyExists, appErr.Code)
	})

	t.Run("email check failure returns internal error", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
					return false, fmt.Errorf("db connection failed")
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("password hash failure returns internal error", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
					return false, nil
				},
			},
			nil,
			&mockHasher{
				hashFn: func(_ string) (string, error) {
					return "", fmt.Errorf("hash failed")
				},
			},
			nil, nil,
		)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("repo create failure returns internal error", func(t *testing.T) {
		svc := newTestService(
			&mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
					return false, nil
				},
				createFn: func(_ context.Context, _ *entity.User) error {
					return fmt.Errorf("db write failed")
				},
			},
			nil, nil, nil, nil,
		)

		output, appErr := svc.Register(ctx, RegisterInput{
			Email:    "user@example.com",
			Password: "password",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
