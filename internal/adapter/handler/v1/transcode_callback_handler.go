package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	videouc "github.com/zenfulcode/zencial/internal/usecase/video"
)

// TranscodeCallbackHandler handles internal callbacks from the CDN service when a transcode
// job finishes (success or failure). The route is mounted behind the internal-token middleware.
type TranscodeCallbackHandler struct {
	videoService *videouc.Service
}

// NewTranscodeCallbackHandler creates a new TranscodeCallbackHandler.
func NewTranscodeCallbackHandler(videoService *videouc.Service) *TranscodeCallbackHandler {
	return &TranscodeCallbackHandler{videoService: videoService}
}

// transcodeCallbackRequest is the body posted by the CDN.
type transcodeCallbackRequest struct {
	Status string `json:"status"`          // "completed" | "failed"
	Error  string `json:"error,omitempty"` // populated when status == "failed"
}

const (
	transcodeStatusCompleted = "completed"
	transcodeStatusFailed    = "failed"
)

// Handle godoc
// @Summary      Transcode completion callback
// @Description  Internal CDN callback that reports the final status of a transcode job. Authenticated via the internal shared-secret middleware, not BearerAuth.
// @Tags         internal
// @Accept       json
// @Param        id path string true "Video ID" format(uuid)
// @Param        body body transcodeCallbackRequest true "Transcode result"
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /internal/videos/{id}/transcode-callback [post]
func (h *TranscodeCallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req transcodeCallbackRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}

	switch req.Status {
	case transcodeStatusCompleted:
		if _, appErr := h.videoService.MarkTranscodeComplete(r.Context(), id); appErr != nil {
			httputil.Error(w, appErr)
			return
		}
	case transcodeStatusFailed:
		if _, appErr := h.videoService.MarkTranscodeFailed(r.Context(), id, req.Error); appErr != nil {
			httputil.Error(w, appErr)
			return
		}
	default:
		httputil.BadRequest(w, apperror.CodeValidationFailed, "status must be 'completed' or 'failed'")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
