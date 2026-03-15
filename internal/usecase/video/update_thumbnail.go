package video

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateThumbnailInput holds the data needed to update a video's thumbnail.
type UpdateThumbnailInput struct {
	VideoID     uuid.UUID
	File        io.Reader
	FileName    string
	ContentType string
}

// UpdateThumbnail replaces a video's thumbnail image.
func (s *Service) UpdateThumbnail(ctx context.Context, input UpdateThumbnailInput) (*entity.Video, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		s.log.Error("getting video for thumbnail update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	// Determine thumbnail extension
	thumbExt := filepath.Ext(input.FileName)
	if thumbExt == "" {
		thumbExt = thumbnailExtFromContentType(input.ContentType)
	}

	thumbnailKey := fmt.Sprintf("videos/%s/thumbnail%s", input.VideoID.String(), thumbExt)

	// Upload new thumbnail
	if _, err := s.storage.Upload(ctx, thumbnailKey, input.File, input.ContentType); err != nil {
		s.log.Error("uploading thumbnail", "error", err)
		return nil, apperror.Internal(apperror.CodeUploadFailed, "failed to upload thumbnail", err)
	}

	// Delete old thumbnail if it exists and is different
	if video.ThumbnailKey != "" && video.ThumbnailKey != thumbnailKey {
		_ = s.storage.Delete(ctx, video.ThumbnailKey)
	}

	video.ThumbnailKey = thumbnailKey
	video.UpdatedAt = time.Now().UTC()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("updating video thumbnail key", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video", err)
	}

	return video, nil
}
