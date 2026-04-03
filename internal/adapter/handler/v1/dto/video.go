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
	CreatedAt        string   `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt        string   `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// VideoStreamResponse returns the streaming URL (presigned or HLS).
type VideoStreamResponse struct {
	URL       string `json:"url" example:"https://s3.example.com/bucket/videos/..."`
	ExpiresAt string `json:"expires_at" example:"2025-01-01T04:00:00Z"`
	Type      string `json:"type" example:"hls"`
}

// UpdateVideoRequest represents a video metadata update.
type UpdateVideoRequest struct {
	Title            *string  `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description      *string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	Creator          *string  `json:"creator,omitempty" validate:"omitempty,max=255"`
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
