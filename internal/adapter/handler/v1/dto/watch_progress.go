package dto

// UpsertWatchProgressRequest is the body for PUT /me/watch-progress/{video_id}.
type UpsertWatchProgressRequest struct {
	PositionSeconds int64 `json:"position_seconds" example:"120" validate:"gte=0"`
}

// WatchProgressResponse represents the user's saved progress for a single video.
type WatchProgressResponse struct {
	VideoID         string  `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PositionSeconds int64   `json:"position_seconds" example:"120"`
	DurationSeconds int64   `json:"duration_seconds" example:"600"`
	PercentWatched  float64 `json:"percent_watched" example:"20.0"`
	UpdatedAt       string  `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// ContinueWatchingItem couples a video with the user's playback progress for the
// "continue watching" feed.
type ContinueWatchingItem struct {
	Video           VideoResponse `json:"video"`
	PositionSeconds int64         `json:"position_seconds" example:"120"`
	DurationSeconds int64         `json:"duration_seconds" example:"600"`
	PercentWatched  float64       `json:"percent_watched" example:"20.0"`
	UpdatedAt       string        `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}
