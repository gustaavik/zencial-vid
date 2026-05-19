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

// dtoLinksToEntity converts a slice of ProfileLinkRequest DTOs to entity ProfileLinks.
func dtoLinksToEntity(links []dto.ProfileLinkRequest) []entity.ProfileLink {
	out := make([]entity.ProfileLink, len(links))
	for i, l := range links {
		out[i] = entity.ProfileLink{Label: l.Label, URL: l.URL}
	}
	return out
}

// dtoPrefsToEntity converts a ProfilePreferencesRequest DTO to an entity ProfilePreferences.
func dtoPrefsToEntity(p *dto.ProfilePreferencesRequest) *entity.ProfilePreferences {
	if p == nil {
		return nil
	}
	prefs := &entity.ProfilePreferences{}
	if p.AllowMatureContent != nil {
		prefs.AllowMatureContent = *p.AllowMatureContent
	}
	if p.AutoplayNextEpisode != nil {
		prefs.AutoplayNextEpisode = *p.AutoplayNextEpisode
	}
	if p.AlwaysShowSubtitles != nil {
		prefs.AlwaysShowSubtitles = *p.AlwaysShowSubtitles
	}
	if p.ShowPaidFirstInFeed != nil {
		prefs.ShowPaidFirstInFeed = *p.ShowPaidFirstInFeed
	}
	return prefs
}

// dtoPrivacyToEntity converts a ProfilePrivacyRequest DTO to an entity ProfilePrivacy.
func dtoPrivacyToEntity(p *dto.ProfilePrivacyRequest) *entity.ProfilePrivacy {
	if p == nil {
		return nil
	}
	privacy := &entity.ProfilePrivacy{}
	if p.ProfileVisibility != nil {
		privacy.ProfileVisibility = *p.ProfileVisibility
	}
	if p.WatchHistory != nil {
		privacy.WatchHistory = *p.WatchHistory
	}
	if p.Watchlist != nil {
		privacy.Watchlist = *p.Watchlist
	}
	if p.Tipping != nil {
		privacy.Tipping = *p.Tipping
	}
	return privacy
}

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

	var linksPtr *[]entity.ProfileLink
	if req.Links != nil {
		links := dtoLinksToEntity(req.Links)
		linksPtr = &links
	}

	user, appErr := h.userService.UpdateProfile(r.Context(), useruc.UpdateProfileInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		DateOfBirth: req.DateOfBirth,
		Language:    req.Language,
		Country:     req.Country,
		Handle:      req.Handle,
		Pronouns:    req.Pronouns,
		Headline:    req.Headline,
		Bio:         req.Bio,
		Links:       linksPtr,
		Preferences: dtoPrefsToEntity(req.Preferences),
		Privacy:     dtoPrivacyToEntity(req.Privacy),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.UserToResponse(user))
}

// CheckHandle godoc
// @Summary      Check handle availability
// @Description  Returns whether a given handle is available for use
// @Tags         users
// @Produce      json
// @Param        handle query string true "Handle to check"
// @Success      200 {object} httputil.Response{data=map[string]bool}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/handle/check [get]
func (h *UserHandler) CheckHandle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	handle := r.URL.Query().Get("handle")
	if handle == "" {
		httputil.BadRequest(w, apperror.CodeBadRequest, "handle query parameter is required")
		return
	}

	out, appErr := h.userService.CheckHandle(r.Context(), useruc.CheckHandleInput{
		Handle:        handle,
		RequestingUID: userID,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, map[string]bool{"available": out.Available})
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

	if callerID, ok := middleware.GetUserID(r.Context()); ok && callerID == id {
		httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "cannot change your own status", nil))
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

// CreateUser godoc
// @Summary      Create user (admin)
// @Description  Create a new user account (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body dto.AdminCreateUserRequest true "User fields"
// @Success      201 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.AdminCreateUserRequest
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

	roles := make([]entity.UserRole, len(req.Roles))
	for i, r := range req.Roles {
		roles[i] = entity.UserRole(r)
	}

	user, appErr := h.userService.AdminCreate(r.Context(), &useruc.AdminCreateInput{
		Email:       req.Email,
		Password:    req.Password,
		Roles:       roles,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Language:    req.Language,
		Country:     req.Country,
		DateOfBirth: req.DateOfBirth,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.UserToResponse(user))
}

// UpdateUser godoc
// @Summary      Update user (admin)
// @Description  Update a user's profile, email, role, or password (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        body body dto.AdminUpdateUserRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.UserResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	var req dto.AdminUpdateUserRequest
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

	if callerID, ok := middleware.GetUserID(r.Context()); ok && callerID == id {
		hasAdmin := false
		for _, r := range req.Roles {
			if entity.UserRole(r) == entity.RoleAdmin {
				hasAdmin = true
				break
			}
		}
		if len(req.Roles) > 0 && !hasAdmin {
			httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "cannot demote yourself", nil))
			return
		}
	}

	var rolesPtr *[]entity.UserRole
	if req.Roles != nil {
		converted := make([]entity.UserRole, len(req.Roles))
		for i, r := range req.Roles {
			converted[i] = entity.UserRole(r)
		}
		rolesPtr = &converted
	}

	user, appErr := h.userService.AdminUpdate(r.Context(), &useruc.AdminUpdateInput{
		UserID:      id,
		Email:       req.Email,
		Roles:       rolesPtr,
		Password:    req.Password,
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

// DeleteUser godoc
// @Summary      Delete user (admin)
// @Description  Soft-delete a user account (admin only)
// @Tags         users
// @Param        id path string true "User ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	if callerID, ok := middleware.GetUserID(r.Context()); ok && callerID == id {
		httputil.Error(w, apperror.Forbidden(apperror.CodeForbidden, "cannot delete yourself", nil))
		return
	}

	if appErr := h.userService.AdminDelete(r.Context(), id); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
