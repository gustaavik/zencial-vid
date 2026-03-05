package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	watchlistuc "github.com/zenfulcode/zencial/internal/usecase/watchlist"
)

// WatchlistHandler handles watchlist HTTP requests.
type WatchlistHandler struct {
	watchlistService *watchlistuc.Service
}

// NewWatchlistHandler creates a new WatchlistHandler.
func NewWatchlistHandler(watchlistService *watchlistuc.Service) *WatchlistHandler {
	return &WatchlistHandler{watchlistService: watchlistService}
}

// List godoc
// @Summary      Get user's watchlist
// @Tags         watchlist
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.WatchlistItemResponse}
// @Security     BearerAuth
// @Router       /watchlist [get]
func (h *WatchlistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)

	_, total, appErr := h.watchlistService.List(r.Context(), userID, page, perPage)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, []dto.WatchlistItemResponse{}, pagination.NewMeta(page, perPage, total))
}

// Add godoc
// @Summary      Add to watchlist
// @Tags         watchlist
// @Param        contentId path string true "Content ID"
// @Success      201
// @Security     BearerAuth
// @Router       /watchlist/{contentId} [post]
func (h *WatchlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	contentID, err := httputil.URLParamUUID(r, "contentId")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid content ID")
		return
	}
	if appErr := h.watchlistService.Add(r.Context(), userID, contentID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Remove godoc
// @Summary      Remove from watchlist
// @Tags         watchlist
// @Param        contentId path string true "Content ID"
// @Success      204
// @Security     BearerAuth
// @Router       /watchlist/{contentId} [delete]
func (h *WatchlistHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	contentID, err := httputil.URLParamUUID(r, "contentId")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid content ID")
		return
	}
	if appErr := h.watchlistService.Remove(r.Context(), userID, contentID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Status godoc
// @Summary      Check watchlist status
// @Tags         watchlist
// @Produce      json
// @Param        contentId path string true "Content ID"
// @Success      200 {object} httputil.Response{data=dto.WatchlistStatusResponse}
// @Security     BearerAuth
// @Router       /watchlist/{contentId}/status [get]
func (h *WatchlistHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	contentID, err := httputil.URLParamUUID(r, "contentId")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid content ID")
		return
	}
	inWatchlist, appErr := h.watchlistService.Status(r.Context(), userID, contentID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, dto.WatchlistStatusResponse{InWatchlist: inWatchlist})
}
