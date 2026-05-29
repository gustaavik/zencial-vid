package dto

// VideoStatsResponse represents viewing statistics for a single video.
type VideoStatsResponse struct {
	VideoID        string  `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TotalViewers   int64   `json:"total_viewers" example:"1234"`
	AvgProgressPct float64 `json:"avg_progress_pct" example:"72.5"`
	CompletionRate float64 `json:"completion_rate" example:"45.2"`
}
