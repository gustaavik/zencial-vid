package dto

// WatchlistStatusResponse indicates whether a video is in the authenticated user's watchlist.
type WatchlistStatusResponse struct {
	InWatchlist bool `json:"in_watchlist"`
}
