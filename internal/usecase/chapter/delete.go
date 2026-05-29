package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// DeleteChapter removes a single chapter by ID.
func (s *Service) DeleteChapter(ctx context.Context, chapterID, uploaderID uuid.UUID) *apperror.AppError {
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch chapter", err)
	}
	if ch == nil {
		return apperror.NotFound(apperror.CodeChapterNotFound, "chapter not found", nil)
	}

	video, err := s.videoRepo.GetByID(ctx, ch.VideoID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil || video.UploadedBy != uploaderID {
		return apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}

	if err := s.chapterRepo.DeleteByID(ctx, chapterID); err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to delete chapter", err)
	}
	return nil
}
