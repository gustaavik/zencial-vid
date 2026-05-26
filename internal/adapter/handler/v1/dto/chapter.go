package dto

// ChapterResponse represents a video chapter in API responses.
type ChapterResponse struct {
	ID            string `json:"id"`
	VideoID       string `json:"video_id"`
	StartTimeSecs int    `json:"start_time_secs" example:"252"`
	Title         string `json:"title" example:"Cold open"`
	Source        string `json:"source" example:"auto"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ChapterItem is a single entry in a ReplaceChaptersRequest.
type ChapterItem struct {
	StartTimeSecs int    `json:"start_time_secs" validate:"gte=0"`
	Title         string `json:"title" validate:"required,min=1,max=500"`
	Source        string `json:"source,omitempty" validate:"omitempty,oneof=auto manual"`
}

// ReplaceChaptersRequest replaces all chapters for a video atomically.
type ReplaceChaptersRequest struct {
	Chapters []ChapterItem `json:"chapters" validate:"required"`
}
