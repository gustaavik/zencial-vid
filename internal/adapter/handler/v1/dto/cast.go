package dto

// CastResponse represents a cast member in API responses.
type CastResponse struct {
	ID         string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	VideoID    string `json:"video_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name       string `json:"name" example:"Jane Doe"`
	Role       string `json:"role" example:"actor"`
	SortOrder  int    `json:"sort_order" example:"0"`
	PictureURL string `json:"picture_url,omitempty" example:"https://cdn.example.com/cast/picture.jpg"`
	CreatedAt  string `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CreateCastRequest is the body for adding a cast member.
type CreateCastRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=255" example:"Jane Doe"`
	Role      string `json:"role" validate:"required,min=1,max=100" example:"actor"`
	SortOrder int    `json:"sort_order" validate:"gte=0" example:"0"`
}

// UpdateCastRequest is the body for updating a cast member.
type UpdateCastRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255" example:"Jane Doe"`
	Role      *string `json:"role,omitempty" validate:"omitempty,min=1,max=100" example:"director"`
	SortOrder *int    `json:"sort_order,omitempty" validate:"omitempty,gte=0" example:"1"`
}
