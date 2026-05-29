package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// CheckHandleInput holds the data needed to check handle availability.
type CheckHandleInput struct {
	Handle        string
	RequestingUID uuid.UUID
}

// CheckHandleOutput holds the result of a handle availability check.
type CheckHandleOutput struct {
	Available bool
}

// CheckHandle reports whether the given handle is available for the requesting user.
func (s *Service) CheckHandle(ctx context.Context, input CheckHandleInput) (*CheckHandleOutput, *apperror.AppError) {
	if input.Handle == "" {
		return nil, apperror.BadRequest(apperror.CodeBadRequest, "handle must not be empty", nil)
	}

	taken, err := s.userRepo.HandleExists(ctx, input.Handle, input.RequestingUID)
	if err != nil {
		s.log.Error("checking handle existence", "error", err, "handle", input.Handle)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to check handle availability", err)
	}

	return &CheckHandleOutput{Available: !taken}, nil
}
