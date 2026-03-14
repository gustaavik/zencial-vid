package video

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/thumbnail"
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

	// Optional thumbnail upload. If nil, a thumbnail is extracted from the video.
	Thumbnail            io.Reader
	ThumbnailFileName    string
	ThumbnailContentType string
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

	videoID := uuid.New()
	ext := extensionFromContentType(input.ContentType)
	if ext == "" {
		ext = filepath.Ext(input.FileName)
	}
	if ext == "" {
		ext = ".mp4"
	}
	storageKey := fmt.Sprintf("videos/%s/original%s", videoID.String(), ext)

	// Save video to temp file so we can both upload to storage and extract a thumbnail.
	tmpVideo, err := os.CreateTemp("", "video-upload-*"+ext)
	if err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to create temp file", err)
	}
	defer os.Remove(tmpVideo.Name())
	defer tmpVideo.Close()

	if _, err := io.Copy(tmpVideo, input.File); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to write temp file", err)
	}

	// Upload video to storage from temp file
	if _, err := tmpVideo.Seek(0, io.SeekStart); err != nil {
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to seek temp file", err)
	}
	if _, err := s.storage.Upload(ctx, storageKey, tmpVideo, input.ContentType); err != nil {
		s.log.Error("uploading video to storage", "error", err)
		return nil, apperror.Internal(apperror.CodeUploadFailed, "failed to upload video", err)
	}

	// Handle thumbnail (non-fatal on failure)
	thumbnailKey := s.handleThumbnail(ctx, input, videoID, tmpVideo.Name())

	// Create video entity
	video := entity.NewVideo(
		input.Title, slug, input.Description, input.Creator,
		contentRating, quality, storageKey, input.ContentType,
		input.FileSize, input.UploadedBy,
	)
	video.ID = videoID
	video.ThumbnailKey = thumbnailKey
	video.SetGenres(input.GenreIDs)

	if err := s.videoRepo.Create(ctx, video); err != nil {
		_ = s.storage.Delete(ctx, storageKey)
		if thumbnailKey != "" {
			_ = s.storage.Delete(ctx, thumbnailKey)
		}
		s.log.Error("creating video record", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to save video", err)
	}

	s.dispatcher.Dispatch(event.VideoUploaded{
		VideoID:    video.ID,
		Title:      video.Title,
		UploadedBy: input.UploadedBy,
		Timestamp:  time.Now().UTC(),
	})

	return video, nil
}

// handleThumbnail uploads a user-provided thumbnail or extracts one from the video.
// Returns the storage key on success, or empty string on failure (non-fatal).
func (s *Service) handleThumbnail(ctx context.Context, input UploadInput, videoID uuid.UUID, videoPath string) string {
	if input.Thumbnail != nil {
		return s.uploadUserThumbnail(ctx, input, videoID)
	}
	return s.extractThumbnail(ctx, videoID, videoPath)
}

func (s *Service) uploadUserThumbnail(ctx context.Context, input UploadInput, videoID uuid.UUID) string {
	thumbExt := filepath.Ext(input.ThumbnailFileName)
	if thumbExt == "" {
		thumbExt = thumbnailExtFromContentType(input.ThumbnailContentType)
	}

	key := fmt.Sprintf("videos/%s/thumbnail%s", videoID.String(), thumbExt)
	if _, err := s.storage.Upload(ctx, key, input.Thumbnail, input.ThumbnailContentType); err != nil {
		s.log.Error("uploading user thumbnail", "error", err)
		return ""
	}
	return key
}

func (s *Service) extractThumbnail(ctx context.Context, videoID uuid.UUID, videoPath string) string {
	tmpThumb, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		s.log.Error("creating temp thumbnail file", "error", err)
		return ""
	}
	tmpThumbPath := tmpThumb.Name()
	tmpThumb.Close()
	defer os.Remove(tmpThumbPath)

	if err := thumbnail.ExtractFirstFrame(videoPath, tmpThumbPath); err != nil {
		s.log.Error("extracting thumbnail with ffmpeg", "error", err)
		return ""
	}

	thumbFile, err := os.Open(tmpThumbPath)
	if err != nil {
		s.log.Error("opening extracted thumbnail", "error", err)
		return ""
	}
	defer thumbFile.Close()

	key := fmt.Sprintf("videos/%s/thumbnail.jpg", videoID.String())
	if _, err := s.storage.Upload(ctx, key, thumbFile, "image/jpeg"); err != nil {
		s.log.Error("uploading extracted thumbnail", "error", err)
		return ""
	}
	return key
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

func thumbnailExtFromContentType(ct string) string {
	switch ct {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
