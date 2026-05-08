package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// LogoutInput identifies the session to revoke. UserID is used purely for
// audit purposes; the session is revoked by ID regardless.
type LogoutInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// Logout revokes the session associated with the bearer token.
func (s *Service) Logout(ctx context.Context, input LogoutInput) *apperror.AppError {
	now := s.clock.Now()
	if err := s.sessionRepo.Revoke(ctx, input.SessionID, now); err != nil {
		s.log.Error("revoking session", "error", err, "session_id", input.SessionID)
		return apperror.Internal(apperror.CodeInternalError, "failed to logout", err)
	}

	if err := s.dispatcher.Dispatch(event.UserLoggedOut{
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Timestamp: now,
	}); err != nil {
		s.log.Error("dispatching user logged out event", "error", err)
	}

	return nil
}
