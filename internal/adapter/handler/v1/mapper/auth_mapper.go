package mapper

import (
	"time"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// AuthToResponse maps a successful login/register result to the AuthResponse
// DTO. The raw token is included exactly once; subsequent requests must use
// it as a bearer token.
func AuthToResponse(user *entity.User, session *entity.Session, token string) dto.AuthResponse {
	return dto.AuthResponse{
		User:      UserToResponse(user),
		Token:     token,
		ExpiresAt: session.ExpiresAt().Format(time.RFC3339),
		SessionID: session.ID.String(),
	}
}
