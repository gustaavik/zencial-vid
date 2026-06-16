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
	VideoStatusFailed     VideoStatus = "failed"
)

// SubmissionStatus tracks where an episode is in the moderation pipeline.
type SubmissionStatus string

const (
	SubmissionStatusDraft       SubmissionStatus = "draft"
	SubmissionStatusSubmitted   SubmissionStatus = "submitted"
	SubmissionStatusUnderReview SubmissionStatus = "under_review"
	SubmissionStatusApproved    SubmissionStatus = "approved"
	SubmissionStatusRejected    SubmissionStatus = "rejected"
)

// GeoRestrictionType controls geographic availability.
type GeoRestrictionType string

const (
	GeoRestrictionWorldwide GeoRestrictionType = "worldwide"
	GeoRestrictionInclude   GeoRestrictionType = "include"
	GeoRestrictionExclude   GeoRestrictionType = "exclude"
)

// Video is the core video entity.
type Video struct {
	ID                  uuid.UUID
	Title               string
	Slug                valueobject.Slug
	Description         string
	Logline             string
	Creator             string
	Duration            valueobject.Duration
	ContentRating       string
	PrimaryLanguage     string
	Status              VideoStatus
	Visibility          VideoVisibility
	StorageKey          string
	ContentType         string
	FileSize            int64
	ThumbnailKey        string
	ThumbnailCandidates []string
	UploadedBy          uuid.UUID
	GenreIDs            []uuid.UUID
	MinimumPlanLevel    *int
	TranscodeError      string
	SeriesID            *uuid.UUID
	SeasonNumber        *int
	EpisodeNumber       *int
	// Scheduling
	ScheduledPublishAt *time.Time
	// Monetization
	MonetizationTypes  []string
	PPVPriceCents      *int
	FreePreviewSeconds *int
	AdBreakPositions   []int
	// Geo restrictions
	GeoRestrictionType    GeoRestrictionType
	GeoRestrictionRegions []string
	RequireSignin         bool
	// Submission / moderation
	SubmissionStatus SubmissionStatus
	SubmittedAt      *time.Time
	ModeratorNotes   string
	// Featured
	IsFeatured          bool
	FeaturedDescription string
	FeaturedAt          *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	// Views is a transient, read-only aggregate populated only by admin list
	// queries (all-time qualifying playback views). It is never persisted.
	Views int64
}

// NewVideo creates a new Video entity in draft status.
func NewVideo(
	title string,
	slug valueobject.Slug,
	description string,
	creator string,
	contentRating string,
	storageKey string,
	contentType string,
	fileSize int64,
	uploadedBy uuid.UUID,
) *Video {
	now := time.Now().UTC()
	return &Video{
		ID:                    uuid.Must(uuid.NewV7()),
		Title:                 title,
		Slug:                  slug,
		Description:           description,
		Creator:               creator,
		ContentRating:         contentRating,
		PrimaryLanguage:       "en",
		Status:                VideoStatusDraft,
		Visibility:            VideoVisibilityPublic,
		StorageKey:            storageKey,
		ContentType:           contentType,
		FileSize:              fileSize,
		UploadedBy:            uploadedBy,
		MonetizationTypes:     []string{},
		AdBreakPositions:      []int{},
		GeoRestrictionType:    GeoRestrictionWorldwide,
		GeoRestrictionRegions: []string{},
		ThumbnailCandidates:   []string{},
		SubmissionStatus:      SubmissionStatusDraft,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// Publish kicks off transcoding by moving the video into the processing state.
// The video only becomes truly published once the CDN signals completion via MarkTranscoded().
func (v *Video) Publish() {
	v.Status = VideoStatusProcessing
	v.TranscodeError = ""
	v.UpdatedAt = time.Now().UTC()
}

// MarkTranscoded transitions a processing video to published.
// No-op if the video is already published (idempotent for callback retries).
func (v *Video) MarkTranscoded() {
	v.Status = VideoStatusPublished
	v.TranscodeError = ""
	v.UpdatedAt = time.Now().UTC()
}

// MarkTranscodeFailed records a transcoding failure and sets status to failed.
func (v *Video) MarkTranscodeFailed(reason string) {
	v.Status = VideoStatusFailed
	v.TranscodeError = reason
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

// RequiresSubscription reports whether the video requires a subscription to stream.
func (v *Video) RequiresSubscription() bool {
	return v.MinimumPlanLevel != nil && *v.MinimumPlanLevel > 0
}

// SetGenres replaces the genre associations.
func (v *Video) SetGenres(genreIDs []uuid.UUID) {
	v.GenreIDs = genreIDs
	v.UpdatedAt = time.Now().UTC()
}

// Submit moves the video into the submitted state for moderation review.
func (v *Video) Submit() {
	now := time.Now().UTC()
	v.SubmissionStatus = SubmissionStatusSubmitted
	v.SubmittedAt = &now
	v.UpdatedAt = now
}

// ApproveSubmission marks the video as approved and triggers the publish flow.
func (v *Video) ApproveSubmission() {
	v.SubmissionStatus = SubmissionStatusApproved
	v.UpdatedAt = time.Now().UTC()
}

// RejectSubmission records a rejection with moderator notes.
func (v *Video) RejectSubmission(notes string) {
	v.SubmissionStatus = SubmissionStatusRejected
	v.ModeratorNotes = notes
	v.UpdatedAt = time.Now().UTC()
}

// CanBeEdited reports whether the video metadata can still be changed.
// Once submitted, edits are locked until the moderation decision is made.
func (v *Video) CanBeEdited() bool {
	return v.SubmissionStatus == SubmissionStatusDraft ||
		v.SubmissionStatus == SubmissionStatusRejected
}
