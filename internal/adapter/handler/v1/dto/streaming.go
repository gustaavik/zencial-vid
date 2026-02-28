package dto

// StartSessionRequest represents a stream session start request.
type StartSessionRequest struct {
	ContentID  string  `json:"content_id" validate:"required,uuid"`
	EpisodeID  *string `json:"episode_id,omitempty" validate:"omitempty,uuid"`
	DeviceInfo string  `json:"device_info" validate:"required"`
}

// SessionResponse represents a stream session in API responses.
type SessionResponse struct {
	ID          string `json:"id"`
	ManifestURL string `json:"manifest_url"`
	ExpiresAt   string `json:"expires_at"`
}

// UpdateProgressRequest represents a playback progress update.
type UpdateProgressRequest struct {
	ContentID string  `json:"content_id" validate:"required,uuid"`
	EpisodeID *string `json:"episode_id,omitempty" validate:"omitempty,uuid"`
	Position  int64   `json:"position" validate:"required,min=0"`
	Duration  int64   `json:"duration" validate:"required,min=1"`
}

// ProgressResponse represents playback progress in API responses.
type ProgressResponse struct {
	ContentID  string  `json:"content_id"`
	EpisodeID  *string `json:"episode_id,omitempty"`
	Position   int64   `json:"position"`
	Duration   int64   `json:"duration"`
	Percentage float64 `json:"percentage"`
	Completed  bool    `json:"completed"`
}

// ContinueWatchingResponse represents a continue watching item.
type ContinueWatchingResponse struct {
	Content  ContentListResponse `json:"content"`
	Episode  *EpisodeResponse    `json:"episode,omitempty"`
	Progress ProgressResponse    `json:"progress"`
}
