package video

import (
	"context"

	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Stats returns platform-wide catalog aggregates for the admin dashboard:
// counts by status and submission status plus a per-genre title count.
func (s *Service) Stats(ctx context.Context) (*repository.VideoStats, *apperror.AppError) {
	stats, err := s.videoRepo.Stats(ctx)
	if err != nil {
		s.log.Error("computing video stats", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to compute video stats", err)
	}
	return stats, nil
}
