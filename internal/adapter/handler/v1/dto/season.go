package dto

// SeasonResponse represents a season in API responses.
type SeasonResponse struct {
	ID              string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeriesID        string  `json:"series_id"`
	SeasonNumber    int     `json:"season_number" example:"1"`
	SeasonTag       string  `json:"season_tag,omitempty" example:"Summer 2026"`
	PlannedEpisodes int     `json:"planned_episodes" example:"6"`
	AvgRuntimeSecs  int     `json:"avg_runtime_secs" example:"2520"`
	ReleaseCadence  string  `json:"release_cadence" example:"weekly"`
	PremiereDate    *string `json:"premiere_date,omitempty" example:"2026-06-06T10:00:00Z"`
	CadenceDay      *int    `json:"cadence_day,omitempty" example:"5"`
	Timezone        string  `json:"timezone" example:"America/New_York"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// CreateSeasonRequest is the body for POST /publisher/series/{id}/seasons.
type CreateSeasonRequest struct {
	SeasonNumber    int     `json:"season_number" validate:"required,gte=1" example:"1"`
	SeasonTag       string  `json:"season_tag,omitempty" validate:"omitempty,max=200"`
	PlannedEpisodes int     `json:"planned_episodes,omitempty" validate:"omitempty,gte=0"`
	AvgRuntimeSecs  int     `json:"avg_runtime_secs,omitempty" validate:"omitempty,gte=0"`
	ReleaseCadence  string  `json:"release_cadence,omitempty" validate:"omitempty,oneof=all_at_once weekly bi_weekly on_demand"`
	PremiereDate    *string `json:"premiere_date,omitempty"`
	CadenceDay      *int    `json:"cadence_day,omitempty" validate:"omitempty,gte=0,lte=6"`
	Timezone        string  `json:"timezone,omitempty" validate:"omitempty,max=100"`
}

// UpdateSeasonRequest is the body for PUT /publisher/series/{id}/seasons/{n}.
type UpdateSeasonRequest struct {
	SeasonTag       *string `json:"season_tag,omitempty" validate:"omitempty,max=200"`
	PlannedEpisodes *int    `json:"planned_episodes,omitempty" validate:"omitempty,gte=0"`
	AvgRuntimeSecs  *int    `json:"avg_runtime_secs,omitempty" validate:"omitempty,gte=0"`
	ReleaseCadence  *string `json:"release_cadence,omitempty" validate:"omitempty,oneof=all_at_once weekly bi_weekly on_demand"`
	PremiereDate    *string `json:"premiere_date,omitempty"`
	CadenceDay      *int    `json:"cadence_day,omitempty" validate:"omitempty,gte=0,lte=6"`
	Timezone        *string `json:"timezone,omitempty" validate:"omitempty,max=100"`
}
