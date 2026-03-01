package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success - rotates token", func(t *testing.T) {
		userID := uuid.New()
		user := entity.NewUser(
			valueobject.EmailFromTrusted("user@example.com"),
			valueobject.NewHashedPassword("hash"),
		)
		user.ID = userID

		var deletedToken string
		var storedToken string

		svc := newTestService(
			&mockUserRepo{
				getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
					if id == userID {
						return user, nil
					}
					return nil, nil
				},
			},
			nil,
			nil,
			&mockSessionStore{
				getUserIDByRefreshTokenFn: func(_ context.Context, token string) (uuid.UUID, error) {
					if token == "old-refresh-token" {
						return userID, nil
					}
					return uuid.Nil, fmt.Errorf("not found")
				},
				deleteRefreshTokenFn: func(_ context.Context, token string) error {
					deletedToken = token
					return nil
				},
				storeRefreshTokenFn: func(_ context.Context, token string, _ uuid.UUID) error {
					storedToken = token
					return nil
				},
			},
			nil,
		)

		output, appErr := svc.RefreshToken(ctx, RefreshInput{
			RefreshToken: "old-refresh-token",
		})

		require.Nil(t, appErr)
		require.NotNil(t, output)
		require.NotNil(t, output.TokenPair)
		assert.Equal(t, "old-refresh-token", deletedToken, "old token should be deleted")
		assert.NotEmpty(t, storedToken, "new token should be stored")
	})

	t.Run("invalid refresh token returns unauthorized", func(t *testing.T) {
		svc := newTestService(
			nil, nil, nil,
			&mockSessionStore{
				getUserIDByRefreshTokenFn: func(_ context.Context, _ string) (uuid.UUID, error) {
					return uuid.Nil, fmt.Errorf("not found")
				},
			},
			nil,
		)

		output, appErr := svc.RefreshToken(ctx, RefreshInput{
			RefreshToken: "invalid-token",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeRefreshTokenInvalid, appErr.Code)
	})

	t.Run("user not found returns unauthorized", func(t *testing.T) {
		userID := uuid.New()
		svc := newTestService(
			&mockUserRepo{
				getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
					return nil, nil // not found
				},
			},
			nil, nil,
			&mockSessionStore{
				getUserIDByRefreshTokenFn: func(_ context.Context, _ string) (uuid.UUID, error) {
					return userID, nil
				},
			},
			nil,
		)

		output, appErr := svc.RefreshToken(ctx, RefreshInput{
			RefreshToken: "valid-token",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("suspended user returns forbidden", func(t *testing.T) {
		userID := uuid.New()
		user := entity.NewUser(
			valueobject.EmailFromTrusted("user@example.com"),
			valueobject.NewHashedPassword("hash"),
		)
		user.ID = userID
		user.Suspend()

		svc := newTestService(
			&mockUserRepo{
				getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.User, error) {
					return user, nil
				},
			},
			nil, nil,
			&mockSessionStore{
				getUserIDByRefreshTokenFn: func(_ context.Context, _ string) (uuid.UUID, error) {
					return userID, nil
				},
			},
			nil,
		)

		output, appErr := svc.RefreshToken(ctx, RefreshInput{
			RefreshToken: "valid-token",
		})

		assert.Nil(t, output)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserSuspended, appErr.Code)
	})
}
