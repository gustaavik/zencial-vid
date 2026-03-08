package v1

import (
	"net/http"

	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/pagination"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	cataloguc "github.com/zenfulcode/zencial/internal/usecase/catalog"
)

// CatalogHandler handles catalog HTTP requests.
type CatalogHandler struct {
	catalogService *cataloguc.Service
	validator      *validator.Validator
}

// NewCatalogHandler creates a new CatalogHandler.
func NewCatalogHandler(catalogService *cataloguc.Service) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService, validator: validator.New()}
}

// ListGenres godoc
// @Summary      List all genres
// @Tags         catalog
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.GenreResponse}
// @Security     BearerAuth
// @Router       /genres [get]
func (h *CatalogHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	genres, appErr := h.catalogService.ListGenres(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.GenresToResponse(genres))
}

// ContentByGenre godoc
// @Summary      List content by genre
// @Tags         catalog
// @Produce      json
// @Param        slug path string true "Genre slug"
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Success      200 {object} httputil.Response{data=[]dto.ContentListResponse}
// @Security     BearerAuth
// @Router       /genres/{slug}/content [get]
func (h *CatalogHandler) ContentByGenre(w http.ResponseWriter, r *http.Request) {
	slug := httputil.URLParam(r, "slug")
	page := httputil.QueryInt(r, "page", 1)
	perPage := httputil.QueryInt(r, "per_page", 20)

	contents, total, appErr := h.catalogService.ContentByGenre(r.Context(), slug, page, perPage)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.SuccessWithMeta(w, mapper.ContentSummariesToListResponse(contents), pagination.NewMeta(page, perPage, total))
}

// ListCategories godoc
// @Summary      List all categories
// @Tags         catalog
// @Produce      json
// @Success      200 {object} httputil.Response{data=[]dto.CategoryResponse}
// @Security     BearerAuth
// @Router       /categories [get]
func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, appErr := h.catalogService.ListCategories(r.Context())
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.CategoriesToResponse(categories))
}

// CreateGenre godoc
// @Summary      Create genre (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateGenreRequest true "Genre data"
// @Success      201 {object} httputil.Response{data=dto.GenreResponse}
// @Security     BearerAuth
// @Router       /admin/genres [post]
func (h *CatalogHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGenreRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		httputil.ErrorWithDetails(w, apperror.BadRequest("VALIDATION_FAILED", "validation failed", nil), errors)
		return
	}
	genre, appErr := h.catalogService.CreateGenre(r.Context(), req.Name)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusCreated, mapper.GenreToResponse(genre))
}

// UpdateGenre godoc
// @Summary      Update genre (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path string true "Genre ID"
// @Param        body body dto.UpdateGenreRequest true "Genre data"
// @Success      200 {object} httputil.Response{data=dto.GenreResponse}
// @Security     BearerAuth
// @Router       /admin/genres/{id} [put]
func (h *CatalogHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid genre ID")
		return
	}
	var req dto.UpdateGenreRequest
	if decodeErr := httputil.DecodeJSON(r, &req); decodeErr != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid request body")
		return
	}
	genre, appErr := h.catalogService.UpdateGenre(r.Context(), id, req.Name)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	httputil.Success(w, http.StatusOK, mapper.GenreToResponse(genre))
}

// DeleteGenre godoc
// @Summary      Delete genre (admin)
// @Tags         admin
// @Param        id path string true "Genre ID"
// @Success      204
// @Security     BearerAuth
// @Router       /admin/genres/{id} [delete]
func (h *CatalogHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, "BAD_REQUEST", "invalid genre ID")
		return
	}
	if appErr := h.catalogService.DeleteGenre(r.Context(), id); appErr != nil {
		httputil.Error(w, appErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
