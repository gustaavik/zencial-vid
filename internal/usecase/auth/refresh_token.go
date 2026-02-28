package auth

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// RefreshInput holds the refresh token.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput holds the new token pair.
type RefreshOutput struct {
	TokenPair *auth.TokenPair
}

// RefreshToken exchanges a valid refresh token for a new token pair.
func (s *Service) RefreshToken(ctx context.Context, input RefreshInput) (*RefreshOutput, *apperror.AppError) {
	userID, err := s.sessionStore.GetUserIDByRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, apperror.Unauthorized(apperror.CodeRefreshTokenInvalid, "invalid refresh token", domain.ErrRefreshTokenNotFound)
	}

	// Delete old refresh token (rotation)
	if err := s.sessionStore.DeleteRefreshToken(ctx, input.RefreshToken); err != nil {
		s.log.Error("deleting old refresh token", "error", err)
	}

	// Get user to include current role
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperror.Unauthorized(apperror.CodeInvalidCredentials, "user not found", domain.ErrUserNotFound)
	}

	if !user.IsActive() {
		return nil, apperror.Forbidden(apperror.CodeUserSuspended, "account is not active", domain.ErrUserSuspended)
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		s.log.Error("generating tokens", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to generate tokens", err)
	}

	if err := s.sessionStore.StoreRefreshToken(ctx, tokenPair.RefreshToken, user.ID); err != nil {
		s.log.Error("storing refresh token", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to store session", err)
	}

	return &RefreshOutput{TokenPair: tokenPair}, nil
}
