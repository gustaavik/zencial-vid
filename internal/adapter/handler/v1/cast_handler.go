package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	castuc "github.com/zenfulcode/zencial/internal/usecase/cast"
)

// CastHandler handles cast HTTP requests.
type CastHandler struct {
	castService *castuc.Service
	validator   *validator.Validator
}

// NewCastHandler creates a new CastHandler.
func NewCastHandler(castService *castuc.Service) *CastHandler {
	return &CastHandler{
		castService: castService,
		validator:   validator.New(),
	}
}

// List godoc
// @Summary      List cast
// @Description  Returns all cast members for a video, ordered by sort_order
// @Tags         cast
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=[]dto.CastResponse}
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /videos/{id}/cast [get]
func (h *CastHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	cast, appErr := h.castService.List(r.Context(), videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.CastListToResponse(cast))
}

// Create godoc
// @Summary      Add cast member
// @Description  Adds a cast member to a video. Publishers may only add cast to their own videos.
// @Tags         cast
// @Accept       json
// @Produce      json
// @Param        id   path string               true "Video ID"
// @Param        body body dto.CreateCastRequest true "Cast data"
// @Success      201 {object} httputil.Response{data=dto.CastResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
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

	c, appErr := h.castService.Create(r.Context(), &castuc.CreateInput{
		VideoID:     videoID,
		Name:        req.Name,
		Role:        req.Role,
		SortOrder:   req.SortOrder,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.CastToResponse(c))
}

// Update godoc
// @Summary      Update cast member
// @Description  Updates a cast member. Publishers may only update cast for their own videos.
// @Tags         cast
// @Accept       json
// @Produce      json
// @Param        id   path string               true "Cast ID"
// @Param        body body dto.UpdateCastRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.CastResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /cast/{id} [put]
func (h *CastHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	c, appErr := h.castService.Update(r.Context(), &castuc.UpdateInput{
		ID:          castID,
		Name:        req.Name,
		Role:        req.Role,
		SortOrder:   req.SortOrder,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.CastToResponse(c))
}

// UploadPicture godoc
// @Summary      Upload cast member picture
// @Description  Uploads or replaces a cast member's picture via multipart form. Publishers may only update pictures for cast on their own videos.
// @Tags         cast
// @Accept       multipart/form-data
// @Produce      json
// @Param        id      path     string true "Cast ID" format(uuid)
// @Param        picture formData file   true "Picture image (JPEG, PNG, WebP, or GIF)"
// @Success      200 {object} httputil.Response{data=dto.CastResponse}
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

	httputil.Success(w, http.StatusOK, mapper.CastToResponse(out.Cast))
}

// Delete godoc
// @Summary      Remove cast member
// @Description  Removes a cast member. Publishers may only remove cast from their own videos.
// @Tags         cast
// @Produce      json
// @Param        id path string true "Cast ID"
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
