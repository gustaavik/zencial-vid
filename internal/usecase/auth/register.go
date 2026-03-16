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

// RegisterInput holds the data needed to register a user.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// RegisterOutput holds the result of a successful registration.
type RegisterOutput struct {
	User      *entity.User
	TokenPair *auth.TokenPair
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, *apperror.AppError) {
	email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid email address", err)
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		s.log.Error("checking email existence", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check email", err)
	}
	if exists {
		return nil, apperror.Conflict(apperror.CodeEmailAlreadyExists, "email already registered", domain.ErrEmailAlreadyExists)
	}

	hashedPassword, err := s.hasher.Hash(input.Password)
	if err != nil {
		s.log.Error("hashing password", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to process password", err)
	}

	user := entity.NewUser(email, valueobject.NewHashedPassword(hashedPassword))
	user.Profile.DisplayName = input.Name

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("creating user", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create user", err)
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

	if err := s.dispatcher.Dispatch(event.UserRegistered{
		UserID:    user.ID,
		Email:     user.Email.String(),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching user registered event", "error", err)
	}

	return &RegisterOutput{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}
