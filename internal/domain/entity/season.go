package entity

import (
	"time"

	"github.com/google/uuid"
)

// ReleaseCadence describes how episodes are released within a season.
type ReleaseCadence string

const (
	CadenceAllAtOnce ReleaseCadence = "all_at_once"
	CadenceWeekly    ReleaseCadence = "weekly"
	CadenceBiWeekly  ReleaseCadence = "bi_weekly"
	CadenceOnDemand  ReleaseCadence = "on_demand"
)

// Season groups episodes of a series into a numbered block with release metadata.
type Season struct {
	ID              uuid.UUID
	SeriesID        uuid.UUID
	SeasonNumber    int
	SeasonTag       string
	PlannedEpisodes int
	AvgRuntimeSecs  int
	ReleaseCadence  ReleaseCadence
	PremiereDate    *time.Time
	CadenceDay      *int // 0=Sunday … 6=Saturday
	Timezone        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewSeason creates a new Season for the given series.
func NewSeason(seriesID uuid.UUID, seasonNumber int) *Season {
	now := time.Now().UTC()
	return &Season{
		ID:             uuid.New(),
		SeriesID:       seriesID,
		SeasonNumber:   seasonNumber,
		ReleaseCadence: CadenceOnDemand,
		Timezone:       "UTC",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
