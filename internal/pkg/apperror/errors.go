package apperror

import (
	"fmt"
	"net/http"
)

// AppError represents an application-level error with an HTTP status code.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NotFound creates a 404 error.
func NotFound(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
		Err:        err,
	}
}

// BadRequest creates a 400 error.
func BadRequest(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// Unauthorized creates a 401 error.
func Unauthorized(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Err:        err,
	}
}

// Forbidden creates a 403 error.
func Forbidden(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Err:        err,
	}
}

// Conflict creates a 409 error.
func Conflict(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusConflict,
		Err:        err,
	}
}

// Internal creates a 500 error.
func Internal(code, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}
