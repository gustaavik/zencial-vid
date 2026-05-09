package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_Logout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	sessionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		var revoked uuid.UUID
		var revokedAt time.Time
		dispatcher := &mockDispatcher{}
		svc := newTestService(
			nil,
			&mockSessionRepo{
				revokeFn: func(_ context.Context, id uuid.UUID, t time.Time) error {
					revoked = id
					revokedAt = t
					return nil
				},
			},
			nil, nil, dispatcher,
		)

		appErr := svc.Logout(ctx, LogoutInput{UserID: userID, SessionID: sessionID})

		require.Nil(t, appErr)
		assert.Equal(t, sessionID, revoked)
		assert.Equal(t, fixedNow, revokedAt)
		require.Len(t, dispatcher.dispatched, 1)
		evt, ok := dispatcher.dispatched[0].(event.UserLoggedOut)
		require.True(t, ok)
		assert.Equal(t, sessionID, evt.SessionID)
		assert.Equal(t, userID, evt.UserID)
	})

	t.Run("session repo error returns internal error", func(t *testing.T) {
		svc := newTestService(
			nil,
			&mockSessionRepo{
				revokeFn: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
					return fmt.Errorf("db error")
				},
			},
			nil, nil, nil,
		)

		appErr := svc.Logout(ctx, LogoutInput{UserID: userID, SessionID: sessionID})

		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
