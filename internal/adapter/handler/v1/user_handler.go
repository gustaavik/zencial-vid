package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
)

// UserHandler handles user profile HTTP requests.
type UserHandler struct {
	validator *validator.Validator
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler() *UserHandler {
	return &UserHandler{
		validator: validator.New(),
	}
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile
// @Tags         users
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}

// UpdateMe godoc
// @Summary      Update current user profile
// @Description  Update the authenticated user's profile fields
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body dto.UpdateProfileRequest true "Profile update data"
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}

// DeleteMe godoc
// @Summary      Delete current user account
// @Description  Soft-delete the authenticated user's account
// @Tags         users
// @Produce      json
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [delete]
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	httputil.NotFound(w, "NOT_FOUND", "not implemented")
}
