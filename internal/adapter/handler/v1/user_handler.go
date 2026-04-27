package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
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
// @Summary      Get my profile
// @Description  Return the authenticated user's profile
// @Tags         users
// @Produce      json
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
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
// @Summary      Update my profile
// @Description  Update the authenticated user's profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body dto.UpdateProfileRequest true "Profile fields to update"
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me [put]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	var req dto.UpdateProfileRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errors,
		)
		return
	}

	user, appErr := h.userService.UpdateProfile(r.Context(), useruc.UpdateProfileInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		DateOfBirth: req.DateOfBirth,
		Language:    req.Language,
		Country:     req.Country,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
}

// DeleteMe godoc
// @Summary      Delete my account
// @Description  Soft-delete the authenticated user's account
// @Tags         users
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me [delete]
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	appErr := h.userService.DeleteAccount(r.Context(), userID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers godoc
// @Summary      List users
// @Description  Return a paginated list of users (admin only)
// @Tags         users
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. -created_at)"
// @Success      200 {object} httputil.Response{data=[]dto.UserResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.UserFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	users, total, appErr := h.userService.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.UsersToResponse(users), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Return a single user by their UUID (admin only)
// @Tags         users
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	user, appErr := h.userService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
}

// UpdateUserStatus godoc
// @Summary      Update user status
// @Description  Activate or suspend a user (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        body body dto.UpdateStatusRequest true "New status"
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	var req dto.UpdateStatusRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errors,
		)
		return
	}

	user, appErr := h.userService.UpdateStatus(r.Context(), useruc.UpdateStatusInput{
		UserID: id,
		Status: entity.UserStatus(req.Status),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
}
