package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	sessionuc "github.com/zenfulcode/zencial/internal/usecase/session"
)

// SessionToResponse maps a Session entity to a SessionResponse DTO. Pass
// uuid.Nil as currentSessionID when listing for an admin (no session is
// considered "current" from the admin's perspective).
func SessionToResponse(s *entity.Session, currentSessionID uuid.UUID) dto.SessionResponse {
	return dto.SessionResponse{
		ID:                s.ID.String(),
		UserID:            s.UserID.String(),
		DeviceName:        s.DeviceName,
		UserAgent:         s.UserAgent,
		IPAddress:         s.IPAddress,
		CreatedAt:         s.CreatedAt.Format(time.RFC3339),
		LastActivityAt:    s.LastActivityAt.Format(time.RFC3339),
		ExpiresAt:         s.ExpiresAt().Format(time.RFC3339),
		AbsoluteExpiresAt: s.AbsoluteExpiresAt.Format(time.RFC3339),
		IsCurrent:        currentSessionID != uuid.Nil && s.ID == currentSessionID,
	}
}

// SessionsToResponse maps a slice of Session entities to SessionResponse DTOs.
func SessionsToResponse(sessions []entity.Session, currentSessionID uuid.UUID) []dto.SessionResponse {
	result := make([]dto.SessionResponse, len(sessions))
	for i := range sessions {
		result[i] = SessionToResponse(&sessions[i], currentSessionID)
	}
	return result
}

// RevokeOthersToResponse maps a session use-case revoke-output to its DTO.
func RevokeOthersToResponse(out *sessionuc.RevokeOthersOutput) dto.RevokeOthersResponse {
	return dto.RevokeOthersResponse{RevokedCount: out.RevokedCount}
}
