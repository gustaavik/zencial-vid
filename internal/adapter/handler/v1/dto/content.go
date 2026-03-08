package dto

// ContentListResponse represents a content item in list views.
// Covers Film, Video, and Series rows.
type ContentListResponse struct {
	ID          string         `json:"id"`
	Type        string         `json:"type" example:"film"`
	Title       string         `json:"title" example:"The Matrix"`
	Slug        string         `json:"slug" example:"the-matrix"`
	Description string         `json:"description"`
	Status      string         `json:"status" example:"published"`
	Rating      string         `json:"rating,omitempty" example:"R"`
	PosterURL   string         `json:"poster_url,omitempty"`
	Genre       *GenreResponse `json:"genre,omitempty"`
	IsFeatured  bool           `json:"is_featured"`
	IsFree      bool           `json:"is_free"`
	CreatorName string         `json:"creator_name,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// ContentDetailResponse represents full content details.
// Film-specific fields (BackdropURL, TrailerURL, Director, Cast, ReleaseYear)
// are omitted (zero/nil) for Video and Series responses.
type ContentDetailResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Title        string               `json:"title"`
	Slug         string               `json:"slug"`
	Description  string               `json:"description"`
	Synopsis     string               `json:"synopsis,omitempty"`
	Rating       string               `json:"rating,omitempty"`
	PosterURL    string               `json:"poster_url,omitempty"`
	BackdropURL  string               `json:"backdrop_url,omitempty"`
	TrailerURL   string               `json:"trailer_url,omitempty"`
	Director     string               `json:"director,omitempty"`
	ReleaseYear  int                  `json:"release_year,omitempty"`
	IsFeatured   bool                 `json:"is_featured"`
	IsFree       bool                 `json:"is_free"`
	Genre        *GenreResponse       `json:"genre,omitempty"`
	Cast         []CastMemberResponse `json:"cast,omitempty"`
	Asset        *VideoAssetResponse  `json:"asset,omitempty"`
	TotalSeasons int                  `json:"total_seasons,omitempty"`
	CreatorName  string               `json:"creator_name,omitempty"`
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
}

// VideoAssetResponse represents a video asset.
type VideoAssetResponse struct {
	ID         string                   `json:"id"`
	StorageKey string                   `json:"storage_key"`
	Status     string                   `json:"status"`
	Qualities  []VideoRenditionResponse `json:"qualities"`
}

// VideoRenditionResponse represents a single quality rendition.
type VideoRenditionResponse struct {
	Quality    string `json:"quality"`
	URL        string `json:"url"`
	Bitrate    int    `json:"bitrate"`
	Resolution string `json:"resolution"`
}

// AttachVideoAssetRequest represents a request to attach a video asset to content.
type AttachVideoAssetRequest struct {
	StorageKey string `json:"storage_key" validate:"required"`
}

// SeasonResponse represents a season.
type SeasonResponse struct {
	ID       string `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Episodes int    `json:"episode_count"`
}

// EpisodeResponse represents an episode.
type EpisodeResponse struct {
	ID              string  `json:"id"`
	Number          int     `json:"number"`
	Title           string  `json:"title"`
	Synopsis        string  `json:"synopsis"`
	DurationMinutes float64 `json:"duration_minutes"`
	AirDate         *string `json:"air_date,omitempty"`
}

// CastMemberResponse represents a cast member.
type CastMemberResponse struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Character string `json:"character,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
}

// CreateFilmRequest represents a film creation request.
type CreateFilmRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required"`
	Synopsis    string `json:"synopsis"`
	Rating      string `json:"rating" validate:"omitempty,oneof=G PG PG13 R NC17"`
	ReleaseYear int    `json:"release_year"`
	PosterURL   string `json:"poster_url" validate:"omitempty,url"`
	BackdropURL string `json:"backdrop_url" validate:"omitempty,url"`
	TrailerURL  string `json:"trailer_url" validate:"omitempty,url"`
	Director    string `json:"director"`
}

// CreateVideoRequest represents a video creation request.
type CreateVideoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required"`
	Synopsis    string `json:"synopsis"`
	Rating      string `json:"rating" validate:"omitempty,oneof=G PG PG13 R NC17"`
	PosterURL   string `json:"poster_url" validate:"omitempty,url"`
	CreatorName string `json:"creator_name" validate:"required"`
}

// CreateSeriesRequest represents a series creation request.
type CreateSeriesRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required"`
	Synopsis    string `json:"synopsis"`
	PosterURL   string `json:"poster_url" validate:"omitempty,url"`
	BackdropURL string `json:"backdrop_url" validate:"omitempty,url"`
	TrailerURL  string `json:"trailer_url" validate:"omitempty,url"`
}

// UpdateContentRequest represents a content update request.
type UpdateContentRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Synopsis    *string `json:"synopsis,omitempty"`
	Rating      *string `json:"rating,omitempty" validate:"omitempty,oneof=G PG PG13 R NC17"`
	ReleaseYear *int    `json:"release_year,omitempty"`
	PosterURL   *string `json:"poster_url,omitempty"`
	BackdropURL *string `json:"backdrop_url,omitempty"`
	TrailerURL  *string `json:"trailer_url,omitempty"`
	Director    *string `json:"director,omitempty"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
	CreatorName *string `json:"creator_name,omitempty"`
}
