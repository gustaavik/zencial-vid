package dto

// WatchlistItemResponse represents a watchlist item in API responses.
type WatchlistItemResponse struct {
	ContentID string              `json:"content_id"`
	Content   ContentListResponse `json:"content"`
	AddedAt   string              `json:"added_at"`
}

// WatchlistStatusResponse represents the watchlist status of a content item.
type WatchlistStatusResponse struct {
	InWatchlist bool `json:"in_watchlist"`
}
