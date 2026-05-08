package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// ListOutput holds the listing result.
type ListOutput struct {
	Sessions []entity.Session
	Total    int64
}

// ListMine returns active sessions for the calling user.
func (s *Service) ListMine(
	ctx context.Context,
	userID uuid.UUID,
	fs *filter.FilterSet,
) (*ListOutput, *apperror.AppError) {
	sessions, total, err := s.sessionRepo.ListByUserID(ctx, userID, fs)
	if err != nil {
		s.log.Error("listing sessions", "error", err, "user_id", userID)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list sessions", err)
	}
	return &ListOutput{Sessions: sessions, Total: total}, nil
}

// AdminList returns active sessions for any user. Authorization (admin role)
// is enforced at the middleware layer; this method does not double-check.
func (s *Service) AdminList(
	ctx context.Context,
	targetUserID uuid.UUID,
	fs *filter.FilterSet,
) (*ListOutput, *apperror.AppError) {
	return s.ListMine(ctx, targetUserID, fs)
}
