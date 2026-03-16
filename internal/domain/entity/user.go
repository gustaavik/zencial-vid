package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// UserRole represents a user's role in the system.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus represents a user's account status.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User is the core user entity.
type User struct {
	ID           uuid.UUID
	Email        valueobject.Email
	PasswordHash valueobject.HashedPassword
	Role         UserRole
	Status       UserStatus
	Profile      UserProfile
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserProfile holds additional user information.
type UserProfile struct {
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   string
	DateOfBirth *time.Time
	Language    string // ISO 639-1
	Country     string // ISO 3166-1 alpha-2
	UpdatedAt   time.Time
}

// NewUser creates a new User with default values.
func NewUser(email valueobject.Email, passwordHash valueobject.HashedPassword) *User {
	now := time.Now().UTC()
	id := uuid.New()
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		Status:       UserStatusActive,
		Profile: UserProfile{
			UserID:   id,
			Language: "en",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive reports whether the user account is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin reports whether the user is an admin.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// Suspend marks the user account as suspended.
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now().UTC()
}

// Activate marks the user account as active.
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
}

// SoftDelete marks the user account as deleted.
func (u *User) SoftDelete() {
	u.Status = UserStatusDeleted
	u.UpdatedAt = time.Now().UTC()
}
