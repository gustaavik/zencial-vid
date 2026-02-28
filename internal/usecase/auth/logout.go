package auth

import (
	"context"

	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// LogoutInput holds the refresh token to invalidate.
type LogoutInput struct {
	RefreshToken string
}

// Logout invalidates a refresh token.
func (s *Service) Logout(ctx context.Context, input LogoutInput) *apperror.AppError {
	if err := s.sessionStore.DeleteRefreshToken(ctx, input.RefreshToken); err != nil {
		s.log.Error("deleting refresh token", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to logout", err)
	}
	return nil
}
