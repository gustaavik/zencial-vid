package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_CheckHandle(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("available handle", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			handleExistsFn: func(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
				return false, nil
			},
		}, nil)

		out, appErr := svc.CheckHandle(ctx, CheckHandleInput{Handle: "freehandle", RequestingUID: userID})

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.True(t, out.Available)
	})

	t.Run("taken handle", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			handleExistsFn: func(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		}, nil)

		out, appErr := svc.CheckHandle(ctx, CheckHandleInput{Handle: "takenhandle", RequestingUID: userID})

		require.Nil(t, appErr)
		require.NotNil(t, out)
		assert.False(t, out.Available)
	})

	t.Run("empty handle returns bad request", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{}, nil)

		out, appErr := svc.CheckHandle(ctx, CheckHandleInput{Handle: "", RequestingUID: userID})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
	})

	t.Run("repository error returns internal error", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			handleExistsFn: func(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
				return false, fmt.Errorf("db error")
			},
		}, nil)

		out, appErr := svc.CheckHandle(ctx, CheckHandleInput{Handle: "somehandle", RequestingUID: userID})

		assert.Nil(t, out)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
