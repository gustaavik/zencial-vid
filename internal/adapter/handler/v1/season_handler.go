package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	seasonuc "github.com/zenfulcode/zencial/internal/usecase/season"
)

// SeasonHandler handles season HTTP requests.
type SeasonHandler struct {
	service   *seasonuc.Service
	validator *validator.Validator
}

// NewSeasonHandler creates a new SeasonHandler.
func NewSeasonHandler(service *seasonuc.Service) *SeasonHandler {
	return &SeasonHandler{
		service:   service,
		validator: validator.New(),
	}
}

// List godoc
// @Summary      List seasons
// @Description  Returns all seasons for a series ordered by season number.
// @Tags         seasons
// @Produce      json
// @Param        id path string true "Series ID"
// @Success      200 {object} httputil.Response{data=[]dto.SeasonResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /publisher/series/{id}/seasons [get]
func (h *SeasonHandler) List(w http.ResponseWriter, r *http.Request) {
	seriesID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	seasons, appErr := h.service.ListSeasons(r.Context(), seriesID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, seasonsToResponse(seasons))
}

// Create godoc
// @Summary      Create season
// @Description  Adds a new season to a series. Publisher must own the series.
// @Tags         seasons
// @Accept       json
// @Produce      json
// @Param        id   path string                true "Series ID"
// @Param        body body dto.CreateSeasonRequest true "Season data"
// @Success      201 {object} httputil.Response{data=dto.SeasonResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series/{id}/seasons [post]
func (h *SeasonHandler) Create(w http.ResponseWriter, r *http.Request) {
	seriesID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	var req dto.CreateSeasonRequest
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

	input := &seasonuc.CreateSeasonInput{
		SeriesID:        seriesID,
		UploaderID:      callerID,
		SeasonNumber:    req.SeasonNumber,
		SeasonTag:       req.SeasonTag,
		PlannedEpisodes: req.PlannedEpisodes,
		AvgRuntimeSecs:  req.AvgRuntimeSecs,
		ReleaseCadence:  entity.ReleaseCadence(req.ReleaseCadence),
		CadenceDay:      req.CadenceDay,
		Timezone:        req.Timezone,
	}
	if req.PremiereDate != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.PremiereDate)
		if parseErr != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, "premiere_date must be RFC3339")
			return
		}
		input.PremiereDate = &t
	}

	season, appErr := h.service.CreateSeason(r.Context(), input)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, seasonToResponse(season))
}

// Update godoc
// @Summary      Update season
// @Description  Updates a season's metadata. Publisher must own the series.
// @Tags         seasons
// @Accept       json
// @Produce      json
// @Param        id       path string                true "Series ID"
// @Param        seasonID path string                true "Season ID"
// @Param        body     body dto.UpdateSeasonRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.SeasonResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series/{id}/seasons/{seasonID} [put]
func (h *SeasonHandler) Update(w http.ResponseWriter, r *http.Request) {
	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid season ID")
		return
	}

	var req dto.UpdateSeasonRequest
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

	input := &seasonuc.UpdateSeasonInput{
		SeasonID:        seasonID,
		UploaderID:      callerID,
		SeasonTag:       req.SeasonTag,
		PlannedEpisodes: req.PlannedEpisodes,
		AvgRuntimeSecs:  req.AvgRuntimeSecs,
		CadenceDay:      req.CadenceDay,
		Timezone:        req.Timezone,
	}
	if req.ReleaseCadence != nil {
		rc := entity.ReleaseCadence(*req.ReleaseCadence)
		input.ReleaseCadence = &rc
	}
	if req.PremiereDate != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.PremiereDate)
		if parseErr != nil {
			httputil.BadRequest(w, apperror.CodeValidationFailed, "premiere_date must be RFC3339")
			return
		}
		input.PremiereDate = &t
	}

	season, appErr := h.service.UpdateSeason(r.Context(), input)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, seasonToResponse(season))
}

// Delete godoc
// @Summary      Delete season
// @Description  Removes a season. Fails if it has episodes. Publisher must own the series.
// @Tags         seasons
// @Produce      json
// @Param        id       path string true "Series ID"
// @Param        seasonID path string true "Season ID"
// @Success      204
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series/{id}/seasons/{seasonID} [delete]
func (h *SeasonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid season ID")
		return
	}

	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	if appErr := h.service.DeleteSeason(r.Context(), seasonID, callerID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func seasonToResponse(s *entity.Season) dto.SeasonResponse {
	r := dto.SeasonResponse{
		ID:              s.ID.String(),
		SeriesID:        s.SeriesID.String(),
		SeasonNumber:    s.SeasonNumber,
		SeasonTag:       s.SeasonTag,
		PlannedEpisodes: s.PlannedEpisodes,
		AvgRuntimeSecs:  s.AvgRuntimeSecs,
		ReleaseCadence:  string(s.ReleaseCadence),
		CadenceDay:      s.CadenceDay,
		Timezone:        s.Timezone,
		CreatedAt:       s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       s.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if s.PremiereDate != nil {
		t := s.PremiereDate.UTC().Format("2006-01-02T15:04:05Z")
		r.PremiereDate = &t
	}
	return r
}

func seasonsToResponse(seasons []entity.Season) []dto.SeasonResponse {
	out := make([]dto.SeasonResponse, len(seasons))
	for i := range seasons {
		out[i] = seasonToResponse(&seasons[i])
	}
	return out
}
