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

func TestService_UpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("full update", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, dispatcher)

		displayName := "New Name"
		avatarURL := "https://example.com/avatar.jpg"
		dob := "1990-05-15"
		lang := "sv"
		country := "SE"

		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DisplayName: &displayName,
			AvatarURL:   &avatarURL,
			DateOfBirth: &dob,
			Language:    &lang,
			Country:     &country,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, "New Name", result.Profile.DisplayName)
		assert.Equal(t, "https://example.com/avatar.jpg", result.Profile.AvatarURL)
		assert.NotNil(t, result.Profile.DateOfBirth)
		assert.Equal(t, "1990-05-15", result.Profile.DateOfBirth.Format("2006-01-02"))
		assert.Equal(t, "sv", result.Profile.Language)
		assert.Equal(t, "SE", result.Profile.Country)

		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.UserProfileUpdated)
		assert.True(t, ok)
	})

	t.Run("partial update only display name", func(t *testing.T) {
		user := newActiveUser()
		user.Profile.Language = "en"
		user.Profile.Country = "Denmark"
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		displayName := "Updated Name"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DisplayName: &displayName,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Name", result.Profile.DisplayName)
		assert.Equal(t, "en", result.Profile.Language)
		assert.Equal(t, "Denmark", result.Profile.Country)
	})

	t.Run("clear date of birth with empty string", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		emptyDOB := ""
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DateOfBirth: &emptyDOB,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Nil(t, result.Profile.DateOfBirth)
	})

	t.Run("invalid date format", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		badDate := "not-a-date"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DateOfBirth: &badDate,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInvalidDateFormat, appErr.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
		}, nil)

		name := "Test"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      uuid.New(),
			DisplayName: &name,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("deleted user returns not found", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		name := "Test"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DisplayName: &name,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("repository update error", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
			updateFn: func(_ context.Context, _ *entity.User) error {
				return fmt.Errorf("db error")
			},
		}, nil)

		name := "Test"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			DisplayName: &name,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	})
}
