package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Delete removes content by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	if err := s.contentRepo.Delete(ctx, id); err != nil {
		s.log.Error("deleting content", "error", err)
		return apperror.Internal(apperror.CodeInternalError, "failed to delete content", err)
	}
	return nil
}
