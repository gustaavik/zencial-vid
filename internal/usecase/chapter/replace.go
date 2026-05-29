package chapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// ReplaceChaptersInput holds the full chapter list to persist.
type ReplaceChaptersInput struct {
	VideoID    uuid.UUID
	UploaderID uuid.UUID
	Chapters   []ChapterItem
}

// ChapterItem represents a single chapter in a replace request.
type ChapterItem struct {
	StartTimeSecs int
	Title         string
	Source        entity.ChapterSource
}

// ReplaceChapters atomically replaces all chapters for a video.
func (s *Service) ReplaceChapters(ctx context.Context, input *ReplaceChaptersInput) ([]entity.Chapter, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", nil)
	}
	if video.UploadedBy != input.UploaderID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}

	chapters := make([]entity.Chapter, len(input.Chapters))
	for i, item := range input.Chapters {
		ch := entity.NewChapter(input.VideoID, item.StartTimeSecs, item.Title)
		ch.Source = item.Source
		chapters[i] = *ch
	}

	if err := s.chapterRepo.BulkReplace(ctx, input.VideoID, chapters); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to replace chapters", err)
	}

	result, err := s.chapterRepo.ListByVideo(ctx, input.VideoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch chapters", err)
	}
	return result, nil
}
