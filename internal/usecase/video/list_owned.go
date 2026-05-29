package video

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
)

// ListOwned returns all videos uploaded by the given uploader (for publisher dashboard).
func (s *Service) ListOwned(ctx context.Context, uploaderID uuid.UUID, fs *filter.FilterSet) ([]entity.Video, int64, *apperror.AppError) {
	videos, total, err := s.videoRepo.ListByUploader(ctx, uploaderID, fs)
	if err != nil {
		s.log.Error("listing owned videos", "error", err)
		return nil, 0, apperror.Internal(apperror.CodeInternalError, "failed to list videos", err)
	}
	return videos, total, nil
}
