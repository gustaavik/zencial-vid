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

// Auth errors.
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)
