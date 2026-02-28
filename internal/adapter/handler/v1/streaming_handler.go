package v1

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	streaminguc "github.com/zenfulcode/zencial/internal/usecase/streaming"
)

// StreamingHandler handles streaming HTTP requests.
type StreamingHandler struct {
	streamingService *streaminguc.Service
	validator        *validator.Validator
}

// NewStreamingHandler creates a new StreamingHandler.
func NewStreamingHandler(streamingService *streaminguc.Service) *StreamingHandler {
	return &StreamingHandler{streamingService: streamingService, validator: validator.New()}
}

// StartSession godoc
// @Summary      Start a streaming session
// @Tags         streaming
// @Accept       json
// @Produce      json
// @Param        body body dto.StartSessionRequest true "Session data"
// @Success      201 {object} httputil.Response{data=dto.SessionResponse}
// @Security     BearerAuth
// @Router       /streaming/sessions [post]
func (h *StreamingHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}

	var req dto.StartSessionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	contentID, _ := uuid.Parse(req.ContentID)
	var episodeID *uuid.UUID
	if req.EpisodeID != nil {
		eid, _ := uuid.Parse(*req.EpisodeID)
		episodeID = &eid
	}

	output, appErr := h.streamingService.StartSession(r.Context(), streaminguc.StartSessionInput{
		UserID: userID, ContentID: contentID, EpisodeID: episodeID,
		DeviceInfo: req.DeviceInfo, IPAddress: r.RemoteAddr,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, dto.SessionResponse{
		ID:          output.Session.ID.String(),
		ManifestURL: output.ManifestURL,
		ExpiresAt:   output.Session.StartedAt.Add(4 * time.Hour).Format("2006-01-02T15:04:05Z"),
	})
}

// EndSession godoc
// @Summary      End a streaming session
// @Tags         streaming
// @Param        id path string true "Session ID"
// @Success      204
// @Security     BearerAuth
// @Router       /streaming/sessions/{id} [delete]
func (h *StreamingHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	sessionID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid session ID")
		return
	}
	if appErr := h.streamingService.EndSession(r.Context(), userID, sessionID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateProgress godoc
// @Summary      Update playback progress
// @Tags         streaming
// @Accept       json
// @Param        body body dto.UpdateProgressRequest true "Progress data"
// @Success      204
// @Security     BearerAuth
// @Router       /streaming/progress [put]
func (h *StreamingHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	var req dto.UpdateProgressRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	contentID, _ := uuid.Parse(req.ContentID)
	var episodeID *uuid.UUID
	if req.EpisodeID != nil {
		eid, _ := uuid.Parse(*req.EpisodeID)
		episodeID = &eid
	}

	if appErr := h.streamingService.UpdateProgress(r.Context(), streaminguc.UpdateProgressInput{
		UserID: userID, ContentID: contentID, EpisodeID: episodeID,
		Position: req.Position, Duration: req.Duration,
	}); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetProgress godoc
// @Summary      Get playback progress
// @Tags         streaming
// @Produce      json
// @Param        contentId path string true "Content ID"
// @Success      200 {object} httputil.Response{data=dto.ProgressResponse}
// @Security     BearerAuth
// @Router       /streaming/progress/{contentId} [get]
func (h *StreamingHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	contentID, err := httputil.URLParamUUID(r, "contentId")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid content ID")
		return
	}

	progress, appErr := h.streamingService.GetProgress(r.Context(), userID, contentID, nil)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	if progress == nil {
		httputil.Success(w, http.StatusOK, dto.ProgressResponse{ContentID: contentID.String()})
		return
	}

	resp := dto.ProgressResponse{
		ContentID: progress.ContentID.String(), Position: progress.Position,
		Duration: progress.Duration, Percentage: progress.Percentage(), Completed: progress.Completed,
	}
	if progress.EpisodeID != nil {
		eid := progress.EpisodeID.String()
		resp.EpisodeID = &eid
	}
	httputil.Success(w, http.StatusOK, resp)
}

// ContinueWatching godoc
// @Summary      Get continue watching list
// @Tags         streaming
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.ContinueWatchingResponse}
// @Security     BearerAuth
// @Router       /streaming/continue-watching [get]
func (h *StreamingHandler) ContinueWatching(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, "UNAUTHORIZED", "authentication required")
		return
	}
	limit := httputil.QueryInt(r, "limit", 20)
	_, appErr := h.streamingService.ContinueWatching(r.Context(), userID, limit)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, []dto.ContinueWatchingResponse{})
}
