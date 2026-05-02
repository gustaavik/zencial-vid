package dto

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string          `json:"email" example:"user@example.com"`
	Role      string          `json:"role" example:"user"`
	Status    string          `json:"status" example:"active"`
	Profile   ProfileResponse `json:"profile"`
	CreatedAt string          `json:"created_at" example:"2025-01-01T00:00:00Z"`
}

// ProfileResponse represents a user profile in API responses.
type ProfileResponse struct {
	DisplayName string  `json:"display_name" example:"John Doe"`
	AvatarURL   string  `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    string  `json:"language" example:"en"`
	Country     string  `json:"country" example:"Denmark"`
}

// UpdateProfileRequest represents a profile update request.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100" example:"Jane Doe"`
	AvatarURL   *string `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    *string `json:"language,omitempty" validate:"omitempty" example:"en"`
	Country     *string `json:"country,omitempty" validate:"omitempty" example:"Denmark"`
}

// UpdateStatusRequest represents a user status update (admin).
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended" example:"suspended"`
}

// AdminCreateUserRequest represents an admin-initiated user creation request.
type AdminCreateUserRequest struct {
	Email       string  `json:"email" validate:"required,email" example:"user@example.com"`
	Password    string  `json:"password" validate:"required,min=8,max=128" example:"securepassword"`
	Role        string  `json:"role,omitempty" validate:"omitempty,oneof=user admin" example:"user"`
	DisplayName string  `json:"display_name" validate:"required,min=3,max=100" example:"Jane Doe"`
	AvatarURL   string  `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    string  `json:"language,omitempty" validate:"omitempty" example:"en"`
	Country     string  `json:"country,omitempty" validate:"omitempty" example:"Denmark"`
}

// AdminUpdateUserRequest represents an admin-initiated user update request.
type AdminUpdateUserRequest struct {
	Email       *string `json:"email,omitempty" validate:"omitempty,email" example:"user@example.com"`
	Role        *string `json:"role,omitempty" validate:"omitempty,oneof=user admin" example:"admin"`
	Password    *string `json:"password,omitempty" validate:"omitempty,min=8,max=128" example:"newpassword"`
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100" example:"Jane Doe"`
	AvatarURL   *string `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	DateOfBirth *string `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Language    *string `json:"language,omitempty" validate:"omitempty" example:"en"`
	Country     *string `json:"country,omitempty" validate:"omitempty" example:"Denmark"`
}
