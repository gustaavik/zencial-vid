package v1

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	castuc "github.com/zenfulcode/zencial/internal/usecase/cast"
)

// CastHandler handles cast HTTP requests.
type CastHandler struct {
	castService *castuc.Service
	cdnURLs     mapper.ThumbnailURLBuilder
	validator   *validator.Validator
}

// NewCastHandler creates a new CastHandler.
func NewCastHandler(castService *castuc.Service, cdnURLs mapper.ThumbnailURLBuilder) *CastHandler {
	return &CastHandler{
		castService: castService,
		cdnURLs:     cdnURLs,
		validator:   validator.New(),
	}
}

// ListAll godoc
// @Summary      List all cast members
// @Description  Returns a paginated list of all cast members ordered by name. Publisher or admin access required.
// @Tags         cast
// @Produce      json
// @Param        page             query int    false "Page number" default(1)
// @Param        per_page         query int    false "Items per page" default(20)
// @Param        include_archived query bool   false "Include archived cast members (admin only)"
// @Success      200 {object} httputil.Response{data=[]dto.CastMemberResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/cast [get]
func (h *CastHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	includeArchived := r.URL.Query().Get("include_archived") == "true"

	out, appErr := h.castService.ListAll(r.Context(), castuc.ListAllInput{
		Page:            page,
		PerPage:         perPage,
		IncludeArchived: includeArchived,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.CastMembersToResponse(out.Members), &httputil.Meta{
		Page:    out.Page,
		PerPage: out.PerPage,
		Total:   int64(out.Total),
	})
}

// List godoc
// @Summary      List cast
// @Description  Returns all cast credits for a video, ordered by sort_order
// @Tags         cast
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=[]dto.CastCreditResponse}
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /videos/{id}/cast [get]
func (h *CastHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	credits, appErr := h.castService.List(r.Context(), videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoCastListToResponse(credits))
}

// Create godoc
// @Summary      Add cast member to video
// @Description  Adds a cast member to a video using find-or-create by name. Publishers may only add cast to their own videos.
// @Tags         cast
// @Accept       json
// @Produce      json
// @Param        id   path string               true "Video ID"
// @Param        body body dto.CreateCastRequest true "Cast data"
// @Success      201 {object} httputil.Response{data=dto.CastCreditResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/cast [post]
func (h *CastHandler) Create(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.CreateCastRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil), errs)
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	vc, appErr := h.castService.Create(r.Context(), &castuc.CreateInput{
		VideoID:     videoID,
		Name:        req.Name,
		Role:        req.Role,
		Department:  entity.CastDepartment(req.Department),
		SortOrder:   req.SortOrder,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.VideoCastToResponse(vc))
}

// UpdateCast godoc
// @Summary      Update cast member name
// @Description  Updates a cast member's name globally. Admins are unrestricted; publishers must have the member credited on one of their videos.
// @Tags         cast
// @Accept       json
// @Produce      json
// @Param        id   path string               true "Cast member ID"
// @Param        body body dto.UpdateCastRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.CastMemberResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /cast/{id} [put]
func (h *CastHandler) UpdateCast(w http.ResponseWriter, r *http.Request) {
	castID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cast ID")
		return
	}

	var req dto.UpdateCastRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil), errs)
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	c, appErr := h.castService.UpdateCast(r.Context(), &castuc.UpdateCastInput{
		ID:          castID,
		Name:        req.Name,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.CastToMemberResponse(c))
}

// UpdateCredit godoc
// @Summary      Update cast credit
// @Description  Updates the role or sort_order for a specific cast credit. Publishers may only update credits for their own videos.
// @Tags         cast
// @Accept       json
// @Produce      json
// @Param        videoID  path string                 true "Video ID"
// @Param        creditID path string                 true "Cast credit ID"
// @Param        body     body dto.UpdateCreditRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.CastCreditResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{videoID}/cast/{creditID} [put]
func (h *CastHandler) UpdateCredit(w http.ResponseWriter, r *http.Request) {
	creditID, err := uuid.Parse(chi.URLParam(r, "creditID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid credit ID")
		return
	}

	var req dto.UpdateCreditRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil), errs)
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	updateInput := &castuc.UpdateCreditInput{
		CreditID:    creditID,
		Role:        req.Role,
		SortOrder:   req.SortOrder,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	}
	if req.Department != nil {
		dept := entity.CastDepartment(*req.Department)
		updateInput.Department = &dept
	}
	vc, appErr := h.castService.UpdateCredit(r.Context(), updateInput)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoCastToResponse(vc))
}

// UploadPicture godoc
// @Summary      Upload cast member picture
// @Description  Uploads or replaces a cast member's picture via multipart form.
// @Tags         cast
// @Accept       multipart/form-data
// @Produce      json
// @Param        id      path     string true "Cast member ID" format(uuid)
// @Param        picture formData file   true "Picture image (JPEG, PNG, WebP, or GIF)"
// @Success      200 {object} httputil.Response{data=dto.CastMemberResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /cast/{id}/picture [put]
func (h *CastHandler) UploadPicture(w http.ResponseWriter, r *http.Request) {
	castID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cast ID")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("picture")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "picture file is required")
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	extByContentType := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	ext, ok := extByContentType[contentType]
	if !ok {
		httputil.BadRequest(w, apperror.CodeValidationFailed, "unsupported image format; use JPEG, PNG, WebP, or GIF")
		return
	}

	callerID, ok2 := middleware.GetUserID(r.Context())
	if !ok2 {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	out, appErr := h.castService.UploadPicture(r.Context(), &castuc.UploadPictureInput{
		ID:          castID,
		Body:        file,
		ContentType: contentType,
		Ext:         ext,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.CastToMemberResponse(out.Cast))
}

// DeleteFromVideo godoc
// @Summary      Remove cast credit from video
// @Description  Removes a specific cast credit from a video. Publishers may only remove credits from their own videos.
// @Tags         cast
// @Produce      json
// @Param        videoID  path string true "Video ID"
// @Param        creditID path string true "Cast credit ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{videoID}/cast/{creditID} [delete]
func (h *CastHandler) DeleteFromVideo(w http.ResponseWriter, r *http.Request) {
	creditID, err := uuid.Parse(chi.URLParam(r, "creditID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid credit ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	if appErr := h.castService.DeleteFromVideo(r.Context(), creditID, callerID, callerRoles); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
// @Summary      Archive cast member globally
// @Description  Soft-deletes a cast member by archiving them. The record and credits are preserved but the member is hidden from normal listings. Admin only.
// @Tags         cast
// @Produce      json
// @Param        id path string true "Cast member ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /cast/{id} [delete]
func (h *CastHandler) Delete(w http.ResponseWriter, r *http.Request) {
	castID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cast ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	if appErr := h.castService.Delete(r.Context(), castID, callerID, callerRoles); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unarchive godoc
// @Summary      Unarchive cast member
// @Description  Restores an archived cast member to active status. Admin only.
// @Tags         cast
// @Produce      json
// @Param        id path string true "Cast member ID"
// @Success      200 {object} httputil.Response{data=dto.CastMemberResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /cast/{id}/unarchive [post]
func (h *CastHandler) Unarchive(w http.ResponseWriter, r *http.Request) {
	castID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cast ID")
		return
	}

	callerRoles, _ := middleware.GetUserRoles(r.Context())

	out, appErr := h.castService.Unarchive(r.Context(), &castuc.UnarchiveInput{
		ID:          castID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.CastToMemberResponse(out.Cast))
}

// ListVideos godoc
// @Summary      List videos for a cast member
// @Description  Returns a paginated list of published videos a cast member appears in, ordered by release date (newest first).
// @Tags         cast
// @Produce      json
// @Param        id       path  string true  "Cast member ID" format(uuid)
// @Param        page     query int    false "Page number"    default(1)
// @Param        per_page query int    false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.CastVideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /cast/{id}/videos [get]
func (h *CastHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	castID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cast ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	out, appErr := h.castService.ListVideos(r.Context(), castuc.ListVideosInput{
		CastID:  castID,
		Page:    page,
		PerPage: perPage,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.VideoCastToVideoResponses(out.Credits, h.cdnURLs), &httputil.Meta{
		Page:       out.Page,
		PerPage:    out.PerPage,
		Total:      int64(out.Total),
		TotalPages: (out.Total + out.PerPage - 1) / out.PerPage,
	})
}
