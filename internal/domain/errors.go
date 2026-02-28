package domain

import "errors"

// User errors.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrUserDeleted        = errors.New("user account is deleted")
)

// Content errors.
var (
	ErrContentNotFound    = errors.New("content not found")
	ErrContentNotPlayable = errors.New("content is not playable")
	ErrEpisodeNotFound    = errors.New("episode not found")
	ErrSeasonNotFound     = errors.New("season not found")
)

// Catalog errors.
var (
	ErrGenreNotFound    = errors.New("genre not found")
	ErrCategoryNotFound = errors.New("category not found")
)

// Subscription errors.
var (
	ErrNoActiveSubscription = errors.New("no active subscription")
	ErrPlanNotFound         = errors.New("plan not found")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription")
)

// Streaming errors.
var (
	ErrMaxStreamsReached = errors.New("maximum concurrent streams reached")
	ErrSessionNotFound   = errors.New("stream session not found")
)

// Watchlist errors.
var (
	ErrAlreadyInWatchlist = errors.New("content already in watchlist")
	ErrNotInWatchlist     = errors.New("content not in watchlist")
)

// Auth errors.
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)
