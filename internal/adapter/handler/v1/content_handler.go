package v1

import (
	"net/http"
	"strconv"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	contentuc "github.com/zenfulcode/zencial/internal/usecase/content"
)

// contentFilterCfg defines which URL query params are accepted for content filtering.
var contentFilterCfg = filter.Config{
	Columns: map[string]filter.ColumnDef{
		"type":         {DBColumn: "c.type", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn}, Type: filter.TypeString},
		"rating":       {DBColumn: "c.rating", AllowedOps: []filter.Op{filter.OpEq, filter.OpIn}, Type: filter.TypeString},
		"release_year": {DBColumn: "c.release_year", AllowedOps: []filter.Op{filter.OpEq, filter.OpGte, filter.OpLte}, Type: filter.TypeInt},
		"title":        {DBColumn: "c.title", AllowedOps: []filter.Op{filter.OpLike}, Type: filter.TypeString},
		"is_featured":  {DBColumn: "c.is_featured", AllowedOps: []filter.Op{filter.OpEq}, Type: filter.TypeBool},
		"director":     {DBColumn: "c.director", AllowedOps: []filter.Op{filter.OpLike, filter.OpEq}, Type: filter.TypeString},
	},
	SortColumns: map[string]filter.SortDef{
		"title":        {DBColumn: "c.title"},
		"release_date": {DBColumn: "c.release_year"},
		"created_at":   {DBColumn: "c.created_at"},
	},
	DefaultSort: "c.created_at DESC",
}

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
// @Description  List and filter content with pagination. Supports filters: type, rating, release_year[gte], release_year[lte], title[like], is_featured, director[like].
// @Tags         content
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        type query string false "Content type (comma-separated for multiple)" Enums(film, series)
// @Param        rating query string false "Rating (comma-separated for multiple)" Enums(G, PG, PG13, R, NC17)
// @Param        release_year[gte] query int false "Minimum release year"
// @Param        release_year[lte] query int false "Maximum release year"
// @Param        title[like] query string false "Title search (ILIKE)"
// @Param        is_featured query bool false "Featured content only"
// @Param        director[like] query string false "Director search (ILIKE)"
// @Param        q query string false "Free-text search across title and description"
// @Param        sort_by query string false "Sort field" Enums(title, release_date, created_at)
// @Param        sort_order query string false "Sort direction" Enums(asc, desc)
// @Success      200 {object} httputil.Response{data=[]dto.ContentListResponse}
// @Security     BearerAuth
// @Router       /content [get]
func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, contentFilterCfg)
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", err.Error())
		return
	}

	searchQuery := r.URL.Query().Get("q")

	contents, total, appErr := h.contentService.List(r.Context(), fs, searchQuery)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, mapper.ContentsToListResponse(contents), pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total))
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
	fs, err := filter.FromRequest(r, contentFilterCfg)
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", err.Error())
		return
	}

	searchQuery := r.URL.Query().Get("q")

	contents, total, appErr := h.contentService.Search(r.Context(), fs, searchQuery)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, mapper.ContentsToListResponse(contents), pagination.NewMeta(fs.Pagination.Page, fs.Pagination.PerPage, total))
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
