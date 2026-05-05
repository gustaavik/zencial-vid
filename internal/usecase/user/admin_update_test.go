package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

func TestService_AdminUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("noop returns user unchanged and dispatches profile event", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, dispatcher, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{UserID: user.ID})

		require.Nil(t, appErr)
		require.NotNil(t, result)
		assert.Equal(t, user.Email.String(), result.Email.String())
		require.Len(t, dispatcher.dispatched, 1)
		_, ok := dispatcher.dispatched[0].(event.UserProfileUpdated)
		assert.True(t, ok)
	})

	t.Run("profile update", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID:      user.ID,
			DisplayName: new("New Name"),
			Language:    new("sv"),
			Country:     new("SE"),
		})

		require.Nil(t, appErr)
		assert.Equal(t, "New Name", result.Profile.DisplayName)
		assert.Equal(t, "sv", result.Profile.Language)
		assert.Equal(t, "SE", result.Profile.Country)
	})

	t.Run("email change checks conflict", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
			existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
				return false, nil
			},
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID: user.ID,
			Email:  new("new@example.com"),
		})

		require.Nil(t, appErr)
		assert.Equal(t, "new@example.com", result.Email.String())
	})

	t.Run("email conflict returns 409", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
			existsByEmailFn: func(_ context.Context, _ valueobject.Email) (bool, error) {
				return true, nil
			},
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID: user.ID,
			Email:  new("taken@example.com"),
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeEmailAlreadyExists, appErr.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		user := newActiveUser()
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID: user.ID,
			Email:  new("not-an-email"),
		})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeValidationFailed, appErr.Code)
	})

	t.Run("role change emits UserRoleChanged event", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, dispatcher, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID: user.ID,
			Role:   new(entity.RoleAdmin),
		})

		require.Nil(t, appErr)
		assert.Equal(t, entity.RoleAdmin, result.Role)

		var found bool
		for _, evt := range dispatcher.dispatched {
			if rc, ok := evt.(event.UserRoleChanged); ok {
				found = true
				assert.Equal(t, "user", rc.OldRole)
				assert.Equal(t, "admin", rc.NewRole)
			}
		}
		assert.True(t, found, "UserRoleChanged should have been dispatched")
	})

	t.Run("role unchanged does not emit role event", func(t *testing.T) {
		user := newActiveUser()
		dispatcher := &mockDispatcher{}
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, dispatcher, nil)

		_, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID: user.ID,
			Role:   new(entity.RoleUser),
		})
		require.Nil(t, appErr)

		for _, evt := range dispatcher.dispatched {
			_, ok := evt.(event.UserRoleChanged)
			assert.False(t, ok, "should not dispatch role change when role is the same")
		}
	})

	t.Run("password reset re-hashes", func(t *testing.T) {
		user := newActiveUser()
		var hashed string
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, nil, &mockHasher{
			hashFn: func(p string) (string, error) {
				hashed = "h:" + p
				return hashed, nil
			},
		})

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{
			UserID:   user.ID,
			Password: new("brand-new"),
		})

		require.Nil(t, appErr)
		assert.Equal(t, "h:brand-new", result.PasswordHash.String())
	})

	t.Run("user not found", func(t *testing.T) {
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return nil, nil },
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{UserID: uuid.New()})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})

	t.Run("deleted user is treated as not found", func(t *testing.T) {
		user := newActiveUser()
		user.Status = entity.UserStatusDeleted
		svc := newTestServiceWithHasher(&mockUserRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.User, error) { return user, nil },
		}, nil, nil)

		result, appErr := svc.AdminUpdate(ctx, &AdminUpdateInput{UserID: user.ID})

		assert.Nil(t, result)
		require.NotNil(t, appErr)
		assert.Equal(t, apperror.CodeUserNotFound, appErr.Code)
	})
}
