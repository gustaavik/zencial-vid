package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// UserRole represents a user's role in the system.
type UserRole string

const (
	RoleUser      UserRole = "user"
	RolePublisher UserRole = "publisher"
	RoleAdmin     UserRole = "admin"
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
	ID               uuid.UUID
	Email            valueobject.Email
	PasswordHash     valueobject.HashedPassword
	Roles            []UserRole
	Status           UserStatus
	StripeCustomerID *string
	Profile          UserProfile
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProfileLink is a single user-defined link (e.g. website, social).
type ProfileLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ProfilePreferences stores per-user content preference toggles.
type ProfilePreferences struct {
	AllowMatureContent  bool `json:"allow_mature_content"`
	AutoplayNextEpisode bool `json:"autoplay_next_episode"`
	AlwaysShowSubtitles bool `json:"always_show_subtitles"`
	ShowPaidFirstInFeed bool `json:"show_paid_first_in_feed"`
}

// ProfilePrivacy stores visibility settings for profile sections.
// Values are "Public", "Followers", or "Private".
type ProfilePrivacy struct {
	ProfileVisibility string `json:"profile_visibility"`
	WatchHistory      string `json:"watch_history"`
	Watchlist         string `json:"watchlist"`
	Tipping           string `json:"tipping"`
}

// UserProfile holds additional user information.
type UserProfile struct {
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   string
	DateOfBirth *time.Time
	Language    string // ISO 639-1
	Country     string // ISO 3166-1 alpha-2
	Handle      *string
	Pronouns    *string
	Headline    *string
	Bio         *string
	Links       []ProfileLink
	Preferences ProfilePreferences
	Privacy     ProfilePrivacy
	UpdatedAt   time.Time
}

// NewUser creates a new User with default values.
func NewUser(email valueobject.Email, passwordHash valueobject.HashedPassword) *User {
	now := time.Now().UTC()
	id := uuid.Must(uuid.NewV7())
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Roles:        []UserRole{RoleUser},
		Status:       UserStatusActive,
		Profile: UserProfile{
			UserID:   id,
			Language: "en",
			Links:    []ProfileLink{},
			Privacy: ProfilePrivacy{
				ProfileVisibility: "Public",
				WatchHistory:      "Public",
				Watchlist:         "Public",
				Tipping:           "Public",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasRole reports whether a roles slice contains the given role.
func HasRole(roles []UserRole, role UserRole) bool {
	return slices.Contains(roles, role)
}

// HasRole reports whether the user holds the given role.
func (u *User) HasRole(role UserRole) bool {
	return HasRole(u.Roles, role)
}

// IsActive reports whether the user account is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin reports whether the user is an admin.
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// IsPublisher reports whether the user is a publisher.
func (u *User) IsPublisher() bool {
	return u.HasRole(RolePublisher)
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
