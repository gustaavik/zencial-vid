package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	authpkg "github.com/zenfulcode/zencial/internal/infrastructure/auth"
)

// UserToResponse maps a User entity to a UserResponse DTO.
func UserToResponse(user *entity.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email.String(),
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Profile:   ProfileToResponse(&user.Profile),
	}
	return resp
}

// ProfileToResponse maps a UserProfile to a ProfileResponse DTO.
func ProfileToResponse(profile *entity.UserProfile) dto.ProfileResponse {
	resp := dto.ProfileResponse{
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Language:    profile.Language,
		Country:     profile.Country,
	}
	if profile.DateOfBirth != nil {
		dob := profile.DateOfBirth.Format("2006-01-02")
		resp.DateOfBirth = &dob
	}
	return resp
}

// AuthToResponse maps user and token pair to an AuthResponse DTO.
func AuthToResponse(user *entity.User, tokenPair *authpkg.TokenPair) dto.AuthResponse {
	return dto.AuthResponse{
		User:         UserToResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}
}

// TokenPairToResponse maps a token pair to a TokenResponse DTO.
func TokenPairToResponse(tokenPair *authpkg.TokenPair) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UsersToResponse maps a slice of User entities to UserResponse DTOs.
func UsersToResponse(users []entity.User) []dto.UserResponse {
	result := make([]dto.UserResponse, len(users))
	for i := range users {
		result[i] = UserToResponse(&users[i])
	}
	return result
}
