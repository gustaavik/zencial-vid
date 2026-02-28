package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDuration(t *testing.T) {
	tests := []struct {
		name        string
		seconds     int64
		wantSeconds int64
	}{
		{"positive seconds", 3600, 3600},
		{"zero seconds", 0, 0},
		{"negative seconds clamped to zero", -100, 0},
		{"large negative clamped to zero", -999999, 0},
		{"one second", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDuration(tt.seconds)
			assert.Equal(t, tt.wantSeconds, d.Seconds)
		})
	}
}

func TestDuration_Minutes(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    float64
	}{
		{"60 seconds is 1 minute", 60, 1.0},
		{"90 seconds is 1.5 minutes", 90, 1.5},
		{"0 seconds is 0 minutes", 0, 0.0},
		{"3600 seconds is 60 minutes", 3600, 60.0},
		{"30 seconds is 0.5 minutes", 30, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDuration(tt.seconds)
			assert.InDelta(t, tt.want, d.Minutes(), 0.001)
		})
	}
}

func TestDuration_String(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    string
	}{
		{"1 hour 30 minutes", 5400, "1h 30m"},
		{"2 hours 0 minutes", 7200, "2h 0m"},
		{"30 minutes only", 1800, "30m"},
		{"0 minutes", 0, "0m"},
		{"1 minute", 60, "1m"},
		{"59 seconds shows 0m", 59, "0m"},
		{"exactly 1 hour", 3600, "1h 0m"},
		{"1 hour 1 minute", 3660, "1h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDuration(tt.seconds)
			assert.Equal(t, tt.want, d.String())
		})
	}
}

func TestDuration_IsZero(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    bool
	}{
		{"zero duration", 0, true},
		{"non-zero duration", 100, false},
		{"negative becomes zero", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDuration(tt.seconds)
			assert.Equal(t, tt.want, d.IsZero())
		})
	}
}
