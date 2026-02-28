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
	assert.Equal(t, RoleUser, user.Role, "default role should be user")
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
			user.Role = tt.role
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

func TestUser_CanAccessContent_NilDOB(t *testing.T) {
	user := newTestUser(t)
	assert.Nil(t, user.Profile.DateOfBirth, "DOB should be nil for this test")

	// Without DOB, only unrestricted content is accessible
	assert.True(t, user.CanAccessContent(valueobject.RatingG), "should access G with nil DOB")
	assert.True(t, user.CanAccessContent(valueobject.RatingPG), "should access PG with nil DOB")
	assert.False(t, user.CanAccessContent(valueobject.RatingPG13), "should not access PG13 with nil DOB")
	assert.False(t, user.CanAccessContent(valueobject.RatingR), "should not access R with nil DOB")
	assert.False(t, user.CanAccessContent(valueobject.RatingNC17), "should not access NC17 with nil DOB")
}

func TestUser_CanAccessContent_WithDOB(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		age     int
		rating  valueobject.ContentRating
		allowed bool
	}{
		// Child (age 10)
		{"child can access G", 10, valueobject.RatingG, true},
		{"child can access PG", 10, valueobject.RatingPG, true},
		{"child cannot access PG13", 10, valueobject.RatingPG13, false},
		{"child cannot access R", 10, valueobject.RatingR, false},
		{"child cannot access NC17", 10, valueobject.RatingNC17, false},

		// Teenager (age 14)
		{"teen can access G", 14, valueobject.RatingG, true},
		{"teen can access PG", 14, valueobject.RatingPG, true},
		{"teen can access PG13", 14, valueobject.RatingPG13, true},
		{"teen cannot access R", 14, valueobject.RatingR, false},
		{"teen cannot access NC17", 14, valueobject.RatingNC17, false},

		// Young adult (age 17)
		{"17yo can access G", 17, valueobject.RatingG, true},
		{"17yo can access PG13", 17, valueobject.RatingPG13, true},
		{"17yo can access R", 17, valueobject.RatingR, true},
		{"17yo cannot access NC17", 17, valueobject.RatingNC17, false},

		// Adult (age 25)
		{"adult can access G", 25, valueobject.RatingG, true},
		{"adult can access PG13", 25, valueobject.RatingPG13, true},
		{"adult can access R", 25, valueobject.RatingR, true},
		{"adult can access NC17", 25, valueobject.RatingNC17, true},

		// Boundary: exactly 18
		{"18yo can access NC17", 18, valueobject.RatingNC17, true},

		// Boundary: exactly 13
		{"13yo can access PG13", 13, valueobject.RatingPG13, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := newTestUser(t)
			dob := now.AddDate(-tt.age, 0, 0)
			user.Profile.DateOfBirth = &dob

			assert.Equal(t, tt.allowed, user.CanAccessContent(tt.rating))
		})
	}
}
