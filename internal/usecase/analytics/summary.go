package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// GetSummary returns aggregate statistics for all videos uploaded by the given uploader.
func (s *Service) GetSummary(ctx context.Context, uploaderID uuid.UUID) ([]repository.VideoStats, *apperror.AppError) {
	stats, err := s.analyticsRepo.GetUploaderSummary(ctx, uploaderID)
	if err != nil {
		s.log.Error("getting uploader summary", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get analytics summary", err)
	}
	return stats, nil
}
