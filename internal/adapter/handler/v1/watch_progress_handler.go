package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	watchprogressuc "github.com/zenfulcode/zencial/internal/usecase/watchprogress"
)

// WatchProgressHandler handles "resume watching" / "continue watching" HTTP requests
// for the authenticated user.
type WatchProgressHandler struct {
	service   *watchprogressuc.Service
	cdnURLs   mapper.ThumbnailURLBuilder
	validator *validator.Validator
}

// NewWatchProgressHandler creates a new WatchProgressHandler.
func NewWatchProgressHandler(service *watchprogressuc.Service, cdnURLs mapper.ThumbnailURLBuilder) *WatchProgressHandler {
	return &WatchProgressHandler{
		service:   service,
		cdnURLs:   cdnURLs,
		validator: validator.New(),
	}
}

// List godoc
// @Summary      List my continue-watching feed
// @Description  Return videos the authenticated user has started but not finished, ordered by most recently watched.
// @Tags         watch-progress
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.ContinueWatchingItem,meta=httputil.Meta}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watch-progress [get]
func (h *WatchProgressHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	page := valueobject.NewPagination(
		httputil.QueryInt(r, "page", valueobject.DefaultPage),
		httputil.QueryInt(r, "per_page", valueobject.DefaultPerPage),
	)

	items, total, appErr := h.service.ListInProgress(r.Context(), userID, page)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.ContinueWatchingItemsToResponse(r.Context(), items, h.cdnURLs),
		pagination.NewMeta(page.Page, page.PerPage, total),
	)
}

// ListByUser godoc
// @Summary      List a user's continue-watching feed (admin)
// @Description  Return the given user's started-but-unfinished videos (admin only).
// @Tags         watch-progress
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.ContinueWatchingItem,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/watch-progress [get]
func (h *WatchProgressHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	page := valueobject.NewPagination(
		httputil.QueryInt(r, "page", valueobject.DefaultPage),
		httputil.QueryInt(r, "per_page", valueobject.DefaultPerPage),
	)

	items, total, appErr := h.service.ListInProgress(r.Context(), userID, page)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.ContinueWatchingItemsToResponse(r.Context(), items, h.cdnURLs),
		pagination.NewMeta(page.Page, page.PerPage, total),
	)
}

// Get godoc
// @Summary      Get my saved progress for a video
// @Description  Returns the authenticated user's last saved playback position for the video.
// @Tags         watch-progress
// @Produce      json
// @Param        video_id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.WatchProgressResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watch-progress/{video_id} [get]
func (h *WatchProgressHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	videoID, err := httputil.URLParamUUID(r, "video_id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	progress, durationSeconds, appErr := h.service.Get(r.Context(), userID, videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.WatchProgressToResponse(progress, durationSeconds))
}

// Upsert godoc
// @Summary      Save my progress for a video
// @Description  Create or update the authenticated user's playback position for the video. Idempotent.
// @Tags         watch-progress
// @Accept       json
// @Param        video_id path string true "Video ID" format(uuid)
// @Param        body body dto.UpsertWatchProgressRequest true "Position in seconds"
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watch-progress/{video_id} [put]
func (h *WatchProgressHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	videoID, err := httputil.URLParamUUID(r, "video_id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	var req dto.UpsertWatchProgressRequest
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

	if appErr := h.service.Upsert(r.Context(), userID, videoID, req.PositionSeconds); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
// @Summary      Clear my progress for a video
// @Description  Remove the authenticated user's saved playback position for the video.
// @Tags         watch-progress
// @Param        video_id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watch-progress/{video_id} [delete]
func (h *WatchProgressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	videoID, err := httputil.URLParamUUID(r, "video_id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	if appErr := h.service.Delete(r.Context(), userID, videoID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
