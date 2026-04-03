package dto

// GenreTranslationRequest represents a translation in create/update requests.
type GenreTranslationRequest struct {
	LanguageCode string `json:"language_code" validate:"required,min=2,max=5" example:"en"`
	Name         string `json:"name" validate:"required,min=1,max=255" example:"Action"`
	Description  string `json:"description" validate:"max=2000" example:"Action-packed movies and series"`
}

// CreateGenreRequest represents a genre creation request.
type CreateGenreRequest struct {
	Slug         string                    `json:"slug" validate:"required,min=1,max=255" example:"action"`
	Translations []GenreTranslationRequest `json:"translations" validate:"required,min=1,dive"`
}

// UpdateGenreRequest represents a genre update request.
type UpdateGenreRequest struct {
	Slug         *string                   `json:"slug,omitempty" validate:"omitempty,min=1,max=255" example:"action"`
	Translations []GenreTranslationRequest `json:"translations" validate:"required,min=1,dive"`
}

// GenreTranslationResponse represents a translation in API responses.
type GenreTranslationResponse struct {
	LanguageCode string `json:"language_code" example:"en"`
	Name         string `json:"name" example:"Action"`
	Description  string `json:"description" example:"Action-packed movies and series"`
}

// GenreResponse represents a genre in API responses.
type GenreResponse struct {
	ID           string                     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Slug         string                     `json:"slug" example:"action"`
	Translations []GenreTranslationResponse `json:"translations"`
	CreatedAt    string                     `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt    string                     `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// BulkCreateGenreRequest represents a bulk genre creation request.
type BulkCreateGenreRequest struct {
	Genres []CreateGenreRequest `json:"genres" validate:"required,min=1,dive"`
}

// BulkGenreIDsRequest represents a bulk action request with a list of genre IDs.
type BulkGenreIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid"`
}

// BulkCreateGenreFailureResponse represents a single failure in a bulk create operation.
type BulkCreateGenreFailureResponse struct {
	Slug  string `json:"slug"`
	Error string `json:"error"`
}

// BulkCreateGenreResultResponse represents the result of a bulk create operation.
type BulkCreateGenreResultResponse struct {
	Succeeded []GenreResponse                  `json:"succeeded"`
	Failed    []BulkCreateGenreFailureResponse `json:"failed"`
}

// BulkDeleteGenreFailureResponse represents a single failure in a bulk delete operation.
type BulkDeleteGenreFailureResponse struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BulkDeleteGenreResultResponse represents the result of a bulk delete operation.
type BulkDeleteGenreResultResponse struct {
	Succeeded []string                         `json:"succeeded"`
	Failed    []BulkDeleteGenreFailureResponse `json:"failed"`
}
