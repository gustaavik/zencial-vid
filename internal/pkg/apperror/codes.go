package apperror

// Error code constants used across the application.
const (
	// Auth
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeInvalidToken        = "INVALID_TOKEN"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeRefreshTokenInvalid = "REFRESH_TOKEN_INVALID"

	// User
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	CodeUserSuspended      = "USER_SUSPENDED"

	// Content
	CodeContentNotFound    = "CONTENT_NOT_FOUND"
	CodeContentNotPlayable = "CONTENT_NOT_PLAYABLE"
	CodeEpisodeNotFound    = "EPISODE_NOT_FOUND"
	CodeSeasonNotFound     = "SEASON_NOT_FOUND"

	// Catalog
	CodeGenreNotFound    = "GENRE_NOT_FOUND"
	CodeCategoryNotFound = "CATEGORY_NOT_FOUND"

	// Subscription
	CodeNoActiveSubscription = "NO_ACTIVE_SUBSCRIPTION"
	CodePlanNotFound         = "PLAN_NOT_FOUND"
	CodeAlreadySubscribed    = "ALREADY_SUBSCRIBED"

	// Streaming
	CodeMaxStreamsReached = "MAX_STREAMS_REACHED"
	CodeSessionNotFound   = "SESSION_NOT_FOUND"

	// Watchlist
	CodeAlreadyInWatchlist = "ALREADY_IN_WATCHLIST"
	CodeNotInWatchlist     = "NOT_IN_WATCHLIST"

	// General
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeNotFound         = "NOT_FOUND"
)
