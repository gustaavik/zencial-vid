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

// Create creates a new genre.
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

// GetByID returns a genre by ID.
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

// List returns a paginated list of genres.
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

// Update updates an existing genre.
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

// Delete removes a genre.
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

// BulkCreate creates multiple genres.
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

// BulkDelete removes multiple genres.
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
