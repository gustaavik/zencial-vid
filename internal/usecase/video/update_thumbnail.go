package video

import (
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/pkg/actor"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UpdateThumbnailInput holds the data needed to update a video's thumbnail.
type UpdateThumbnailInput struct {
	VideoID     uuid.UUID
	File        io.Reader
	FileName    string
	ContentType string
	// CallerID and CallerRole enforce publisher ownership when CallerRole == RolePublisher.
	CallerID   uuid.UUID
	CallerRole entity.UserRole
}

// UpdateThumbnail replaces a video's thumbnail image. Bytes are streamed to
// the CDN over the internal network — this API never writes thumbnails to S3
// directly, so the frontend never needs an S3 host for reads or writes.
func (s *Service) UpdateThumbnail(ctx context.Context, input *UpdateThumbnailInput) (*entity.Video, *apperror.AppError) {
	if s.cdn == nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "CDN not configured", nil)
	}

	video, err := s.videoRepo.GetByID(ctx, input.VideoID)
	if err != nil {
		s.log.Error("getting video for thumbnail update", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to get video", err)
	}
	if video == nil {
		return nil, apperror.NotFound(apperror.CodeVideoNotFound, "video not found", domain.ErrVideoNotFound)
	}

	if input.CallerRole == entity.RolePublisher && video.UploadedBy != input.CallerID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", domain.ErrVideoOwnershipRequired)
	}

	// Determine thumbnail extension
	thumbExt := filepath.Ext(input.FileName)
	if thumbExt == "" {
		thumbExt = thumbnailExtFromContentType(input.ContentType)
	}

	thumbnailKey, uploadErr := s.cdn.UploadThumbnail(ctx, input.VideoID.String(), thumbExt, input.ContentType, input.File)
	if uploadErr != nil {
		s.log.Error("uploading thumbnail to CDN", "error", uploadErr)
		return nil, apperror.Internal(apperror.CodeUploadFailed, "failed to upload thumbnail", uploadErr)
	}

	// Delete old thumbnail if the extension changed (resulting in a new key).
	// Same-key replacement is handled by the new PUT overwriting in S3.
	if video.ThumbnailKey != "" && video.ThumbnailKey != thumbnailKey {
		_ = s.storage.Delete(ctx, video.ThumbnailKey)
	}

	video.ThumbnailKey = thumbnailKey
	video.UpdatedAt = time.Now().UTC()

	if err := s.videoRepo.Update(ctx, video); err != nil {
		s.log.Error("updating video thumbnail key", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update video", err)
	}

	if err := s.dispatcher.Dispatch(event.VideoUpdated{
		VideoID:   video.ID,
		ActorID:   actor.FromContext(ctx),
		Field:     "thumbnail",
		Timestamp: time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching video updated event", "error", err)
	}

	return video, nil
}
