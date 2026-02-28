package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// Response is the standard API response envelope.
type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody holds error details.
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Success writes a success response with data.
func Success(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, Response{Data: data})
}

// SuccessWithMeta writes a success response with data and pagination metadata.
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	JSON(w, http.StatusOK, Response{Data: data, Meta: meta})
}

// Error writes an error response from an AppError.
func Error(w http.ResponseWriter, err *apperror.AppError) {
	JSON(w, err.HTTPStatus, ErrorResponse{
		Error: ErrorBody{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}

// ErrorWithDetails writes an error response with additional details.
func ErrorWithDetails(w http.ResponseWriter, err *apperror.AppError, details interface{}) {
	JSON(w, err.HTTPStatus, ErrorResponse{
		Error: ErrorBody{
			Code:    err.Code,
			Message: err.Message,
			Details: details,
		},
	})
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, code, message string) {
	Error(w, apperror.BadRequest(code, message, nil))
}

// NotFound writes a 404 error response.
func NotFound(w http.ResponseWriter, code, message string) {
	Error(w, apperror.NotFound(code, message, nil))
}

// Unauthorized writes a 401 error response.
func Unauthorized(w http.ResponseWriter, code, message string) {
	Error(w, apperror.Unauthorized(code, message, nil))
}

// InternalError writes a 500 error response.
func InternalError(w http.ResponseWriter) {
	Error(w, apperror.Internal(apperror.CodeInternalError, "an internal error occurred", nil))
}
