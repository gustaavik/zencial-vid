package user

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

func TestService_AdminCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success creates active user with default role", func(t *testing.T) {
		var created *entity.User
		dispatcher := &mockDispatcher{}
		svc := newTestServiceWithHasher(&mockUserRepo{
			existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
				return false, nil
			},
			createFn: func(_ context.Context, u *entity.User) error {
				created = u
				return nil
			},
		}, dispatcher, &mockHasher{})

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "new@example.com",
			Password:    "supersecret",
			DisplayName: "New User",
			Language:    "en",
			Country:     "Denmark",
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, "new@example.com", result.Email.String())
		assert.Equal(t, entity.RoleUser, result.Role)
		assert.Equal(t, entity.UserStatusActive, result.Status)
		assert.Equal(t, "New User", result.Profile.DisplayName)
		assert.NotNil(t, created)
		assert.Equal(t, "hashed:supersecret", result.PasswordHash.String())

		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.UserRegistered)
		assert.True(t, ok)
	})

	t.Run("success creates admin when role specified", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{}, nil, nil)

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "admin@example.com",
			Password:    "supersecret",
			Role:        entity.RoleAdmin,
			DisplayName: "Admin",
		})

		require.Nil(t, appErr)
		assert.Equal(t, entity.RoleAdmin, result.Role)
	})

	t.Run("invalid email", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{}, nil, nil)

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "not-an-email",
			Password:    "supersecret",
			DisplayName: "X",
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("duplicate email returns conflict", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{
			existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
				return true, nil
			},
		}, nil, nil)

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "exists@example.com",
			Password:    "supersecret",
			DisplayName: "X",
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeEmailAlreadyExists, appErr.Code)
	})

	t.Run("hasher failure", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{}, nil, &mockHasher{
			hashFn: func(_ string) (string, error) {
				return "", fmt.Errorf("boom")
			},
		})

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "ok@example.com",
			Password:    "supersecret",
			DisplayName: "X",
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})

	t.Run("invalid date of birth", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{}, nil, nil)

		bad := "not-a-date"
		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "ok@example.com",
			Password:    "supersecret",
			DisplayName: "X",
			DateOfBirth: &bad,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidDateFormat, appErr.Code)
	})

	t.Run("repository create error", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{
			createFn: func(_ context.Context, _ *entity.User) error {
				return fmt.Errorf("db down")
			},
		}, nil, nil)

		result, appErr := svc.AdminCreate(ctx, AdminCreateInput{
			Email:       "ok@example.com",
			Password:    "supersecret",
			DisplayName: "X",
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
