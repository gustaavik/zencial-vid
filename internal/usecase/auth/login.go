package auth

import (
	"context"
	"time"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/auth"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// LoginInput holds login credentials.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput holds the result of a successful login.
type LoginOutput struct {
	User      *entity.User
	TokenPair *auth.TokenPair
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, *apperror.AppError) {
	email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid email address", err)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("getting user by email", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to find user", err)
	}
	if user == nil {
		return nil, apperror.Unauthorized(apperror.CodeInvalidCredentials, "invalid email or password", domain.ErrInvalidCredentials)
	}

	if !user.IsActive() {
		if user.Status == entity.UserStatusSuspended {
			return nil, apperror.Forbidden(apperror.CodeUserSuspended, "account is suspended", domain.ErrUserSuspended)
		}
		return nil, apperror.Unauthorized(apperror.CodeInvalidCredentials, "invalid email or password", domain.ErrInvalidCredentials)
	}

	if err := s.hasher.Compare(user.PasswordHash.String(), input.Password); err != nil {
		return nil, apperror.Unauthorized(apperror.CodeInvalidCredentials, "invalid email or password", domain.ErrInvalidCredentials)
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

	s.dispatcher.Dispatch(event.UserLoggedIn{
		UserID:    user.ID,
		Timestamp: time.Now(),
	})

	return &LoginOutput{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}
