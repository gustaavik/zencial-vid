package dto

// SeriesResponse represents a series in API responses.
type SeriesResponse struct {
	ID               string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title            string   `json:"title" example:"My Great Show"`
	Slug             string   `json:"slug" example:"my-great-show-a3f8b2c1"`
	Description      string   `json:"description" example:"A compelling drama series..."`
	Creator          string   `json:"creator" example:"Studio ABC"`
	Status           string   `json:"status" example:"published"`
	CoverImageKey    string   `json:"cover_image_key,omitempty" example:"series/550e8400/cover.jpg"`
	GenreIDs         []string `json:"genre_ids"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" example:"1"`
	CreatedAt        string   `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt        string   `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// CreateSeriesRequest is the body for POST /series.
type CreateSeriesRequest struct {
	Title            string   `json:"title" validate:"required,min=1,max=500" example:"My Great Show"`
	Description      string   `json:"description,omitempty" validate:"omitempty,max=5000"`
	Creator          string   `json:"creator,omitempty" validate:"omitempty,min=3,max=200"`
	CoverImageKey    string   `json:"cover_image_key,omitempty" validate:"omitempty,max=1000"`
	GenreIDs         []string `json:"genre_ids,omitempty" validate:"omitempty,dive,uuid"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" validate:"omitempty,gte=0"`
}

// UpdateSeriesRequest represents a series metadata update.
type UpdateSeriesRequest struct {
	Title            *string  `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description      *string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	Creator          *string  `json:"creator,omitempty" validate:"omitempty,min=3,max=200"`
	CoverImageKey    *string  `json:"cover_image_key,omitempty" validate:"omitempty,max=1000"`
	GenreIDs         []string `json:"genre_ids,omitempty" validate:"omitempty,dive,uuid"`
	MinimumPlanLevel *int     `json:"minimum_plan_level,omitempty" validate:"omitempty,gte=0"`
}

// AddEpisodeRequest links an existing video to a series as a numbered episode.
type AddEpisodeRequest struct {
	VideoID       string `json:"video_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeasonNumber  int    `json:"season_number" validate:"required,gte=1" example:"1"`
	EpisodeNumber int    `json:"episode_number" validate:"required,gte=1" example:"3"`
}

// SeriesWatchProgressResponse represents a user's watch progress for a series.
type SeriesWatchProgressResponse struct {
	SeriesID      string `json:"series_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LastEpisodeID string `json:"last_episode_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UpdatedAt     string `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

// UpdateSeriesWatchProgressRequest sets the last-watched episode for a series.
type UpdateSeriesWatchProgressRequest struct {
	EpisodeID string `json:"episode_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
}
