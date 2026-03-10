package video

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// UploadInput holds the data needed to upload a video.
type UploadInput struct {
	Title         string
	Description   string
	Creator       string
	ContentRating string
	Quality       string
	GenreIDs      []uuid.UUID
	File          io.Reader
	FileName      string
	ContentType   string
	FileSize      int64
	UploadedBy    uuid.UUID
}

// Upload uploads a video file and creates its metadata record.
func (s *Service) Upload(ctx context.Context, input UploadInput) (*entity.Video, *apperror.AppError) {
	contentRating := input.ContentRating
	if contentRating == "" {
		contentRating = "G"
	}

	quality := input.Quality
	if quality == "" {
		quality = "HD"
	}

	// Generate slug from title
	slug, err := valueobject.NewSlug(input.Title)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug generation", err)
	}
	slug = slug.WithRandomID()

	// Upload file to MinIO
	videoID := uuid.New()
	ext := extensionFromContentType(input.ContentType)
	if ext == "" {
		ext = filepath.Ext(input.FileName)
	}
	if ext == "" {
		ext = ".mp4"
	}
	storageKey := fmt.Sprintf("videos/%s/original%s", videoID.String(), ext)

	_, err = s.storage.Upload(ctx, storageKey, input.File, input.ContentType)
	if err != nil {
		s.log.Error("uploading video to storage", "error", err)
		return nil, apperror.Internal(apperror.CodeUploadFailed, "failed to upload video", err)
	}

	// Create video entity
	video := entity.NewVideo(
		input.Title, slug, input.Description, input.Creator,
		contentRating, quality, storageKey, input.ContentType,
		input.FileSize, input.UploadedBy,
	)
	video.ID = videoID
	video.SetGenres(input.GenreIDs)

	if err := s.videoRepo.Create(ctx, video); err != nil {
		// Attempt cleanup of uploaded file
		_ = s.storage.Delete(ctx, storageKey)
		s.log.Error("creating video record", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to save video", err)
	}

	s.dispatcher.Dispatch(event.VideoUploaded{
		VideoID:    video.ID,
		Title:      video.Title,
		UploadedBy: input.UploadedBy,
		Timestamp:  time.Now(),
	})

	return video, nil
}

func extensionFromContentType(ct string) string {
	switch ct {
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/x-matroska":
		return ".mkv"
	default:
		return ""
	}
}
