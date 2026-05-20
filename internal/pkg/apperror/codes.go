package apperror

// Error code constants used across the application.
const (
	// Auth
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeSessionNotFound    = "SESSION_NOT_FOUND"
	CodeSessionRevoked     = "SESSION_REVOKED"
	CodeSessionExpired     = "SESSION_EXPIRED"

	// User
	CodeUserNotFound        = "USER_NOT_FOUND"
	CodeEmailAlreadyExists  = "EMAIL_ALREADY_EXISTS"
	CodeHandleAlreadyExists = "HANDLE_ALREADY_EXISTS"
	CodeUserSuspended       = "USER_SUSPENDED"
	CodeUserDeleted         = "USER_DELETED"
	CodeInvalidDateFormat   = "INVALID_DATE_FORMAT"

	// Videos
	CodeSlugConflict = "SLUG_CONFLICT"

	// Catalog
	CodeGenreNotFound = "GENRE_NOT_FOUND"

	// Video
	CodeVideoNotFound       = "VIDEO_NOT_FOUND"
	CodeVideoNotPlayable    = "VIDEO_NOT_PLAYABLE"
	CodeVideoNotTranscoding = "VIDEO_NOT_TRANSCODING"
	CodeStorageError        = "STORAGE_ERROR"
	CodeUploadFailed        = "UPLOAD_FAILED"

	// Plan
	CodePlanNotFound     = "PLAN_NOT_FOUND"
	CodePlanSlugConflict = "PLAN_SLUG_CONFLICT"

	// Subscription
	CodeSubscriptionNotFound     = "SUBSCRIPTION_NOT_FOUND"
	CodeActiveSubscriptionExists = "ACTIVE_SUBSCRIPTION_EXISTS"
	CodeInsufficientPlanLevel    = "INSUFFICIENT_PLAN_LEVEL"

	// Billing
	CodeBillingNotConfigured = "BILLING_NOT_CONFIGURED"
	CodeBillingFailed        = "BILLING_FAILED"

	// Watchlist
	CodeWatchlistEntryNotFound = "WATCHLIST_ENTRY_NOT_FOUND"

	// Watch progress
	CodeWatchProgressNotFound = "WATCH_PROGRESS_NOT_FOUND"

	// Cast
	CodeCastNotFound = "CAST_NOT_FOUND"

	// Publisher / ownership
	CodeVideoOwnershipRequired  = "VIDEO_OWNERSHIP_REQUIRED"
	CodeSeriesOwnershipRequired = "SERIES_OWNERSHIP_REQUIRED"

	// Analytics
	CodeAnalyticsNotFound = "ANALYTICS_NOT_FOUND"

	// Series
	CodeSeriesNotFound              = "SERIES_NOT_FOUND"
	CodeSeriesSlugConflict          = "SERIES_SLUG_CONFLICT"
	CodeEpisodeAlreadyExists        = "EPISODE_ALREADY_EXISTS"
	CodeSeriesWatchProgressNotFound = "SERIES_WATCH_PROGRESS_NOT_FOUND"

	// General
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeNotFound         = "NOT_FOUND"
)
