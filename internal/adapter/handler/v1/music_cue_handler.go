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
	musiccueuc "github.com/zenfulcode/zencial/internal/usecase/musiccue"
)

// MusicCueHandler handles music cue HTTP requests.
type MusicCueHandler struct {
	service   *musiccueuc.Service
	validator *validator.Validator
}

// NewMusicCueHandler creates a new MusicCueHandler.
func NewMusicCueHandler(service *musiccueuc.Service) *MusicCueHandler {
	return &MusicCueHandler{
		service:   service,
		validator: validator.New(),
	}
}

// List godoc
// @Summary      List music cues
// @Description  Returns all music cues for a video ordered by timecode.
// @Tags         music-cues
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=[]dto.MusicCueResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /publisher/videos/{id}/music-cues [get]
func (h *MusicCueHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	cues, appErr := h.service.ListCues(r.Context(), videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, musicCuesToResponse(cues))
}

// Create godoc
// @Summary      Add music cue
// @Description  Adds a music cue to a video. Publisher must own the video.
// @Tags         music-cues
// @Accept       json
// @Produce      json
// @Param        id   path string                   true "Video ID"
// @Param        body body dto.CreateMusicCueRequest true "Music cue data"
// @Success      201 {object} httputil.Response{data=dto.MusicCueResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/music-cues [post]
func (h *MusicCueHandler) Create(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.CreateMusicCueRequest
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

	cue, appErr := h.service.CreateCue(r.Context(), &musiccueuc.CreateCueInput{
		VideoID:         videoID,
		UploaderID:      callerID,
		TimecodeSeconds: req.TimecodeSeconds,
		Title:           req.Title,
		ComposerArtist:  req.ComposerArtist,
		UseType:         entity.MusicUseType(req.UseType),
		RightsStatus:    entity.MusicRightsStatus(req.RightsStatus),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, musicCueToResponse(cue))
}

// Update godoc
// @Summary      Update music cue
// @Description  Updates a music cue's metadata. Publisher must own the video.
// @Tags         music-cues
// @Accept       json
// @Produce      json
// @Param        id    path string                   true "Video ID"
// @Param        cueID path string                   true "Music cue ID"
// @Param        body  body dto.UpdateMusicCueRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.MusicCueResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/music-cues/{cueID} [put]
func (h *MusicCueHandler) Update(w http.ResponseWriter, r *http.Request) {
	cueID, err := uuid.Parse(chi.URLParam(r, "cueID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cue ID")
		return
	}

	var req dto.UpdateMusicCueRequest
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

	input := &musiccueuc.UpdateCueInput{
		CueID:           cueID,
		UploaderID:      callerID,
		TimecodeSeconds: req.TimecodeSeconds,
		Title:           req.Title,
		ComposerArtist:  req.ComposerArtist,
	}
	if req.UseType != nil {
		ut := entity.MusicUseType(*req.UseType)
		input.UseType = &ut
	}
	if req.RightsStatus != nil {
		rs := entity.MusicRightsStatus(*req.RightsStatus)
		input.RightsStatus = &rs
	}

	cue, appErr := h.service.UpdateCue(r.Context(), input)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, musicCueToResponse(cue))
}

// Delete godoc
// @Summary      Delete music cue
// @Description  Removes a music cue. Publisher must own the video.
// @Tags         music-cues
// @Produce      json
// @Param        id    path string true "Video ID"
// @Param        cueID path string true "Music cue ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/music-cues/{cueID} [delete]
func (h *MusicCueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cueID, err := uuid.Parse(chi.URLParam(r, "cueID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cue ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	if appErr := h.service.DeleteCue(r.Context(), cueID, callerID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InitiateClearanceUpload godoc
// @Summary      Initiate clearance upload
// @Description  Returns a signed PUT URL to upload a music rights clearance document.
// @Tags         music-cues
// @Produce      json
// @Param        id    path string true "Video ID"
// @Param        cueID path string true "Music cue ID"
// @Success      200 {object} httputil.Response{data=dto.InitiateClearanceUploadResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/music-cues/{cueID}/clearance [post]
func (h *MusicCueHandler) InitiateClearanceUpload(w http.ResponseWriter, r *http.Request) {
	cueID, err := uuid.Parse(chi.URLParam(r, "cueID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cue ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	out, appErr := h.service.InitiateClearanceUpload(r.Context(), cueID, callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, dto.InitiateClearanceUploadResponse{
		UploadURL: out.UploadURL,
		ObjectKey: out.ObjectKey,
		ExpiresAt: out.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// CompleteClearanceUpload godoc
// @Summary      Complete clearance upload
// @Description  Records a clearance document and marks the cue as cleared.
// @Tags         music-cues
// @Accept       json
// @Produce      json
// @Param        id    path string                           true "Video ID"
// @Param        cueID path string                           true "Music cue ID"
// @Param        body  body dto.CompleteClearanceUploadRequest true "Object key"
// @Success      200 {object} httputil.Response{data=dto.MusicCueResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/music-cues/{cueID}/clearance/complete [post]
func (h *MusicCueHandler) CompleteClearanceUpload(w http.ResponseWriter, r *http.Request) {
	cueID, err := uuid.Parse(chi.URLParam(r, "cueID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid cue ID")
		return
	}

	var req dto.CompleteClearanceUploadRequest
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

	cue, appErr := h.service.CompleteClearanceUpload(r.Context(), cueID, req.ObjectKey, callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, musicCueToResponse(cue))
}

func musicCueToResponse(c *entity.MusicCue) dto.MusicCueResponse {
	return dto.MusicCueResponse{
		ID:                   c.ID.String(),
		VideoID:              c.VideoID.String(),
		TimecodeSeconds:      c.TimecodeSeconds,
		Title:                c.Title,
		ComposerArtist:       c.ComposerArtist,
		UseType:              string(c.UseType),
		RightsStatus:         string(c.RightsStatus),
		ClearanceDocumentKey: c.ClearanceDocumentKey,
		CreatedAt:            c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func musicCuesToResponse(cues []entity.MusicCue) []dto.MusicCueResponse {
	out := make([]dto.MusicCueResponse, len(cues))
	for i := range cues {
		out[i] = musicCueToResponse(&cues[i])
	}
	return out
}
