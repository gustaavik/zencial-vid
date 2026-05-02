package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/storage"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	watchlistuc "github.com/zenfulcode/zencial/internal/usecase/watchlist"
)

// WatchlistHandler handles watchlist HTTP requests for the authenticated user.
type WatchlistHandler struct {
	service *watchlistuc.Service
	storage storage.StorageService
}

// NewWatchlistHandler creates a new WatchlistHandler.
func NewWatchlistHandler(service *watchlistuc.Service, storageSvc storage.StorageService) *WatchlistHandler {
	return &WatchlistHandler{service: service, storage: storageSvc}
}

// List godoc
// @Summary      List my watchlist
// @Description  Return the authenticated user's watchlist as a paginated list of videos, ordered by add date (newest first).
// @Tags         watchlist
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watchlist [get]
func (h *WatchlistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	page := valueobject.NewPagination(
		httputil.QueryInt(r, "page", valueobject.DefaultPage),
		httputil.QueryInt(r, "per_page", valueobject.DefaultPerPage),
	)

	videos, total, appErr := h.service.List(r.Context(), userID, page)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.VideosToResponse(r.Context(), videos, h.storage),
		pagination.NewMeta(page.Page, page.PerPage, total),
	)
}

// GetStatus godoc
// @Summary      Check watchlist membership
// @Description  Returns 200 with in_watchlist=true if the video is in the authenticated user's watchlist, otherwise 404.
// @Tags         watchlist
// @Produce      json
// @Param        video_id path string true "Video ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.WatchlistStatusResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watchlist/{video_id} [get]
func (h *WatchlistHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
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

	exists, appErr := h.service.IsInWatchlist(r.Context(), userID, videoID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	if !exists {
		httputil.NotFound(w, apperror.CodeWatchlistEntryNotFound, "watchlist entry not found")
		return
	}

	httputil.Success(w, http.StatusOK, dto.WatchlistStatusResponse{InWatchlist: true})
}

// Add godoc
// @Summary      Add a video to my watchlist
// @Description  Add the given video to the authenticated user's watchlist. Idempotent — re-adding an existing entry is a no-op.
// @Tags         watchlist
// @Param        video_id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watchlist/{video_id} [post]
func (h *WatchlistHandler) Add(w http.ResponseWriter, r *http.Request) {
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

	if appErr := h.service.Add(r.Context(), userID, videoID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListByUser godoc
// @Summary      List a user's watchlist (admin)
// @Description  Return the given user's watchlist as a paginated list of videos (admin only).
// @Tags         watchlist
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/users/{id}/watchlist [get]
func (h *WatchlistHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid user ID")
		return
	}

	page := valueobject.NewPagination(
		httputil.QueryInt(r, "page", valueobject.DefaultPage),
		httputil.QueryInt(r, "per_page", valueobject.DefaultPerPage),
	)

	videos, total, appErr := h.service.List(r.Context(), userID, page)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.VideosToResponse(r.Context(), videos, h.storage),
		pagination.NewMeta(page.Page, page.PerPage, total),
	)
}

// Remove godoc
// @Summary      Remove a video from my watchlist
// @Description  Remove the given video from the authenticated user's watchlist. Returns 404 when the entry does not exist.
// @Tags         watchlist
// @Param        video_id path string true "Video ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /me/watchlist/{video_id} [delete]
func (h *WatchlistHandler) Remove(w http.ResponseWriter, r *http.Request) {
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

	if appErr := h.service.Remove(r.Context(), userID, videoID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
