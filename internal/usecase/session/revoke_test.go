package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestRevokeMine(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		sessionID := uuid.New()
		var revokedID uuid.UUID

		repo := &mockSessionRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Session, error) {
				return &entity.Session{ID: id, UserID: userID}, nil
			},
			revokeFn: func(_ context.Context, id uuid.UUID, _ time.Time) error {
				revokedID = id
				return nil
			},
		}
		dispatcher := &mockDispatcher{}
		svc := newTestService(repo, dispatcher)

		appErr := svc.RevokeMine(ctx, userID, sessionID)
		require.Nil(t, appErr)
		assert.Equal(t, sessionID, revokedID)
		require.Len(t, dispatcher.dispatched, 1)
	})

	t.Run("ownership mismatch returns NOT_FOUND", func(t *testing.T) {
		userID := uuid.New()
		otherUserID := uuid.New()
		sessionID := uuid.New()
		repo := &mockSessionRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Session, error) {
				return &entity.Session{ID: id, UserID: otherUserID}, nil
			},
		}
		svc := newTestService(repo, nil)

		appErr := svc.RevokeMine(ctx, userID, sessionID)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSessionNotFound, appErr.Code)
	})

	t.Run("not found returns NOT_FOUND", func(t *testing.T) {
		repo := &mockSessionRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Session, error) {
				return nil, nil
			},
		}
		svc := newTestService(repo, nil)
		appErr := svc.RevokeMine(ctx, uuid.New(), uuid.New())
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSessionNotFound, appErr.Code)
	})
}

func TestRevokeOthers(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	current := uuid.New()

	var capturedException uuid.UUID
	repo := &mockSessionRepo{
		revokeOthersForUserFn: func(_ context.Context, _, except uuid.UUID, _ time.Time) (int64, error) {
			capturedException = except
			return 3, nil
		},
	}
	svc := newTestService(repo, nil)

	out, appErr := svc.RevokeOthers(ctx, userID, current)
	require.Nil(t, appErr)
	assert.EqualValues(t, 3, out.RevokedCount)
	assert.Equal(t, current, capturedException)
}

func TestAdminRevoke(t *testing.T) {
	ctx := context.Background()
	actor := uuid.New()
	target := uuid.New()
	sessionID := uuid.New()

	t.Run("success dispatches event with admin actor", func(t *testing.T) {
		repo := &mockSessionRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Session, error) {
				return &entity.Session{ID: sessionID, UserID: target}, nil
			},
		}
		dispatcher := &mockDispatcher{}
		svc := newTestService(repo, dispatcher)

		appErr := svc.AdminRevoke(ctx, sessionID, actor)
		require.Nil(t, appErr)
		require.Len(t, dispatcher.dispatched, 1)
	})

	t.Run("not found returns NOT_FOUND", func(t *testing.T) {
		repo := &mockSessionRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Session, error) {
				return nil, nil
			},
		}
		svc := newTestService(repo, nil)
		appErr := svc.AdminRevoke(ctx, sessionID, actor)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeSessionNotFound, appErr.Code)
	})
}

func TestAdminRevokeAll(t *testing.T) {
	ctx := context.Background()
	repo := &mockSessionRepo{
		revokeAllForUserFn: func(_ context.Context, _ uuid.UUID, _ time.Time) (int64, error) {
			return 2, nil
		},
	}
	dispatcher := &mockDispatcher{}
	svc := newTestService(repo, dispatcher)

	out, appErr := svc.AdminRevokeAll(ctx, uuid.New(), uuid.New())
	require.Nil(t, appErr)
	assert.EqualValues(t, 2, out.RevokedCount)
	require.Len(t, dispatcher.dispatched, 1)
}
