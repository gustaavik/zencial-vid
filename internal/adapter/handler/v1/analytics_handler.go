package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/repository"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	analyticsuc "github.com/zenfulcode/zencial/internal/usecase/analytics"
)

const dateLayout = "2006-01-02"

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	analyticsService *analyticsuc.Service
	validator        *validator.Validator
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsService *analyticsuc.Service) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		validator:        validator.New(),
	}
}

// VideoStats godoc
// @Summary      Get video analytics
// @Description  Returns viewing statistics for a single video over a reporting range, including a retention curve and source/country/platform breakdowns. Publishers may only query their own videos.
// @Tags         analytics
// @Produce      json
// @Param        id path string true "Video ID"
// @Param        range query string false "Reporting range" Enums(7d, 30d, 90d, all) default(30d)
// @Success      200 {object} httputil.Response{data=dto.VideoAnalyticsResponse}
// @Failure      400 {object} httputil.ErrorResponse
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
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	out, appErr := h.analyticsService.GetVideoStats(r.Context(), &analyticsuc.VideoStatsInput{
		VideoID:     videoID,
		CallerID:    callerID,
		CallerRoles: callerRoles,
		RangeKey:    httputil.QueryString(r, "range", ""),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, dto.VideoAnalyticsResponse{
		VideoID:    out.VideoID.String(),
		Range:      out.Range.Key,
		StartDate:  out.Range.From.Format(dateLayout),
		EndDate:    out.Range.To.Format(dateLayout),
		Totals:     totalsToDTO(out.Totals),
		Deltas:     deltasToDTO(out.Deltas),
		Timeseries: timeseriesToDTO(out.Timeseries),
		Retention:  out.Retention,
		Sources:    breakdownToDTO(out.Sources),
		Countries:  breakdownToDTO(out.Countries),
		Platforms:  breakdownToDTO(out.Platforms),
	})
}

// Summary godoc
// @Summary      Get publisher analytics summary
// @Description  Returns aggregate viewing statistics over a reporting range for all videos uploaded by the authenticated publisher, including a daily time series and top videos.
// @Tags         analytics
// @Produce      json
// @Param        range query string false "Reporting range" Enums(7d, 30d, 90d, all) default(30d)
// @Success      200 {object} httputil.Response{data=dto.PublisherSummaryResponse}
// @Failure      400 {object} httputil.ErrorResponse
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

	h.writeSummary(w, r, &callerID)
}

// AdminSummary godoc
// @Summary      Get platform-wide analytics summary (admin)
// @Description  Returns aggregate viewing statistics over a reporting range for all videos on the platform (admin only).
// @Tags         analytics
// @Produce      json
// @Param        range query string false "Reporting range" Enums(7d, 30d, 90d, all) default(30d)
// @Success      200 {object} httputil.Response{data=dto.PublisherSummaryResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/analytics/summary [get]
func (h *AnalyticsHandler) AdminSummary(w http.ResponseWriter, r *http.Request) {
	h.writeSummary(w, r, nil)
}

func (h *AnalyticsHandler) writeSummary(w http.ResponseWriter, r *http.Request, uploaderID *uuid.UUID) {
	out, appErr := h.analyticsService.GetSummary(r.Context(), &analyticsuc.SummaryInput{
		UploaderID: uploaderID,
		RangeKey:   httputil.QueryString(r, "range", ""),
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	topVideos := make([]dto.TopVideoItem, len(out.TopVideos))
	for i, v := range out.TopVideos {
		topVideos[i] = dto.TopVideoItem{
			VideoID:           v.VideoID.String(),
			Title:             v.Title,
			Status:            string(v.Status),
			Views:             v.Views,
			WatchTimeMinutes:  v.WatchedSeconds / 60,
			AvgPercentWatched: v.AvgPercentWatched,
			FinishRate:        v.FinishRate,
		}
	}

	httputil.Success(w, http.StatusOK, dto.PublisherSummaryResponse{
		Range:      out.Range.Key,
		StartDate:  out.Range.From.Format(dateLayout),
		EndDate:    out.Range.To.Format(dateLayout),
		Totals:     totalsToDTO(out.Totals),
		Deltas:     deltasToDTO(out.Deltas),
		Timeseries: timeseriesToDTO(out.Timeseries),
		TopVideos:  topVideos,
	})
}

// RecordPlayback godoc
// @Summary      Record a playback heartbeat
// @Description  Ingest one cumulative playback heartbeat for a client-generated session ID. All fields are cumulative for the session, so heartbeats are idempotent.
// @Tags         analytics
// @Accept       json
// @Param        id path string true "Video ID"
// @Param        body body dto.RecordPlaybackRequest true "Cumulative playback heartbeat"
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /videos/{id}/playback-events [post]
func (h *AnalyticsHandler) RecordPlayback(w http.ResponseWriter, r *http.Request) {
	videoID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	var req dto.RecordPlaybackRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid request body")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		httputil.ErrorWithDetails(w,
			apperror.BadRequest(apperror.CodeValidationFailed, "validation failed", nil),
			errs,
		)
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid session ID")
		return
	}

	country, _ := middleware.GetCountry(r.Context())

	appErr := h.analyticsService.RecordPlayback(r.Context(), &analyticsuc.RecordPlaybackInput{
		SessionID:       sessionID,
		VideoID:         videoID,
		UserID:          userID,
		Source:          req.Source,
		Platform:        req.Platform,
		CountryCode:     country,
		PositionSeconds: req.PositionSeconds,
		WatchedSeconds:  req.WatchedSeconds,
		WatchedBuckets:  req.WatchedBuckets,
		Completed:       req.Completed,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func totalsToDTO(t repository.PlaybackTotals) dto.AnalyticsTotals {
	return dto.AnalyticsTotals{
		Views:             t.Views,
		WatchTimeMinutes:  t.WatchedSeconds / 60,
		UniqueViewers:     t.UniqueViewers,
		AvgPercentWatched: t.AvgPercentWatched,
		FinishRate:        t.FinishRate,
	}
}

func deltasToDTO(d *analyticsuc.Deltas) *dto.AnalyticsDeltas {
	if d == nil {
		return nil
	}
	return &dto.AnalyticsDeltas{
		ViewsPct:             d.ViewsPct,
		WatchTimePct:         d.WatchTimePct,
		UniqueViewersPct:     d.UniqueViewersPct,
		AvgPercentWatchedPts: d.AvgPercentWatchedPts,
		FinishRatePts:        d.FinishRatePts,
	}
}

func timeseriesToDTO(series []repository.DailyStat) []dto.AnalyticsDailyPoint {
	out := make([]dto.AnalyticsDailyPoint, len(series))
	for i, d := range series {
		out[i] = dto.AnalyticsDailyPoint{
			Date:             d.Day.UTC().Format(dateLayout),
			Views:            d.Views,
			WatchTimeMinutes: d.WatchedSeconds / 60,
		}
	}
	return out
}

func breakdownToDTO(items []repository.BreakdownItem) []dto.BreakdownItem {
	var total int64
	for _, it := range items {
		total += it.Views
	}

	out := make([]dto.BreakdownItem, len(items))
	for i, it := range items {
		var pct float64
		if total > 0 {
			pct = float64(it.Views) / float64(total) * 100
		}
		key := it.Key
		if key == "" {
			key = "unknown"
		}
		out[i] = dto.BreakdownItem{Key: key, Views: it.Views, Pct: pct}
	}
	return out
}
