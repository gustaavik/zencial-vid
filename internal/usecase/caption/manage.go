package caption

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// InitiateCaptionUploadInput holds inputs for starting a caption upload.
type InitiateCaptionUploadInput struct {
	VideoID      uuid.UUID
	UploaderID   uuid.UUID
	LanguageCode string
	Format       string // webvtt | srt
}

// InitiateCaptionUploadOutput contains the signed upload URL.
type InitiateCaptionUploadOutput struct {
	UploadURL string
	ObjectKey string
	ExpiresAt time.Time
}

// InitiateCaptionUpload returns a signed PUT URL for uploading a caption file.
func (s *Service) InitiateCaptionUpload(ctx context.Context, input *InitiateCaptionUploadInput) (*InitiateCaptionUploadOutput, *apperror.AppError) {
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

	ext := "vtt"
	contentType := "text/vtt"
	if input.Format == "srt" {
		ext = "srt"
		contentType = "text/plain"
	}
	objectKey := fmt.Sprintf("captions/%s/%s.%s", input.VideoID, input.LanguageCode, ext)

	expiry := 30 * time.Minute
	uploadURL, signErr := s.storage.PresignedPutURL(ctx, objectKey, contentType, expiry)
	if signErr != nil {
		return nil, apperror.Internal(apperror.CodeStorageError, "failed to sign caption upload URL", signErr)
	}

	expiresAt := time.Now().UTC().Add(expiry)
	return &InitiateCaptionUploadOutput{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresAt: expiresAt,
	}, nil
}

// RegisterCaptionInput holds inputs for completing a caption upload.
type RegisterCaptionInput struct {
	VideoID      uuid.UUID
	UploaderID   uuid.UUID
	LanguageCode string
	Format       string
	StorageKey   string
	Source       entity.CaptionSource
}

// RegisterCaption persists a caption record after the file has been uploaded.
func (s *Service) RegisterCaption(ctx context.Context, input *RegisterCaptionInput) (*entity.Caption, *apperror.AppError) {
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

	caption := entity.NewCaption(input.VideoID, input.LanguageCode, input.Format, input.StorageKey, input.Source)
	if err := s.captionRepo.Upsert(ctx, caption); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to save caption", err)
	}
	return caption, nil
}

// ListCaptions returns all captions for a video.
func (s *Service) ListCaptions(ctx context.Context, videoID uuid.UUID) ([]entity.Caption, *apperror.AppError) {
	captions, err := s.captionRepo.ListByVideo(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list captions", err)
	}
	return captions, nil
}

// DeleteCaption removes a caption for a specific language.
func (s *Service) DeleteCaption(ctx context.Context, videoID uuid.UUID, languageCode string, uploaderID uuid.UUID) *apperror.AppError {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil || video.UploadedBy != uploaderID {
		return apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}

	if err := s.captionRepo.DeleteByVideoAndLang(ctx, videoID, languageCode); err != nil {
		return apperror.Internal(apperror.CodeInternalError, "failed to delete caption", err)
	}
	return nil
}

// PublishCaption marks a caption as reviewed and published.
func (s *Service) PublishCaption(ctx context.Context, videoID uuid.UUID, languageCode string, uploaderID uuid.UUID) (*entity.Caption, *apperror.AppError) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch video", err)
	}
	if video == nil || video.UploadedBy != uploaderID {
		return nil, apperror.Forbidden(apperror.CodeVideoOwnershipRequired, "you do not own this video", nil)
	}

	caption, err := s.captionRepo.GetByVideoAndLang(ctx, videoID, languageCode)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to fetch caption", err)
	}
	if caption == nil {
		return nil, apperror.NotFound(apperror.CodeCaptionNotFound, "caption not found", nil)
	}

	caption.MarkPublished()
	if err := s.captionRepo.Update(ctx, caption); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to update caption", err)
	}
	return caption, nil
}
