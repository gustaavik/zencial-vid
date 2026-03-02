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

