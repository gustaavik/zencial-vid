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

	// Videos
	CodeSlugConflict = "SLUG_CONFLICT"

	// Catalog
	CodeGenreNotFound = "GENRE_NOT_FOUND"

	// Video
	CodeVideoNotFound    = "VIDEO_NOT_FOUND"
	CodeVideoNotPlayable = "VIDEO_NOT_PLAYABLE"
	CodeStorageError     = "STORAGE_ERROR"
	CodeUploadFailed     = "UPLOAD_FAILED"

	// General
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeNotFound         = "NOT_FOUND"
)
