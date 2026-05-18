package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	analyticsuc "github.com/zenfulcode/zencial/internal/usecase/analytics"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	analyticsService *analyticsuc.Service
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsService *analyticsuc.Service) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// VideoStats godoc
// @Summary      Get video analytics
// @Description  Returns viewing statistics for a single video. Publishers may only query their own videos.
// @Tags         analytics
// @Produce      json
// @Param        id path string true "Video ID"
// @Success      200 {object} httputil.Response{data=dto.VideoStatsResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/videos/{id}/analytics [get]
func (h *AnalyticsHandler) VideoStats(w http.ResponseWriter, r *http.Request) {
	videoID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRole, _ := middleware.GetUserRole(r.Context())

	stats, appErr := h.analyticsService.GetVideoStats(r.Context(), videoID, callerID, callerRole)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, dto.VideoStatsResponse{
		VideoID:        stats.VideoID.String(),
		TotalViewers:   stats.TotalViewers,
		AvgProgressPct: stats.AvgProgressPct,
		CompletionRate: stats.CompletionRate,
	})
}

// Summary godoc
// @Summary      Get publisher analytics summary
// @Description  Returns aggregate viewing statistics for all videos uploaded by the authenticated publisher
// @Tags         analytics
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.VideoStatsResponse}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/analytics/summary [get]
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	stats, appErr := h.analyticsService.GetSummary(r.Context(), callerID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	out := make([]dto.VideoStatsResponse, len(stats))
	for i, s := range stats {
		out[i] = dto.VideoStatsResponse{
			VideoID:        s.VideoID.String(),
			TotalViewers:   s.TotalViewers,
			AvgProgressPct: s.AvgProgressPct,
			CompletionRate: s.CompletionRate,
		}
	}
	httputil.Success(w, http.StatusOK, out)
}
