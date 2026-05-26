package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	captionuc "github.com/zenfulcode/zencial/internal/usecase/caption"
)

// CaptionHandler handles caption HTTP requests.
type CaptionHandler struct {
	service   *captionuc.Service
	validator *validator.Validator
}

// NewCaptionHandler creates a new CaptionHandler.
func NewCaptionHandler(service *captionuc.Service) *CaptionHandler {
	return &CaptionHandler{
		service:   service,
		validator: validator.New(),
	}
}

// List godoc
// @Summary      List captions
// @Description  Returns all captions for a video.
// @Tags         captions
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=[]dto.CaptionResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /publisher/videos/{id}/captions [get]
func (h *CaptionHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	captions, appErr := h.service.ListCaptions(r.Context(), videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, captionsToResponse(captions))
}

// InitiateUpload godoc
// @Summary      Initiate caption upload
// @Description  Returns a signed PUT URL to upload a caption file to object storage.
// @Tags         captions
// @Accept       json
// @Produce      json
// @Param        id   path string                          true "Video ID"
// @Param        body body dto.InitiateCaptionUploadRequest true "Caption metadata"
// @Success      200 {object} httputil.Response{data=dto.InitiateCaptionUploadResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/captions [post]
func (h *CaptionHandler) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.InitiateCaptionUploadRequest
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

	out, appErr := h.service.InitiateCaptionUpload(r.Context(), &captionuc.InitiateCaptionUploadInput{
		VideoID:      videoID,
		UploaderID:   callerID,
		LanguageCode: req.LanguageCode,
		Format:       req.Format,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, dto.InitiateCaptionUploadResponse{
		UploadURL:    out.UploadURL,
		ObjectKey:    out.ObjectKey,
		LanguageCode: req.LanguageCode,
		ExpiresAt:    out.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Register godoc
// @Summary      Register caption
// @Description  Persists a caption record after uploading the file to storage.
// @Tags         captions
// @Accept       json
// @Produce      json
// @Param        id   path string                    true "Video ID"
// @Param        body body dto.RegisterCaptionRequest true "Caption registration data"
// @Success      201 {object} httputil.Response{data=dto.CaptionResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/captions/register [post]
func (h *CaptionHandler) Register(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.RegisterCaptionRequest
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

	caption, appErr := h.service.RegisterCaption(r.Context(), &captionuc.RegisterCaptionInput{
		VideoID:      videoID,
		UploaderID:   callerID,
		LanguageCode: req.LanguageCode,
		Format:       req.Format,
		StorageKey:   req.StorageKey,
		Source:       entity.CaptionSource(req.Source),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, captionToResponse(caption))
}

// Delete godoc
// @Summary      Delete caption
// @Description  Removes a caption for a specific language. Publisher must own the video.
// @Tags         captions
// @Produce      json
// @Param        id   path string true "Video ID"
// @Param        lang path string true "Language code (e.g. en)"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/captions/{lang} [delete]
func (h *CaptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	lang := chi.URLParam(r, "lang")
	if lang == "" {
		httputil.BadRequest(w, apperror.CodeBadRequest, "language code is required")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	if appErr := h.service.DeleteCaption(r.Context(), videoID, lang, callerID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Publish godoc
// @Summary      Publish caption
// @Description  Marks a caption as reviewed and published. Publisher must own the video.
// @Tags         captions
// @Produce      json
// @Param        id   path string true "Video ID"
// @Param        lang path string true "Language code (e.g. en)"
// @Success      200 {object} httputil.Response{data=dto.CaptionResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/captions/{lang}/publish [post]
func (h *CaptionHandler) Publish(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	lang := chi.URLParam(r, "lang")
	if lang == "" {
		httputil.BadRequest(w, apperror.CodeBadRequest, "language code is required")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	caption, appErr := h.service.PublishCaption(r.Context(), videoID, lang, callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, captionToResponse(caption))
}

func captionToResponse(c *entity.Caption) dto.CaptionResponse {
	return dto.CaptionResponse{
		ID:           c.ID.String(),
		VideoID:      c.VideoID.String(),
		LanguageCode: c.LanguageCode,
		Format:       c.Format,
		StorageKey:   c.StorageKey,
		Status:       string(c.Status),
		Source:       string(c.Source),
		CreatedAt:    c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func captionsToResponse(captions []entity.Caption) []dto.CaptionResponse {
	out := make([]dto.CaptionResponse, len(captions))
	for i := range captions {
		out[i] = captionToResponse(&captions[i])
	}
	return out
}
