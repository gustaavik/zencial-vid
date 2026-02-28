package dto

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8,max=128" example:"securepassword123"`
	Name     string `json:"name" validate:"required,min=1,max=100" example:"John Doe"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"securepassword123"`
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"abc123def456"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"abc123def456"`
}

// AuthResponse represents the authentication response with tokens.
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string       `json:"refresh_token" example:"abc123def456"`
	ExpiresAt    string       `json:"expires_at" example:"2025-01-01T00:15:00Z"`
}

// TokenResponse represents a token refresh response.
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"abc123def456"`
	ExpiresAt    string `json:"expires_at" example:"2025-01-01T00:15:00Z"`
}
