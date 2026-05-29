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
	chapteruc "github.com/zenfulcode/zencial/internal/usecase/chapter"
)

// ChapterHandler handles chapter HTTP requests.
type ChapterHandler struct {
	service   *chapteruc.Service
	validator *validator.Validator
}

// NewChapterHandler creates a new ChapterHandler.
func NewChapterHandler(service *chapteruc.Service) *ChapterHandler {
	return &ChapterHandler{
		service:   service,
		validator: validator.New(),
	}
}

// List godoc
// @Summary      List chapters
// @Description  Returns all chapters for a video ordered by start time.
// @Tags         chapters
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=[]dto.ChapterResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /publisher/videos/{id}/chapters [get]
func (h *ChapterHandler) List(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	chapters, appErr := h.service.ListChapters(r.Context(), videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, chaptersToResponse(chapters))
}

// Replace godoc
// @Summary      Replace chapters
// @Description  Atomically replaces all chapters for a video. Publisher must own the video.
// @Tags         chapters
// @Accept       json
// @Produce      json
// @Param        id   path string                   true "Video ID"
// @Param        body body dto.ReplaceChaptersRequest true "Chapter list"
// @Success      200 {object} httputil.Response{data=[]dto.ChapterResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/chapters [put]
func (h *ChapterHandler) Replace(w http.ResponseWriter, r *http.Request) {
	videoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.ReplaceChaptersRequest
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

	items := make([]chapteruc.ChapterItem, len(req.Chapters))
	for i, c := range req.Chapters {
		items[i] = chapteruc.ChapterItem{
			StartTimeSecs: c.StartTimeSecs,
			Title:         c.Title,
			Source:        entity.ChapterSource(c.Source),
		}
	}

	chapters, appErr := h.service.ReplaceChapters(r.Context(), &chapteruc.ReplaceChaptersInput{
		VideoID:    videoID,
		UploaderID: callerID,
		Chapters:   items,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, chaptersToResponse(chapters))
}

// Delete godoc
// @Summary      Delete chapter
// @Description  Removes a single chapter. Publisher must own the video.
// @Tags         chapters
// @Produce      json
// @Param        id        path string true "Video ID"
// @Param        chapterID path string true "Chapter ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/chapters/{chapterID} [delete]
func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	chapterID, err := uuid.Parse(chi.URLParam(r, "chapterID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid chapter ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	if appErr := h.service.DeleteChapter(r.Context(), chapterID, callerID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func chapterToResponse(c *entity.Chapter) dto.ChapterResponse {
	return dto.ChapterResponse{
		ID:            c.ID.String(),
		VideoID:       c.VideoID.String(),
		StartTimeSecs: c.StartTimeSecs,
		Title:         c.Title,
		Source:        string(c.Source),
		CreatedAt:     c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func chaptersToResponse(chapters []entity.Chapter) []dto.ChapterResponse {
	out := make([]dto.ChapterResponse, len(chapters))
	for i := range chapters {
		out[i] = chapterToResponse(&chapters[i])
	}
	return out
}
