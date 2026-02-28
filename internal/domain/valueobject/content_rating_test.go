package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentRating_MinAge(t *testing.T) {
	tests := []struct {
		name   string
		rating ContentRating
		want   int
	}{
		{"G requires age 0", RatingG, 0},
		{"PG requires age 0", RatingPG, 0},
		{"PG13 requires age 13", RatingPG13, 13},
		{"R requires age 17", RatingR, 17},
		{"NC17 requires age 18", RatingNC17, 18},
		{"unknown defaults to 0", ContentRating("UNKNOWN"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rating.MinAge())
		})
	}
}

func TestContentRating_AllowedForAge(t *testing.T) {
	tests := []struct {
		name    string
		rating  ContentRating
		age     int
		allowed bool
	}{
		// G rated - all ages
		{"G allowed for age 0", RatingG, 0, true},
		{"G allowed for age 5", RatingG, 5, true},
		{"G allowed for age 18", RatingG, 18, true},

		// PG rated - all ages
		{"PG allowed for age 0", RatingPG, 0, true},
		{"PG allowed for age 10", RatingPG, 10, true},

		// PG13 rated - 13 and older
		{"PG13 not allowed for age 12", RatingPG13, 12, false},
		{"PG13 allowed for age 13", RatingPG13, 13, true},
		{"PG13 allowed for age 14", RatingPG13, 14, true},

		// R rated - 17 and older
		{"R not allowed for age 16", RatingR, 16, false},
		{"R allowed for age 17", RatingR, 17, true},
		{"R allowed for age 18", RatingR, 18, true},

		// NC17 rated - 18 and older
		{"NC17 not allowed for age 17", RatingNC17, 17, false},
		{"NC17 allowed for age 18", RatingNC17, 18, true},
		{"NC17 allowed for age 25", RatingNC17, 25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.allowed, tt.rating.AllowedForAge(tt.age))
		})
	}
}

func TestContentRating_IsUnrestricted(t *testing.T) {
	tests := []struct {
		name         string
		rating       ContentRating
		unrestricted bool
	}{
		{"G is unrestricted", RatingG, true},
		{"PG is unrestricted", RatingPG, true},
		{"PG13 is restricted", RatingPG13, false},
		{"R is restricted", RatingR, false},
		{"NC17 is restricted", RatingNC17, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.unrestricted, tt.rating.IsUnrestricted())
		})
	}
}

func TestContentRating_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		rating ContentRating
		valid  bool
	}{
		{"G is valid", RatingG, true},
		{"PG is valid", RatingPG, true},
		{"PG13 is valid", RatingPG13, true},
		{"R is valid", RatingR, true},
		{"NC17 is valid", RatingNC17, true},
		{"unknown is invalid", ContentRating("UNKNOWN"), false},
		{"empty is invalid", ContentRating(""), false},
		{"lowercase g is invalid", ContentRating("g"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.rating.IsValid())
		})
	}
}

func TestContentRating_String(t *testing.T) {
	tests := []struct {
		name   string
		rating ContentRating
		want   string
	}{
		{"G string", RatingG, "G"},
		{"PG string", RatingPG, "PG"},
		{"PG13 string", RatingPG13, "PG13"},
		{"R string", RatingR, "R"},
		{"NC17 string", RatingNC17, "NC17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rating.String())
		})
	}
}
