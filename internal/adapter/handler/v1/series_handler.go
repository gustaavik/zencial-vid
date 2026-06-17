package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/middleware"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	seriesuc "github.com/zenfulcode/zencial/internal/usecase/series"
)

// SeriesHandler handles series HTTP requests.
type SeriesHandler struct {
	service   *seriesuc.Service
	cdnURLs   mapper.ThumbnailURLBuilder
	validator *validator.Validator
}

// NewSeriesHandler creates a new SeriesHandler.
func NewSeriesHandler(service *seriesuc.Service, cdnURLs mapper.ThumbnailURLBuilder) *SeriesHandler {
	return &SeriesHandler{
		service:   service,
		cdnURLs:   cdnURLs,
		validator: validator.New(),
	}
}

// ListPublished godoc
// @Summary      List published series
// @Description  Returns all published series with optional filtering.
// @Tags         series
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        title query string false "Filter by title (partial match)"
// @Param        creator query string false "Filter by creator (partial match)"
// @Success      200 {object} httputil.Response{data=[]dto.SeriesResponse,meta=httputil.Meta}
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /series [get]
func (h *SeriesHandler) ListPublished(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.SeriesFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid filter parameters")
		return
	}

	series, total, appErr := h.service.ListPublished(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.SeriesToResponseMany(series),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total),
	)
}

// GetByID godoc
// @Summary      Get series by ID
// @Description  Returns a single series by its UUID.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Router       /series/{id} [get]
func (h *SeriesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	series, appErr := h.service.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesToResponse(series))
}

// ListEpisodes godoc
// @Summary      List episodes for a series
// @Description  Returns all videos assigned to the series, ordered by season and episode number.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.VideoResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Router       /series/{id}/episodes [get]
func (h *SeriesHandler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	fs, err := filter.FromRequest(r, postgres.VideoFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid filter parameters")
		return
	}

	episodes, total, appErr := h.service.ListEpisodes(r.Context(), id, &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.VideosToResponse(r.Context(), episodes, h.cdnURLs),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total),
	)
}

// GetNextEpisode godoc
// @Summary      Get next episode to watch
// @Description  Returns the next unwatched episode for the authenticated user. Returns the first episode if no watch progress exists.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/next-episode [get]
func (h *SeriesHandler) GetNextEpisode(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	episode, appErr := h.service.GetNextEpisode(r.Context(), userID, id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.VideoToResponse(r.Context(), episode, h.cdnURLs))
}

// UpdateWatchProgress godoc
// @Summary      Update series watch progress
// @Description  Records the last-watched episode for the authenticated user.
// @Tags         series
// @Accept       json
// @Param        id path string true "Series ID" format(uuid)
// @Param        body body dto.UpdateSeriesWatchProgressRequest true "Episode to mark as watched"
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/watch-progress [put]
func (h *SeriesHandler) UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	seriesID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	var req dto.UpdateSeriesWatchProgressRequest
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

	episodeID, err := uuid.Parse(req.EpisodeID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid episode ID")
		return
	}

	if appErr := h.service.UpdateWatchProgress(r.Context(), userID, seriesID, episodeID); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetWatchProgress godoc
// @Summary      Get series watch progress
// @Description  Returns the last-watched episode for the authenticated user in a given series.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.SeriesWatchProgressResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/watch-progress [get]
func (h *SeriesHandler) GetWatchProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	seriesID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	progress, appErr := h.service.GetWatchProgress(r.Context(), userID, seriesID)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesWatchProgressToResponse(progress))
}

// Create godoc
// @Summary      Create a new series
// @Description  Creates a new series in draft status.
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateSeriesRequest true "Series details"
// @Success      201 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series [post]
func (h *SeriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	var req dto.CreateSeriesRequest
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

	genreIDs, _ := parseUUIDs(req.GenreIDs)

	out, appErr := h.service.Create(r.Context(), &seriesuc.CreateInput{
		Title:               req.Title,
		Description:         req.Description,
		Creator:             req.Creator,
		SeriesType:          req.SeriesType,
		Logline:             req.Logline,
		PrimaryLanguage:     req.PrimaryLanguage,
		OriginCountry:       req.OriginCountry,
		ContentRating:       req.ContentRating,
		CoverImageKey:       req.CoverImageKey,
		PosterKey:           req.PosterKey,
		BannerKey:           req.BannerKey,
		TitleLogoKey:        req.TitleLogoKey,
		UploadedBy:          callerID,
		GenreIDs:            genreIDs,
		MinimumPlanLevel:    req.MinimumPlanLevel,
		AutoplayNext:        req.AutoplayNext,
		BingeMode:           req.BingeMode,
		HideEpisodeCount:    req.HideEpisodeCount,
		DefaultVisibility:   req.DefaultVisibility,
		DefaultMonetization: req.DefaultMonetization,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.SeriesToResponse(out.Series))
}

// Update godoc
// @Summary      Update a series
// @Description  Updates series metadata. Publisher must own the series; admin can update any.
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Param        body body dto.UpdateSeriesRequest true "Fields to update"
// @Success      200 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id} [put]
func (h *SeriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	var req dto.UpdateSeriesRequest
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

	var genreIDs []uuid.UUID
	if req.GenreIDs != nil {
		genreIDs, _ = parseUUIDs(req.GenreIDs)
	}

	series, appErr := h.service.Update(r.Context(), &seriesuc.UpdateInput{
		ID:                  id,
		CallerID:            callerID,
		CallerRoles:         callerRoles,
		Title:               req.Title,
		Description:         req.Description,
		Creator:             req.Creator,
		SeriesType:          req.SeriesType,
		Logline:             req.Logline,
		PrimaryLanguage:     req.PrimaryLanguage,
		OriginCountry:       req.OriginCountry,
		ContentRating:       req.ContentRating,
		CoverImageKey:       req.CoverImageKey,
		PosterKey:           req.PosterKey,
		BannerKey:           req.BannerKey,
		TitleLogoKey:        req.TitleLogoKey,
		GenreIDs:            genreIDs,
		MinimumPlanLevel:    req.MinimumPlanLevel,
		AutoplayNext:        req.AutoplayNext,
		BingeMode:           req.BingeMode,
		HideEpisodeCount:    req.HideEpisodeCount,
		DefaultVisibility:   req.DefaultVisibility,
		DefaultMonetization: req.DefaultMonetization,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesToResponse(series))
}

// AddEpisode godoc
// @Summary      Add an episode to a series
// @Description  Links an existing published video to the series as a numbered episode.
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Param        body body dto.AddEpisodeRequest true "Episode details"
// @Success      201 {object} httputil.Response{data=dto.VideoResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/episodes [post]
func (h *SeriesHandler) AddEpisode(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	seriesID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	var req dto.AddEpisodeRequest
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

	videoID, err := uuid.Parse(req.VideoID)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	video, appErr := h.service.AddEpisode(r.Context(), &seriesuc.AddEpisodeInput{
		SeriesID:      seriesID,
		VideoID:       videoID,
		SeasonNumber:  req.SeasonNumber,
		EpisodeNumber: req.EpisodeNumber,
		CallerID:      callerID,
		CallerRoles:   callerRoles,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.VideoToResponse(r.Context(), video, h.cdnURLs))
}

// RemoveEpisode godoc
// @Summary      Remove an episode from a series
// @Description  Unlinks a video from the series. Publisher must own the series; admin can remove any.
// @Tags         series
// @Param        id path string true "Series ID" format(uuid)
// @Param        videoID path string true "Video (episode) ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/episodes/{videoID} [delete]
func (h *SeriesHandler) RemoveEpisode(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	seriesID, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	videoID, err := httputil.URLParamUUID(r, "videoID")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid video ID")
		return
	}

	if appErr := h.service.RemoveEpisode(r.Context(), &seriesuc.RemoveEpisodeInput{
		SeriesID:    seriesID,
		VideoID:     videoID,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	}); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListOwned godoc
// @Summary      List series owned by the caller
// @Description  Returns all series uploaded by the authenticated publisher.
// @Tags         series
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.SeriesResponse,meta=httputil.Meta}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series [get]
func (h *SeriesHandler) ListOwned(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}

	fs, err := filter.FromRequest(r, postgres.SeriesFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid filter parameters")
		return
	}

	series, total, appErr := h.service.ListOwned(r.Context(), callerID, &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.SeriesToResponseMany(series),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total),
	)
}

// PublishOwned godoc
// @Summary      Publish an owned series
// @Description  Transitions a draft series to published status. Publisher must own the series.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series/{id}/publish [post]
func (h *SeriesHandler) PublishOwned(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	series, appErr := h.service.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	if !entity.HasRole(callerRoles, entity.RoleAdmin) && series.UploadedBy != callerID {
		httputil.Error(w, apperror.Forbidden(apperror.CodeSeriesOwnershipRequired, "you do not own this series", nil))
		return
	}

	result, appErr := h.service.Publish(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesToResponse(result))
}

// ArchiveOwned godoc
// @Summary      Archive an owned series
// @Description  Soft-deletes a series. Publisher must own it; admin can archive any.
// @Tags         series
// @Param        id path string true "Series ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /publisher/series/{id} [delete]
func (h *SeriesHandler) ArchiveOwned(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	if appErr := h.service.Archive(r.Context(), &seriesuc.ArchiveInput{
		SeriesID:    id,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	}); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminPublish godoc
// @Summary      Publish a series (admin)
// @Description  Admin-only: publish any series regardless of ownership.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/publish [post]
func (h *SeriesHandler) AdminPublish(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	series, appErr := h.service.Publish(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesToResponse(series))
}

// AdminArchive godoc
// @Summary      Archive a series (admin)
// @Description  Admin-only: archive any series.
// @Tags         series
// @Param        id path string true "Series ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id} [delete]
func (h *SeriesHandler) AdminArchive(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w, apperror.CodeUnauthorized, "authentication required")
		return
	}
	callerRoles, _ := middleware.GetUserRoles(r.Context())

	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	if appErr := h.service.Archive(r.Context(), &seriesuc.ArchiveInput{
		SeriesID:    id,
		CallerID:    callerID,
		CallerRoles: callerRoles,
	}); appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminUnarchive godoc
// @Summary      Restore an archived series (admin)
// @Description  Admin-only: restore an archived series back to draft status.
// @Tags         series
// @Produce      json
// @Param        id path string true "Series ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.SeriesResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id}/unarchive [post]
func (h *SeriesHandler) AdminUnarchive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid series ID")
		return
	}

	series, appErr := h.service.Unarchive(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.SeriesToResponse(series))
}

// AdminListAll godoc
// @Summary      List all series (admin)
// @Description  Admin-only: returns all series regardless of status.
// @Tags         series
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.SeriesResponse,meta=httputil.Meta}
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/series [get]
func (h *SeriesHandler) AdminListAll(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.SeriesFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid filter parameters")
		return
	}

	series, total, appErr := h.service.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w,
		mapper.SeriesToResponseMany(series),
		pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total),
	)
}
