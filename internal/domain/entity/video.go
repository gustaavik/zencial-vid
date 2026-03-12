package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// VideoStatus represents the lifecycle state of a video.
type VideoStatus string

const (
	VideoStatusDraft      VideoStatus = "draft"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusPublished  VideoStatus = "published"
	VideoStatusArchived   VideoStatus = "archived"
)

// Video is the core video entity.
type Video struct {
	ID            uuid.UUID
	Title         string
	Slug          valueobject.Slug
	Description   string
	Creator       string
	Duration      valueobject.Duration
	ContentRating string
	Quality       string
	Status        VideoStatus
	StorageKey    string
	ContentType   string
	FileSize      int64
	ThumbnailKey  string
	UploadedBy    uuid.UUID
	GenreIDs      []uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewVideo creates a new Video entity in draft status.
func NewVideo(
	title string,
	slug valueobject.Slug,
	description string,
	creator string,
	contentRating string,
	quality string,
	storageKey string,
	contentType string,
	fileSize int64,
	uploadedBy uuid.UUID,
) *Video {
	now := time.Now().UTC()
	return &Video{
		ID:            uuid.New(),
		Title:         title,
		Slug:          slug,
		Description:   description,
		Creator:       creator,
		ContentRating: contentRating,
		Quality:       quality,
		Status:        VideoStatusDraft,
		StorageKey:    storageKey,
		ContentType:   contentType,
		FileSize:      fileSize,
		UploadedBy:    uploadedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Publish marks the video as published.
func (v *Video) Publish() {
	v.Status = VideoStatusPublished
	v.UpdatedAt = time.Now().UTC()
}

// Archive soft-deletes the video.
func (v *Video) Archive() {
	v.Status = VideoStatusArchived
	v.UpdatedAt = time.Now().UTC()
}

// Unarchive restores an archived video back to draft status.
func (v *Video) Unarchive() {
	v.Status = VideoStatusDraft
	v.UpdatedAt = time.Now().UTC()
}

// IsPlayable reports whether the video can be streamed.
func (v *Video) IsPlayable() bool {
	return v.Status == VideoStatusPublished
}

// DeletedStorageKey returns the storage key relocated under the "deleted/" prefix.
func DeletedStorageKey(key string) string {
	if key == "" {
		return ""
	}
	return "deleted/" + key
}

// RestoredStorageKey strips the "deleted/" prefix from a storage key.
func RestoredStorageKey(key string) string {
	return strings.TrimPrefix(key, "deleted/")
}

// SetGenres replaces the genre associations.
func (v *Video) SetGenres(genreIDs []uuid.UUID) {
	v.GenreIDs = genreIDs
	v.UpdatedAt = time.Now().UTC()
}
