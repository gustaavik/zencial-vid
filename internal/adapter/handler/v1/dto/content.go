package dto

// ContentListResponse represents a content item in list views.
type ContentListResponse struct {
	ID          string   `json:"id"`
	Type        string   `json:"type" example:"film"`
	Title       string   `json:"title" example:"The Matrix"`
	Slug        string   `json:"slug" example:"the-matrix"`
	Description string   `json:"description"`
	Rating      string   `json:"rating" example:"R"`
	ReleaseYear int      `json:"release_year" example:"1999"`
	PosterURL   string   `json:"poster_url"`
	Genres      []string `json:"genres"`
	IsFeatured  bool     `json:"is_featured"`
}

// ContentDetailResponse represents full content details.
type ContentDetailResponse struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	Title       string               `json:"title"`
	Slug        string               `json:"slug"`
	Description string               `json:"description"`
	Synopsis    string               `json:"synopsis"`
	Rating      string               `json:"rating"`
	ReleaseYear int                  `json:"release_year"`
	PosterURL   string               `json:"poster_url"`
	BackdropURL string               `json:"backdrop_url"`
	TrailerURL  string               `json:"trailer_url"`
	Director    string               `json:"director"`
	IsFeatured  bool                 `json:"is_featured"`
	Genres      []GenreResponse      `json:"genres"`
	Cast        []CastMemberResponse `json:"cast"`
	Film        *FilmResponse        `json:"film,omitempty"`
	Series      *SeriesResponse      `json:"series,omitempty"`
	CreatedAt   string               `json:"created_at"`
}

// FilmResponse holds film-specific data.
type FilmResponse struct {
	DurationMinutes float64 `json:"duration_minutes"`
}

// SeriesResponse holds series-specific data.
type SeriesResponse struct {
	TotalSeasons int `json:"total_seasons"`
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

// CreateContentRequest represents a content creation request.
type CreateContentRequest struct {
	Type        string `json:"type" validate:"required,oneof=film series"`
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required"`
	Synopsis    string `json:"synopsis"`
	Rating      string `json:"rating" validate:"required,oneof=G PG PG13 R NC17"`
	ReleaseYear int    `json:"release_year" validate:"required"`
	PosterURL   string `json:"poster_url" validate:"omitempty,url"`
	BackdropURL string `json:"backdrop_url" validate:"omitempty,url"`
	TrailerURL  string `json:"trailer_url" validate:"omitempty,url"`
	Director    string `json:"director"`
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
}
