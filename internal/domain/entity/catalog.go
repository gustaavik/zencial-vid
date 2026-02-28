package entity

import "github.com/google/uuid"

// Genre represents a content genre.
type Genre struct {
	ID   uuid.UUID
	Name string
	Slug string
}

// Category represents a content category.
type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	ParentID    *uuid.UUID
	SortOrder   int
}

// Tag represents a content tag for fine-grained classification.
type Tag struct {
	ID   uuid.UUID
	Name string
	Slug string
}

// SearchCriteria encapsulates catalog search parameters.
type SearchCriteria struct {
	Query     string
	GenreIDs  []uuid.UUID
	Type      *ContentType
	Rating    *string
	YearFrom  *int
	YearTo    *int
	SortBy    string // "relevance", "release_date", "title", "popularity"
	SortOrder string // "asc", "desc"
	Page      int
	PerPage   int
}
