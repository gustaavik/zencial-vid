package dto

// VideoResponse represents a video in API responses.
type VideoResponse struct {
	ID               string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title            string   `json:"title" example:"My Awesome Video"`
	Slug             string   `json:"slug" example:"my-awesome-video-a3f8b2c1"`
	Description      string   `json:"description" example:"A great video about..."`
	Creator          string   `json:"creator" example:"John Doe"`
	Duration         int64    `json:"duration" example:"3600"`
	ContentRating    string   `json:"content_rating" example:"PG"`
	Status           string   `json:"status" example:"published"`
	ThumbnailURL     string   `json:"thumbnail_url,omitempty" example:"https://..."`
	FileSize         int64    `json:"file_size" example:"104857600"`
	GenreIDs         []string `json:"genre_ids"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" example:"1"`
	IsAccessible     *bool    `json:"is_accessible,omitempty"`
	TranscodeError   string   `json:"transcode_error,omitempty" example:"ffmpeg exited with status 1"`
	SeriesID         *string  `json:"series_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeasonNumber     *int     `json:"season_number,omitempty" example:"1"`
	EpisodeNumber    *int     `json:"episode_number,omitempty" example:"3"`
	CreatedAt        string   `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt        string   `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// VideoStreamResponse returns the streaming URL (presigned or HLS).
type VideoStreamResponse struct {
	URL       string `json:"url" example:"https://s3.example.com/bucket/videos/..."`
	ExpiresAt string `json:"expires_at" example:"2025-01-01T04:00:00Z"`
	Type      string `json:"type" example:"hls"`
}

// InitiateUploadRequest is the body for POST /videos/uploads.
type InitiateUploadRequest struct {
	FileName    string `json:"filename" validate:"required,min=1,max=512" example:"my-video.mp4"`
	ContentType string `json:"content_type" validate:"required" example:"video/mp4"`
}

// InitiateUploadResponse describes how to upload the binary directly to storage.
type InitiateUploadResponse struct {
	UploadURL string `json:"upload_url" example:"https://pc-s3.zencial.net/zencial-videos/videos/.../original.mp4?X-Amz-Signature=..."`
	ObjectKey string `json:"object_key" example:"videos/550e8400-e29b-41d4-a716-446655440000/original.mp4"`
	ExpiresAt string `json:"expires_at" example:"2026-05-06T13:30:00Z"`
}

// CompleteUploadRequest is the body for POST /videos (commit).
type CompleteUploadRequest struct {
	ObjectKey        string   `json:"object_key" validate:"required" example:"videos/550e8400-e29b-41d4-a716-446655440000/original.mp4"`
	Title            string   `json:"title" validate:"required,min=1,max=500" example:"My Awesome Video"`
	Description      string   `json:"description,omitempty" validate:"omitempty,max=5000"`
	Creator          string   `json:"creator,omitempty" validate:"omitempty,min=3,max=24"`
	ContentRating    string   `json:"content_rating,omitempty" validate:"omitempty,oneof=G PG PG13 R NC17"`
	GenreIDs         []string `json:"genre_ids,omitempty" validate:"omitempty,dive,uuid"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" validate:"omitempty,gte=0"`
	DurationSeconds  int64    `json:"duration_seconds,omitempty" validate:"omitempty,gte=0" example:"3600"`
}

// UpdateVideoRequest represents a video metadata update.
type UpdateVideoRequest struct {
	Title            *string  `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description      *string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	Creator          *string  `json:"creator,omitempty" validate:"omitempty,min=3,max=24"`
	ContentRating    *string  `json:"content_rating,omitempty" validate:"omitempty,oneof=G PG PG13 R NC17"`
	GenreIDs         []string `json:"genre_ids,omitempty" validate:"omitempty,dive,uuid"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" validate:"omitempty,gte=0"`
}

// BulkVideoIDsRequest represents a bulk action request with a list of video IDs.
type BulkVideoIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid"`
}

// BulkFailureResponse represents a single failure in a bulk operation.
type BulkFailureResponse struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BulkResultResponse represents the result of a bulk operation.
type BulkResultResponse struct {
	Succeeded []string              `json:"succeeded"`
	Failed    []BulkFailureResponse `json:"failed"`
}

// PurgeOrphansRequest is the optional body for POST /admin/videos/purge-orphans.
type PurgeOrphansRequest struct {
	// IncludeS3Orphans also scans S3 and deletes objects not referenced by any DB row.
	IncludeS3Orphans bool `json:"include_s3_orphans"`
	// DryRun reports what would be deleted without committing any changes.
	DryRun bool `json:"dry_run"`
}

// PurgeOrphansResponse reports which rows/objects were (or would be) removed.
type PurgeOrphansResponse struct {
	DryRun    bool     `json:"dry_run"`
	DBOrphans []string `json:"db_orphans"`
	S3Orphans []string `json:"s3_orphans"`
}
