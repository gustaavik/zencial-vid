package v1

import (
	"net/http"
	"strconv"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	contentuc "github.com/zenfulcode/zencial/internal/usecase/content"
)

// ContentHandler handles content HTTP requests.
type ContentHandler struct {
	contentService *contentuc.Service
	validator      *validator.Validator
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(contentService *contentuc.Service) *ContentHandler {
	return &ContentHandler{contentService: contentService, validator: validator.New()}
}

// List godoc
// @Summary      List content
// @Description  List and filter content with pagination
// @Tags         content
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        type query string false "Content type" Enums(film, series)
// @Param        sort_by query string false "Sort field" Enums(relevance, release_date, title)
// @Success      200 {object} httputil.Response{data=[]dto.ContentListResponse}
// @Security     BearerAuth
// @Router       /content [get]
func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)
	contentType := httputil.QueryString(r, "type", "")
	sortBy := httputil.QueryString(r, "sort_by", "relevance")

	criteria := entity.SearchCriteria{
		SortBy:  sortBy,
		Page:    page,
		PerPage: perPage,
	}
	if contentType != "" {
		ct := entity.ContentType(contentType)
		criteria.Type = &ct
	}

	contents, total, appErr := h.contentService.List(r.Context(), criteria)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, mapper.ContentsToListResponse(contents), pagination.NewMeta(page, perPage, total))
}

// Featured godoc
// @Summary      Get featured content
// @Tags         content
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.ContentListResponse}
// @Security     BearerAuth
// @Router       /content/featured [get]
func (h *ContentHandler) Featured(w http.ResponseWriter, r *http.Request) {
	limit := httputil.QueryInt(r, "limit", 10)
	contents, appErr := h.contentService.Featured(r.Context(), limit)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.ContentsToListResponse(contents))
}

// GetBySlug godoc
// @Summary      Get content by slug
// @Tags         content
// @Produce      json
// @Param        slug path string true "Content slug"
// @Success      200 {object} httputil.Response{data=dto.ContentDetailResponse}
// @Security     BearerAuth
// @Router       /content/{slug} [get]
func (h *ContentHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := httputil.URLParam(r, "slug")
	content, appErr := h.contentService.GetBySlug(r.Context(), slug)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.ContentToDetailResponse(content))
}

// GetSeasons godoc
// @Summary      Get seasons for a series
// @Tags         content
// @Produce      json
// @Param        slug path string true "Content slug"
// @Success      200 {object} httputil.Response{data=[]dto.SeasonResponse}
// @Security     BearerAuth
// @Router       /content/{slug}/seasons [get]
func (h *ContentHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	slug := httputil.URLParam(r, "slug")
	seasons, appErr := h.contentService.GetSeasons(r.Context(), slug)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.SeasonsToResponse(seasons))
}

// GetEpisodes godoc
// @Summary      Get episodes for a season
// @Tags         content
// @Produce      json
// @Param        slug path string true "Content slug"
// @Param        seasonNumber path int true "Season number"
// @Success      200 {object} httputil.Response{data=[]dto.EpisodeResponse}
// @Security     BearerAuth
// @Router       /content/{slug}/seasons/{seasonNumber}/episodes [get]
func (h *ContentHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	slug := httputil.URLParam(r, "slug")
	seasonNumberStr := httputil.URLParam(r, "seasonNumber")
	seasonNumber, err := strconv.Atoi(seasonNumberStr)
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid season number")
		return
	}
	episodes, appErr := h.contentService.GetEpisodes(r.Context(), slug, seasonNumber)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.EpisodesToResponse(episodes))
}

// Search godoc
// @Summary      Search content
// @Tags         content
// @Produce      json
// @Param        q query string true "Search query"
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.ContentListResponse}
// @Security     BearerAuth
// @Router       /search [get]
func (h *ContentHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := httputil.QueryString(r, "q", "")
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)

	criteria := entity.SearchCriteria{Query: query, Page: page, PerPage: perPage}
	contents, total, appErr := h.contentService.Search(r.Context(), criteria)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, mapper.ContentsToListResponse(contents), pagination.NewMeta(page, perPage, total))
}

// Create godoc
// @Summary      Create content (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateContentRequest true "Content data"
// @Success      201 {object} httputil.Response{data=dto.ContentDetailResponse}
// @Security     BearerAuth
// @Router       /admin/content [post]
func (h *ContentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateContentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}

	content, appErr := h.contentService.Create(r.Context(), contentuc.CreateContentInput{
		Type: req.Type, Title: req.Title, Description: req.Description,
		Synopsis: req.Synopsis, Rating: req.Rating, ReleaseYear: req.ReleaseYear,
		PosterURL: req.PosterURL, BackdropURL: req.BackdropURL,
		TrailerURL: req.TrailerURL, Director: req.Director,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusCreated, mapper.ContentToDetailResponse(content))
}

// Update godoc
// @Summary      Update content (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path string true "Content ID"
// @Param        body body dto.UpdateContentRequest true "Content update data"
// @Success      200 {object} httputil.Response{data=dto.ContentDetailResponse}
// @Security     BearerAuth
// @Router       /admin/content/{id} [put]
func (h *ContentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid content ID")
		return
	}
	var req dto.UpdateContentRequest
	if decodeErr := httputil.DecodeJSON(r, &req); decodeErr != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}

	content, appErr := h.contentService.Update(r.Context(), id, contentuc.UpdateContentInput{
		Title: req.Title, Description: req.Description, Synopsis: req.Synopsis,
		Rating: req.Rating, ReleaseYear: req.ReleaseYear, PosterURL: req.PosterURL,
		BackdropURL: req.BackdropURL, TrailerURL: req.TrailerURL,
		Director: req.Director, IsFeatured: req.IsFeatured,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.ContentToDetailResponse(content))
}

// Delete godoc
// @Summary      Delete content (admin)
// @Tags         admin
// @Param        id path string true "Content ID"
// @Success      204
// @Security     BearerAuth
// @Router       /admin/content/{id} [delete]
func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid content ID")
		return
	}
	if appErr := h.contentService.Delete(r.Context(), id); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Publish godoc
// @Summary      Publish content (admin)
// @Tags         admin
// @Param        id path string true "Content ID"
// @Success      200 {object} httputil.Response{data=dto.ContentDetailResponse}
// @Security     BearerAuth
// @Router       /admin/content/{id}/publish [post]
func (h *ContentHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid content ID")
		return
	}
	content, appErr := h.contentService.Publish(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.ContentToDetailResponse(content))
}

// Archive godoc
// @Summary      Archive content (admin)
// @Tags         admin
// @Param        id path string true "Content ID"
// @Success      200 {object} httputil.Response{data=dto.ContentDetailResponse}
// @Security     BearerAuth
// @Router       /admin/content/{id}/archive [post]
func (h *ContentHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid content ID")
		return
	}
	content, appErr := h.contentService.Archive(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.ContentToDetailResponse(content))
}
