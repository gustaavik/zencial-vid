package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// LoginInput holds login credentials and per-request session metadata.
type LoginInput struct {
	Email    string
	Password string
	Session  SessionContext
}

// LoginOutput holds the result of a successful login.
type LoginOutput struct {
	User      *entity.User
	Session   *entity.Session
	Token     string    // raw token returned once; only its hash is persisted
	ExpiresAt time.Time // effective deadline (earlier of idle/absolute)
}

// Login authenticates a user and creates a new session.
func (s *Service) Login(ctx context.Context, input *LoginInput) (*LoginOutput, *apperror.AppError) {
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

	session, token, appErr := s.createSession(ctx, user.ID, input.Session)
	if appErr != nil {
		return nil, appErr
	}

	if err := s.dispatcher.Dispatch(event.UserLoggedIn{
		UserID:    user.ID,
		SessionID: session.ID,
		Timestamp: session.CreatedAt,
	}); err != nil {
		s.log.Error("dispatching user logged in event", "error", err)
	}

	return &LoginOutput{
		User:      user,
		Session:   session,
		Token:     token,
		ExpiresAt: session.ExpiresAt(),
	}, nil
}

// createSession is shared by Login and Register: generate token, persist
// session row, return both. Returns an apperror.AppError on failure.
func (s *Service) createSession(
	ctx context.Context,
	userID uuid.UUID,
	sessCtx SessionContext,
) (*entity.Session, string, *apperror.AppError) {
	token, hash, err := s.tokens.Generate()
	if err != nil {
		s.log.Error("generating session token", "error", err)
		return nil, "", apperror.Internal(apperror.CodeInternalError, "failed to generate token", err)
	}

	now := s.clock.Now()
	session := entity.NewSession(
		userID,
		hash,
		sessCtx.DeviceName,
		sessCtx.UserAgent,
		sessCtx.IPAddress,
		now,
		s.cfg.IdleTimeout,
		s.cfg.AbsoluteTimeout,
	)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.log.Error("creating session", "error", err)
		return nil, "", apperror.Internal(apperror.CodeInternalError, "failed to create session", err)
	}
	return session, token, nil
}
