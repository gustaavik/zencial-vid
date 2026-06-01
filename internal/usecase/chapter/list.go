package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ListChapters returns all chapters for a video ordered by start time.
func (s *Service) ListChapters(ctx context.Context, videoID uuid.UUID) ([]entity.Chapter, *apperror.AppError) {
	chapters, err := s.chapterRepo.ListByVideo(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list chapters", err)
	}
	return chapters, nil
}
