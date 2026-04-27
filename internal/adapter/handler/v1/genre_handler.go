package v1

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/mapper"
	"github.com/zenfulcode/zencial/internal/infrastructure/persistence/postgres"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
	"github.com/zenfulcode/zencial/internal/pkg/filter"
	"github.com/zenfulcode/zencial/internal/pkg/httputil"
	"github.com/zenfulcode/zencial/internal/pkg/validator"
	genreuc "github.com/zenfulcode/zencial/internal/usecase/genre"
)

// GenreHandler handles genre HTTP requests.
type GenreHandler struct {
	genreService *genreuc.Service
	validator    *validator.Validator
}

// NewGenreHandler creates a new GenreHandler.
func NewGenreHandler(genreService *genreuc.Service) *GenreHandler {
	return &GenreHandler{
		genreService: genreService,
		validator:    validator.New(),
	}
}

// Create godoc
// @Summary      Create genre
// @Description  Create a new genre with translations (admin only)
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateGenreRequest true "Genre data"
// @Success      201 {object} httputil.Response{data=dto.GenreResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /genres [post]
func (h *GenreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGenreRequest
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

	translations := make([]genreuc.TranslationInput, len(req.Translations))
	for i, t := range req.Translations {
		translations[i] = genreuc.TranslationInput{
			LanguageCode: t.LanguageCode,
			Name:         t.Name,
			Description:  t.Description,
		}
	}

	genre, appErr := h.genreService.Create(r.Context(), genreuc.CreateInput{
		Slug:         req.Slug,
		Translations: translations,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusCreated, mapper.GenreToResponse(genre))
}

// GetByID godoc
// @Summary      Get genre by ID
// @Description  Return a single genre by its UUID
// @Tags         genres
// @Produce      json
// @Param        id path string true "Genre ID" format(uuid)
// @Success      200 {object} httputil.Response{data=dto.GenreResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /genres/{id} [get]
func (h *GenreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid genre ID")
		return
	}

	genre, appErr := h.genreService.GetByID(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.GenreToResponse(genre))
}

// List godoc
// @Summary      List genres
// @Description  Return a paginated list of genres with optional filtering and sorting
// @Tags         genres
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(20)
// @Param        sort query string false "Sort field (e.g. created_at,-name)"
// @Success      200 {object} httputil.Response{data=[]dto.GenreResponse,meta=httputil.Meta}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Router       /genres [get]
func (h *GenreHandler) List(w http.ResponseWriter, r *http.Request) {
	fs, err := filter.FromRequest(r, postgres.GenreFilterConfig())
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	genres, total, appErr := h.genreService.List(r.Context(), &fs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.SuccessWithMeta(w, mapper.GenresToResponse(genres), &httputil.Meta{
		Page:       fs.Pagination.Page,
		PerPage:    fs.Pagination.PerPage,
		Total:      total,
		TotalPages: fs.Pagination.TotalPages(total),
	})
}

// Update godoc
// @Summary      Update genre
// @Description  Update an existing genre's slug and translations (admin only)
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        id path string true "Genre ID" format(uuid)
// @Param        body body dto.UpdateGenreRequest true "Updated genre data"
// @Success      200 {object} httputil.Response{data=dto.GenreResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      409 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /genres/{id} [put]
func (h *GenreHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid genre ID")
		return
	}

	var req dto.UpdateGenreRequest
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

	translations := make([]genreuc.TranslationInput, len(req.Translations))
	for i, t := range req.Translations {
		translations[i] = genreuc.TranslationInput{
			LanguageCode: t.LanguageCode,
			Name:         t.Name,
			Description:  t.Description,
		}
	}

	genre, appErr := h.genreService.Update(r.Context(), genreuc.UpdateInput{
		ID:           id,
		Slug:         req.Slug,
		Translations: translations,
	})
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.GenreToResponse(genre))
}

// Delete godoc
// @Summary      Delete genre
// @Description  Remove a genre by its UUID (admin only)
// @Tags         genres
// @Param        id path string true "Genre ID" format(uuid)
// @Success      204
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /genres/{id} [delete]
func (h *GenreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.URLParamUUID(r, "id")
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, "invalid genre ID")
		return
	}

	appErr := h.genreService.Delete(r.Context(), id)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BulkCreate godoc
// @Summary      Bulk create genres
// @Description  Create multiple genres in a single request (admin only). Returns succeeded and failed entries.
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkCreateGenreRequest true "Genres to create"
// @Success      200 {object} httputil.Response{data=dto.BulkCreateGenreResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/genres/bulk-create [post]
func (h *GenreHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkCreateGenreRequest
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

	inputs := make([]genreuc.CreateInput, len(req.Genres))
	for i, g := range req.Genres {
		translations := make([]genreuc.TranslationInput, len(g.Translations))
		for j, t := range g.Translations {
			translations[j] = genreuc.TranslationInput{
				LanguageCode: t.LanguageCode,
				Name:         t.Name,
				Description:  t.Description,
			}
		}
		inputs[i] = genreuc.CreateInput{
			Slug:         g.Slug,
			Translations: translations,
		}
	}

	result, appErr := h.genreService.BulkCreate(r.Context(), inputs)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkCreateGenreResultToResponse(result))
}

// BulkDelete godoc
// @Summary      Bulk delete genres
// @Description  Remove multiple genres in a single request (admin only). Returns succeeded and failed entries.
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        body body dto.BulkGenreIDsRequest true "Genre IDs to delete"
// @Success      200 {object} httputil.Response{data=dto.BulkDeleteGenreResultResponse}
// @Failure      400 {object} httputil.ErrorResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      500 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/genres/bulk-delete [post]
func (h *GenreHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkGenreIDsRequest
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

	ids, err := parseGenreUUIDs(req.IDs)
	if err != nil {
		httputil.BadRequest(w, apperror.CodeBadRequest, err.Error())
		return
	}

	result, appErr := h.genreService.BulkDelete(r.Context(), ids)
	if appErr != nil {
		httputil.Error(w, appErr)
		return
	}

	httputil.Success(w, http.StatusOK, mapper.BulkDeleteGenreResultToResponse(result))
}

// parseGenreUUIDs converts a slice of string IDs to uuid.UUID.
func parseGenreUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", id)
		}
		result[i] = parsed
	}
	return result, nil
}
