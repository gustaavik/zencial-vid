package dto

// AnalyticsTotals holds aggregate viewing statistics for one reporting window.
type AnalyticsTotals struct {
	Views             int64   `json:"views" example:"1382"`
	WatchTimeMinutes  int64   `json:"watch_time_minutes" example:"51234"`
	UniqueViewers     int64   `json:"unique_viewers" example:"411"`
	AvgPercentWatched float64 `json:"avg_percent_watched" example:"46.1"`
	FinishRate        float64 `json:"finish_rate" example:"31.0"`
}

// AnalyticsDeltas holds period-over-period changes vs the previous
// equal-length window. Count metrics are relative percent changes; rate
// metrics are percentage-point differences.
type AnalyticsDeltas struct {
	ViewsPct             float64 `json:"views_pct" example:"22.0"`
	WatchTimePct         float64 `json:"watch_time_pct" example:"18.2"`
	UniqueViewersPct     float64 `json:"unique_viewers_pct" example:"10.4"`
	AvgPercentWatchedPts float64 `json:"avg_percent_watched_pts" example:"-4.1"`
	FinishRatePts        float64 `json:"finish_rate_pts" example:"3.0"`
}

// AnalyticsDailyPoint is one day of viewing activity.
type AnalyticsDailyPoint struct {
	Date             string `json:"date" example:"2026-05-13"`
	Views            int64  `json:"views" example:"42"`
	WatchTimeMinutes int64  `json:"watch_time_minutes" example:"1630"`
}

// TopVideoItem is one row of the dashboard "top videos" table.
type TopVideoItem struct {
	VideoID           string  `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title             string  `json:"title" example:"Saturnine — S1·E1"`
	Status            string  `json:"status" example:"published"`
	Views             int64   `json:"views" example:"412"`
	WatchTimeMinutes  int64   `json:"watch_time_minutes" example:"9000"`
	AvgPercentWatched float64 `json:"avg_percent_watched" example:"70.1"`
	FinishRate        float64 `json:"finish_rate" example:"64.2"`
}

// BreakdownItem is one slice of a categorical breakdown.
type BreakdownItem struct {
	Key   string  `json:"key" example:"home"`
	Views int64   `json:"views" example:"412"`
	Pct   float64 `json:"pct" example:"32.1"`
}

// PublisherSummaryResponse is the dashboard summary report.
type PublisherSummaryResponse struct {
	Range      string                `json:"range" example:"30d"`
	StartDate  string                `json:"start_date" example:"2026-05-13"`
	EndDate    string                `json:"end_date" example:"2026-06-12"`
	Totals     AnalyticsTotals       `json:"totals"`
	Deltas     *AnalyticsDeltas      `json:"deltas,omitempty"`
	Timeseries []AnalyticsDailyPoint `json:"timeseries"`
	TopVideos  []TopVideoItem        `json:"top_videos"`
}

// VideoAnalyticsResponse is the full per-video analytics report.
type VideoAnalyticsResponse struct {
	VideoID    string                `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Range      string                `json:"range" example:"30d"`
	StartDate  string                `json:"start_date" example:"2026-05-13"`
	EndDate    string                `json:"end_date" example:"2026-06-12"`
	Totals     AnalyticsTotals       `json:"totals"`
	Deltas     *AnalyticsDeltas      `json:"deltas,omitempty"`
	Timeseries []AnalyticsDailyPoint `json:"timeseries"`
	// Retention has 100 entries; entry i is the percentage of views that
	// watched the i-th percent of the video.
	Retention []float64       `json:"retention"`
	Sources   []BreakdownItem `json:"sources"`
	Countries []BreakdownItem `json:"countries"`
	Platforms []BreakdownItem `json:"platforms"`
}

// RecordPlaybackRequest is one cumulative playback heartbeat. All numeric
// fields and watched_buckets are cumulative for the session identified by
// session_id, so resending a heartbeat is harmless.
type RecordPlaybackRequest struct {
	SessionID       string `json:"session_id" validate:"required,uuid" example:"01975c8e-1c3a-7000-8000-000000000000"`
	Source          string `json:"source" example:"home"`
	Platform        string `json:"platform" example:"ios"`
	PositionSeconds int64  `json:"position_seconds" validate:"gte=0" example:"184"`
	WatchedSeconds  int64  `json:"watched_seconds" validate:"gte=0" example:"122"`
	WatchedBuckets  []int  `json:"watched_buckets" validate:"omitempty,max=100,dive,gte=0,lte=99"`
	Completed       bool   `json:"completed" example:"false"`
}
