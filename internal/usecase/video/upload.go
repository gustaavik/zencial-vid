package video

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/domain/event"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// PresignedUploadExpiry is the lifetime of presigned PUT URLs returned by InitiateUpload.
const PresignedUploadExpiry = 30 * time.Minute

// InitiateUploadInput holds the metadata needed to mint a presigned PUT URL.
type InitiateUploadInput struct {
	FileName    string
	ContentType string
}

// InitiateUploadOutput is returned to the client so it can PUT the binary directly to storage.
type InitiateUploadOutput struct {
	UploadURL string
	ObjectKey string
	ExpiresAt time.Time
}

// InitiateUpload generates a presigned PUT URL the client uses to upload the binary
// directly to object storage, bypassing this API for the bulk transfer.
func (s *Service) InitiateUpload(ctx context.Context, input *InitiateUploadInput) (*InitiateUploadOutput, *apperror.AppError) {
	contentType := input.ContentType
	if contentType == "" {
		contentType = "video/mp4"
	}

	videoID := uuid.New()
	ext := extensionFromContentType(contentType)
	if ext == "" {
		ext = filepath.Ext(input.FileName)
	}
	if ext == "" {
		ext = ".mp4"
	}
	objectKey := fmt.Sprintf("videos/%s/original%s", videoID.String(), ext)

	url, err := s.storage.PresignedPutURL(ctx, objectKey, contentType, PresignedUploadExpiry)
	if err != nil {
		s.log.Error("generating presigned upload URL", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to generate upload URL", err)
	}

	return &InitiateUploadOutput{
		UploadURL: url,
		ObjectKey: objectKey,
		ExpiresAt: time.Now().UTC().Add(PresignedUploadExpiry),
	}, nil
}

// CompleteUploadInput holds the metadata needed to finalize a previously-initiated upload.
type CompleteUploadInput struct {
	ObjectKey        string
	Title            string
	Description      string
	Creator          string
	ContentRating    string
	GenreIDs         []uuid.UUID
	UploadedBy       uuid.UUID
	MinimumPlanLevel *int
	// DurationSeconds is an optional client-supplied hint. The async transcode
	// pipeline can override it later with the authoritative value.
	DurationSeconds int64
}

// CompleteUpload verifies that the binary has landed at ObjectKey and creates
// the video metadata record. The objectKey must come from a prior InitiateUpload
// response; the videoID is derived from the key path.
func (s *Service) CompleteUpload(ctx context.Context, input *CompleteUploadInput) (*entity.Video, *apperror.AppError) {
	videoID, err := videoIDFromObjectKey(input.ObjectKey)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid object_key", err)
	}

	info, statErr := s.storage.Stat(ctx, input.ObjectKey)
	if statErr != nil {
		s.log.Error("stat uploaded object", "error", statErr, "key", input.ObjectKey)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to verify upload", statErr)
	}
	if info == nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "no object at object_key — upload not completed", nil)
	}

	contentRating := input.ContentRating
	if contentRating == "" {
		contentRating = "G"
	}

	slug, err := valueobject.NewSlug(input.Title)
	if err != nil {
		return nil, apperror.BadRequest(apperror.CodeValidationFailed, "invalid title for slug generation", err)
	}
	slug = slug.WithRandomID()

	contentType := info.ContentType
	if contentType == "" {
		contentType = "video/mp4"
	}

	video := entity.NewVideo(
		input.Title, slug, input.Description, input.Creator,
		contentRating, input.ObjectKey, contentType,
		info.Size, input.UploadedBy,
	)
	video.ID = videoID
	video.Duration = valueobject.NewDuration(input.DurationSeconds)
	video.MinimumPlanLevel = input.MinimumPlanLevel
	video.SetGenres(input.GenreIDs)

	if err := s.videoRepo.Create(ctx, video); err != nil {
		// Best-effort cleanup of the orphaned object — without a DB row, the
		// transcode pipeline will never pick it up and it would leak.
		_ = s.storage.Delete(ctx, input.ObjectKey)
		s.log.Error("creating video record", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to save video", err)
	}

	if err := s.dispatcher.Dispatch(event.VideoUploaded{
		VideoID:    video.ID,
		Title:      video.Title,
		UploadedBy: input.UploadedBy,
		Timestamp:  time.Now().UTC(),
	}); err != nil {
		s.log.Error("dispatching video uploaded event", "error", err)
	}

	return video, nil
}

// videoIDFromObjectKey extracts the video UUID from a key of the form
// "videos/<uuid>/original<ext>".
func videoIDFromObjectKey(key string) (uuid.UUID, error) {
	parts := strings.Split(key, "/")
	if len(parts) < 3 || parts[0] != "videos" {
		return uuid.Nil, fmt.Errorf("expected key prefix videos/<uuid>/, got %q", key)
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing video id from key: %w", err)
	}
	return id, nil
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
