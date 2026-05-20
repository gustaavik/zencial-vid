package dto

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string          `json:"email" example:"user@example.com"`
	Roles     []string        `json:"roles" example:"[\"user\"]"`
	Status    string          `json:"status" example:"active"`
	Profile   ProfileResponse `json:"profile" `
	CreatedAt string          `json:"created_at" example:"2025-01-01T00:00:00Z"`
}

// ProfileLinkDTO is a single link entry in a profile response.
type ProfileLinkDTO struct {
	Label string `json:"label" example:"Website"`
	URL   string `json:"url" example:"https://example.com"`
}

// ProfilePreferencesDTO holds content preference toggles in a profile response.
type ProfilePreferencesDTO struct {
	AllowMatureContent  bool `json:"allow_mature_content" example:"true"`
	AutoplayNextEpisode bool `json:"autoplay_next_episode" example:"false"`
	AlwaysShowSubtitles bool `json:"always_show_subtitles" example:"true"`
	ShowPaidFirstInFeed bool `json:"show_paid_first_in_feed" example:"false"`
}

// ProfilePrivacyDTO holds visibility settings in a profile response.
type ProfilePrivacyDTO struct {
	ProfileVisibility string `json:"profile_visibility" example:"Public"`
	WatchHistory      string `json:"watch_history"      example:"Public"`
	Watchlist         string `json:"watchlist"          example:"Public"`
	Tipping           string `json:"tipping"            example:"Public"`
}

// ProfileResponse represents a user profile in API responses.
type ProfileResponse struct {
	DisplayName string                `json:"display_name" example:"John Doe"`
	AvatarURL   string                `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string               `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    string                `json:"language" example:"en"`
	Country     string                `json:"country" example:"Denmark"`
	Handle      *string               `json:"handle,omitempty" example:"johndoe"`
	Pronouns    *string               `json:"pronouns,omitempty" example:"he/him"`
	Headline    *string               `json:"headline,omitempty" example:"Filmmaker · New York"`
	Bio         *string               `json:"bio,omitempty" example:"I make films."`
	Links       []ProfileLinkDTO      `json:"links"`
	Preferences ProfilePreferencesDTO `json:"preferences"`
	Privacy     ProfilePrivacyDTO     `json:"privacy"`
}

// ProfileLinkRequest is a single link entry in a profile update request.
type ProfileLinkRequest struct {
	Label string `json:"label" validate:"required,max=50"`
	URL   string `json:"url"   validate:"required,max=500"`
}

// ProfilePreferencesRequest holds preference toggle updates (all fields optional).
type ProfilePreferencesRequest struct {
	AllowMatureContent  *bool `json:"allow_mature_content,omitempty"`
	AutoplayNextEpisode *bool `json:"autoplay_next_episode,omitempty"`
	AlwaysShowSubtitles *bool `json:"always_show_subtitles,omitempty"`
	ShowPaidFirstInFeed *bool `json:"show_paid_first_in_feed,omitempty"`
}

// ProfilePrivacyRequest holds visibility setting updates (all fields optional).
type ProfilePrivacyRequest struct {
	ProfileVisibility *string `json:"profile_visibility,omitempty" validate:"omitempty,oneof=Public Followers Private" example:"Public"`
	WatchHistory      *string `json:"watch_history,omitempty"      validate:"omitempty,oneof=Public Followers Private" example:"Public"`
	Watchlist         *string `json:"watchlist,omitempty"          validate:"omitempty,oneof=Public Followers Private" example:"Public"`
	Tipping           *string `json:"tipping,omitempty"            validate:"omitempty,oneof=Public Followers Private" example:"Public"`
}

// UpdateProfileRequest represents a profile update request.
type UpdateProfileRequest struct {
	DisplayName *string                    `json:"display_name,omitempty" validate:"omitempty,min=1,max=100" example:"Jane Doe"`
	AvatarURL   *string                    `json:"avatar_url,omitempty"   validate:"omitempty,url"           example:"https://example.com/avatar.jpg"`
	DateOfBirth *string                    `json:"date_of_birth,omitempty"                                   example:"1990-01-15"`
	Language    *string                    `json:"language,omitempty"     validate:"omitempty"               example:"en"`
	Country     *string                    `json:"country,omitempty"      validate:"omitempty"               example:"Denmark"`
	Handle      *string                    `json:"handle,omitempty"       validate:"omitempty,min=1,max=50,alphanum" example:"johndoe"`
	Pronouns    *string                    `json:"pronouns,omitempty"     validate:"omitempty,max=50"        example:"he/him"`
	Headline    *string                    `json:"headline,omitempty"     validate:"omitempty,max=200"       example:"Filmmaker · New York"`
	Bio         *string                    `json:"bio,omitempty"          validate:"omitempty,max=1000"      example:"I make films."`
	Links       []ProfileLinkRequest       `json:"links,omitempty"        validate:"omitempty,max=5,dive" example:"[{\"label\":\"Website\",\"url\":\"https://example.com\"}]"`
	Preferences *ProfilePreferencesRequest `json:"preferences,omitempty" example:"{\"allow_mature_content\":true}"`
	Privacy     *ProfilePrivacyRequest     `json:"privacy,omitempty" example:"{\"profile_visibility\":\"Public\"}"`
}

// UpdateStatusRequest represents a user status update (admin).
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended" example:"suspended"`
}

// AdminCreateUserRequest represents an admin-initiated user creation request.
type AdminCreateUserRequest struct {
	Email       string   `json:"email" validate:"required,email" example:"user@example.com"`
	Password    string   `json:"password" validate:"required,min=8,max=128" example:"securepassword"`
	Roles       []string `json:"roles,omitempty" validate:"omitempty,dive,oneof=user publisher admin" example:"[\"user\"]"`
	DisplayName string   `json:"display_name" validate:"required,min=3,max=100" example:"Jane Doe"`
	AvatarURL   string   `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string  `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    string   `json:"language,omitempty" validate:"omitempty" example:"en"`
	Country     string   `json:"country,omitempty" validate:"omitempty" example:"Denmark"`
}

// AdminUpdateUserRequest represents an admin-initiated user update request.
type AdminUpdateUserRequest struct {
	Email       *string  `json:"email,omitempty" validate:"omitempty,email" example:"user@example.com"`
	Roles       []string `json:"roles,omitempty" validate:"omitempty,dive,oneof=user publisher admin" example:"[\"user\",\"admin\"]"`
	Password    *string  `json:"password,omitempty" validate:"omitempty,min=8,max=128" example:"newpassword"`
	DisplayName *string  `json:"display_name,omitempty" validate:"omitempty,min=1,max=100" example:"Jane Doe"`
	AvatarURL   *string  `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string  `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    *string  `json:"language,omitempty" validate:"omitempty" example:"en"`
	Country     *string  `json:"country,omitempty" validate:"omitempty" example:"Denmark"`
}
