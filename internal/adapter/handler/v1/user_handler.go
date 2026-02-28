package v1

import (
	"net/http"
	"time"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	useruc "github.com/zenfulcode/zencial/internal/usecase/user"
)

// UserHandler handles user profile HTTP requests.
type UserHandler struct {
	userService *useruc.Service
	validator   *validator.Validator
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *useruc.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
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
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}

	user, appErr := h.userService.GetProfile(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
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
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}

	var req dto.UpdateProfileRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil),
			errors,
		)
		return
	}

	input := useruc.UpdateProfileInput{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Language:    req.Language,
		Country:     req.Country,
	}

	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			httputil.BadRequest(w, "VALIDATION_FAILED", "invalid date_of_birth format, expected YYYY-MM-DD")
			return
		}
		input.DateOfBirth = &dob
	}

	user, appErr := h.userService.UpdateProfile(r.Context(), userID, input)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
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
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}

	appErr := h.userService.DeleteAccount(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminListUsers godoc
// @Summary      List users (admin)
// @Tags         admin
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.UserResponse}
// @Security     BearerAuth
// @Router       /admin/users [get]
func (h *UserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)

	users, total, appErr := h.userService.ListUsers(r.Context(), page, perPage)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	meta := &httputil.Meta{
		Page: page, PerPage: perPage, Total: total,
		TotalPages: int(total) / perPage,
	}
	if int(total)%perPage > 0 {
		meta.TotalPages++
	}
	httputil.SuccessWithMeta(w, mapper.UsersToResponse(users), meta)
}

// AdminUpdateStatus godoc
// @Summary      Update user status (admin)
// @Tags         admin
// @Accept       json
// @Param        id path string true "User ID"
// @Param        body body dto.UpdateStatusRequest true "Status data"
// @Success      204
// @Security     BearerAuth
// @Router       /admin/users/{id}/status [patch]
func (h *UserHandler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid user ID")
		return
	}

	var req dto.UpdateStatusRequest
	if decodeErr := httputil.DecodeJSON(r, &req); decodeErr != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	if appErr := h.userService.UpdateUserStatus(r.Context(), id, entity.UserStatus(req.Status)); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
