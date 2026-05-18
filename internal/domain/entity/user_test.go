package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

func newTestUser(t *testing.T) *User {
	t.Helper()
	email := valueobject.EmailFromTrusted("test@example.com")
	hash := valueobject.NewHashedPassword("hashed-password-123")
	return NewUser(email, hash)
}

func TestNewUser(t *testing.T) {
	email := valueobject.EmailFromTrusted("user@example.com")
	hash := valueobject.NewHashedPassword("$2a$10$somehash")

	user := NewUser(email, hash)

	require.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID, "ID should be generated")
	assert.Equal(t, "user@example.com", user.Email.String())
	assert.Equal(t, "$2a$10$somehash", user.PasswordHash.String())
	assert.Equal(t, []UserRole{RoleUser}, user.Roles, "default roles should be [user]")
	assert.Equal(t, UserStatusActive, user.Status, "default status should be active")
	assert.Equal(t, "en", user.Profile.Language, "default language should be en")
	assert.Equal(t, user.ID, user.Profile.UserID, "profile UserID should match user ID")
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.Nil(t, user.Profile.DateOfBirth, "DateOfBirth should be nil by default")
	assert.Empty(t, user.Profile.DisplayName, "DisplayName should be empty by default")
	assert.Empty(t, user.Profile.AvatarURL, "AvatarURL should be empty by default")
	assert.Empty(t, user.Profile.Country, "Country should be empty by default")
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"active user", UserStatusActive, true},
		{"suspended user", UserStatusSuspended, false},
		{"deleted user", UserStatusDeleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := newTestUser(t)
			user.Status = tt.status
			assert.Equal(t, tt.want, user.IsActive())
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"admin user", RoleAdmin, true},
		{"regular user", RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := newTestUser(t)
			user.Roles = []UserRole{tt.role}
			assert.Equal(t, tt.want, user.IsAdmin())
		})
	}
}

func TestUser_Suspend(t *testing.T) {
	user := newTestUser(t)
	assert.True(t, user.IsActive())

	beforeUpdate := user.UpdatedAt
	time.Sleep(time.Millisecond) // ensure time difference
	user.Suspend()

	assert.Equal(t, UserStatusSuspended, user.Status)
	assert.False(t, user.IsActive())
	assert.True(t, user.UpdatedAt.After(beforeUpdate), "UpdatedAt should be updated")
}

func TestUser_Activate(t *testing.T) {
	user := newTestUser(t)
	user.Suspend()
	assert.False(t, user.IsActive())

	beforeUpdate := user.UpdatedAt
	time.Sleep(time.Millisecond)
	user.Activate()

	assert.Equal(t, UserStatusActive, user.Status)
	assert.True(t, user.IsActive())
	assert.True(t, user.UpdatedAt.After(beforeUpdate), "UpdatedAt should be updated")
}

func TestUser_SoftDelete(t *testing.T) {
	user := newTestUser(t)
	assert.True(t, user.IsActive())

	beforeUpdate := user.UpdatedAt
	time.Sleep(time.Millisecond)
	user.SoftDelete()

	assert.Equal(t, UserStatusDeleted, user.Status)
	assert.False(t, user.IsActive())
	assert.True(t, user.UpdatedAt.After(beforeUpdate), "UpdatedAt should be updated")
}
