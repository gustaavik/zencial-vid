package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		var deletedToken string
		svc := newTestService(
			nil, nil, nil,
			&mockSessionStore{
				deleteRefreshTokenFn: func(_ context.Context, token string) error {
					deletedToken = token
					return nil
				},
			},
			nil,
		)

		appErr := svc.Logout(ctx, LogoutInput{
			RefreshToken: "my-refresh-token",
		})

		require.Nil(t, appErr)
		assert.Equal(t, "my-refresh-token", deletedToken)
	})

	t.Run("session store error returns internal error", func(t *testing.T) {
		svc := newTestService(
			nil, nil, nil,
			&mockSessionStore{
				deleteRefreshTokenFn: func(_ context.Context, _ string) error {
					return fmt.Errorf("redis connection failed")
				},
			},
			nil,
		)

		appErr := svc.Logout(ctx, LogoutInput{
			RefreshToken: "my-refresh-token",
		})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
