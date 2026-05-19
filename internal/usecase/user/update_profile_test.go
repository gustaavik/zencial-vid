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

	t.Run("update handle bio headline pronouns", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		handle := "testuser"
		bio := "I make films."
		headline := "Filmmaker · NYC"
		pronouns := "they/them"

		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:   user.ID,
			Handle:   &handle,
			Bio:      &bio,
			Headline: &headline,
			Pronouns: &pronouns,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		require.NotNil(t, result.Profile.Handle)
		assert.Equal(t, "testuser", *result.Profile.Handle)
		require.NotNil(t, result.Profile.Bio)
		assert.Equal(t, "I make films.", *result.Profile.Bio)
		require.NotNil(t, result.Profile.Headline)
		assert.Equal(t, "Filmmaker · NYC", *result.Profile.Headline)
		require.NotNil(t, result.Profile.Pronouns)
		assert.Equal(t, "they/them", *result.Profile.Pronouns)
	})

	t.Run("handle conflict returns error", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
			handleExistsFn: func(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		}, nil)

		handle := "taken"
		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID: user.ID,
			Handle: &handle,
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeHandleAlreadyExists, appErr.Code)
	})

	t.Run("update links replaces slice", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		links := []entity.ProfileLink{
			{Label: "Website", URL: "https://example.com"},
			{Label: "GitHub", URL: "https://github.com/test"},
		}

		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID: user.ID,
			Links:  &links,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Len(t, result.Profile.Links, 2)
		assert.Equal(t, "Website", result.Profile.Links[0].Label)
	})

	t.Run("update preferences", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		prefs := entity.ProfilePreferences{
			AllowMatureContent:  true,
			AutoplayNextEpisode: false,
			AlwaysShowSubtitles: true,
			ShowPaidFirstInFeed: false,
		}

		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:      user.ID,
			Preferences: &prefs,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.True(t, result.Profile.Preferences.AllowMatureContent)
		assert.True(t, result.Profile.Preferences.AlwaysShowSubtitles)
		assert.False(t, result.Profile.Preferences.AutoplayNextEpisode)
	})

	t.Run("update privacy settings", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestService(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) {
				return user, nil
			},
		}, nil)

		privacy := entity.ProfilePrivacy{
			ProfileVisibility: "Followers",
			WatchHistory:      "Private",
			Watchlist:         "Public",
			Tipping:           "Followers",
		}

		result, appErr := svc.UpdateProfile(ctx, UpdateProfileInput{
			UserID:  user.ID,
			Privacy: &privacy,
		})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, "Followers", result.Profile.Privacy.ProfileVisibility)
		assert.Equal(t, "Private", result.Profile.Privacy.WatchHistory)
	})
}
