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
	ErrSlugAlreadyExists = errors.New("slug already exists")
)

// Video errors.
var (
	ErrVideoNotFound    = errors.New("video not found")
	ErrVideoNotPlayable = errors.New("video is not playable")
)

// Genre errors.
var (
	ErrGenreNotFound = errors.New("genre not found")
)

// Plan errors.
var (
	ErrPlanNotFound   = errors.New("plan not found")
	ErrPlanSlugExists = errors.New("plan slug already exists")
)

// Subscription errors.
var (
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrActiveSubscriptionExists = errors.New("user already has an active subscription")
	ErrInsufficientPlanLevel    = errors.New("insufficient plan level for this content")
)

// Auth errors.
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

// Watchlist errors.
var (
	ErrWatchlistEntryNotFound = errors.New("watchlist entry not found")
)

// Watch progress errors.
var (
	ErrWatchProgressNotFound = errors.New("watch progress not found")
)
