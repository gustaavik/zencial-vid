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

// AuthResponse represents the authentication response with the issued
// session token. The raw token is returned only here; subsequent requests
// authenticate by sending it as Authorization: Bearer <token>.
type AuthResponse struct {
	User      UserResponse `json:"user"`
	Token     string       `json:"token" example:"K7zP_-Qm…"`
	ExpiresAt string       `json:"expires_at" example:"2025-04-01T12:34:56Z"`
	SessionID string       `json:"session_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}
