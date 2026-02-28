package dto

// GenreResponse represents a genre in API responses.
type GenreResponse struct {
	ID   string `json:"id"`
	Name string `json:"name" example:"Action"`
	Slug string `json:"slug" example:"action"`
}

// CategoryResponse represents a category in API responses.
type CategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// CreateGenreRequest represents a genre creation request.
type CreateGenreRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateGenreRequest represents a genre update request.
type UpdateGenreRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
