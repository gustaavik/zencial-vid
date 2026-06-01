package dto

// CastCreditResponse represents a cast member's credit on a specific video.
type CastCreditResponse struct {
	ID           string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CastID       string  `json:"cast_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	VideoID      string  `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string  `json:"name" example:"Jane Doe"`
	Role         string  `json:"role" example:"actor"`
	Department   string  `json:"department" example:"performance"`
	InviteStatus string  `json:"invite_status" example:"accepted"`
	InvitedEmail *string `json:"invited_email,omitempty"`
	SortOrder    int     `json:"sort_order" example:"0"`
	PictureURL   string  `json:"picture_url,omitempty" example:"https://cdn.example.com/cast/picture.jpg"`
	CreatedAt    string  `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt    string  `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CastMemberResponse represents a standalone cast member.
type CastMemberResponse struct {
	ID         string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name       string `json:"name" example:"Jane Doe"`
	Status     string `json:"status" example:"active"`
	PictureURL string `json:"picture_url,omitempty" example:"https://cdn.example.com/cast/picture.jpg"`
	CreatedAt  string `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CastVideoResponse represents a video a cast member appears in, including
// the credit role and sort position within that video.
type CastVideoResponse struct {
	VideoID       string  `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title         string  `json:"title" example:"My Awesome Video"`
	Slug          string  `json:"slug" example:"my-awesome-video-a3f8b2c1"`
	Role          string  `json:"role" example:"actor"`
	SortOrder     int     `json:"sort_order" example:"0"`
	Status        string  `json:"status" example:"published"`
	ThumbnailURL  string  `json:"thumbnail_url,omitempty" example:"https://cdn.example.com/thumbnails/video.jpg"`
	Duration      int64   `json:"duration" example:"3600"`
	ContentRating string  `json:"content_rating" example:"PG"`
	SeriesID      *string `json:"series_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeasonNumber  *int    `json:"season_number,omitempty" example:"1"`
	EpisodeNumber *int    `json:"episode_number,omitempty" example:"3"`
	CreatedAt     string  `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt     string  `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CreateCastRequest is the body for adding a cast member to a video.
type CreateCastRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=255" example:"Jane Doe"`
	Role       string `json:"role" validate:"required,min=1,max=100" example:"actor"`
	Department string `json:"department,omitempty" validate:"omitempty,oneof=performance direction cinematography sound post production writing vfx"`
	SortOrder  int    `json:"sort_order" validate:"gte=0" example:"0"`
}

// UpdateCastRequest is the body for updating a cast member's name globally.
type UpdateCastRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=255" example:"Jane Doe"`
}

// UpdateCreditRequest is the body for updating a cast credit's role or sort order.
type UpdateCreditRequest struct {
	Role       *string `json:"role,omitempty" validate:"omitempty,min=1,max=100" example:"director"`
	Department *string `json:"department,omitempty" validate:"omitempty,oneof=performance direction cinematography sound post production writing vfx"`
	SortOrder  *int    `json:"sort_order,omitempty" validate:"omitempty,gte=0" example:"1"`
}
