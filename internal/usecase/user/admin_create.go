package user

import (
	"context"
	"time"

	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// AdminCreateInput holds the data needed for an admin to create a user.
type AdminCreateInput struct {
	Email       string
	Password    string
	Roles       []entity.UserRole // optional; defaults to [RoleUser] when empty
	DisplayName string
	AvatarURL   string
	Language    string
	Country     string
	DateOfBirth *string // "YYYY-MM-DD" or nil
}

// AdminCreate creates a new user account on behalf of an admin.
func (s *Service) AdminCreate(ctx context.Context, input *AdminCreateInput) (*entity.User, *apperror.AppError) {
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

	hashed, err := s.hasher.Hash(input.Password)
	if err != nil {
		s.log.Error("hashing password", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to process password", err)
	}

	user := entity.NewUser(email, valueobject.NewHashedPassword(hashed))

	if len(input.Roles) > 0 {
		user.Roles = input.Roles
	}
	if input.DisplayName != "" {
		user.Profile.DisplayName = input.DisplayName
	}
	user.Profile.AvatarURL = input.AvatarURL
	if input.Language != "" {
		user.Profile.Language = input.Language
	}
	user.Profile.Country = input.Country
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		dob, parseErr := time.Parse("2006-01-02", *input.DateOfBirth)
		if parseErr != nil {
			return nil, apperror.BadRequest(apperror.CodeInvalidDateFormat, "date of birth must be in YYYY-MM-DD format", parseErr)
		}
		user.Profile.DateOfBirth = &dob
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("creating user", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create user", err)
	}

	if err := s.dispatcher.Dispatch(event.UserRegistered{
		UserID:    user.ID,
		Email:     user.Email.String(),
		ActorID:   actor.FromContext(ctx),
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching user registered event", "error", err)
	}

	return user, nil
}
