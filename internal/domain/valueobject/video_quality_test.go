package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVideoQuality_Rank(t *testing.T) {
	tests := []struct {
		name    string
		quality VideoQuality
		want    int
	}{
		{"SD rank", QualitySD, 1},
		{"HD rank", QualityHD, 2},
		{"FHD rank", QualityFHD, 3},
		{"UHD rank", QualityUHD, 4},
		{"unknown quality rank", VideoQuality("UNKNOWN"), 0},
		{"empty quality rank", VideoQuality(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.quality.Rank())
		})
	}
}

func TestVideoQuality_IsAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		quality  VideoQuality
		other    VideoQuality
		expected bool
	}{
		{"UHD is at least SD", QualityUHD, QualitySD, true},
		{"UHD is at least HD", QualityUHD, QualityHD, true},
		{"UHD is at least FHD", QualityUHD, QualityFHD, true},
		{"UHD is at least UHD", QualityUHD, QualityUHD, true},
		{"FHD is at least SD", QualityFHD, QualitySD, true},
		{"FHD is at least HD", QualityFHD, QualityHD, true},
		{"FHD is at least FHD", QualityFHD, QualityFHD, true},
		{"FHD is not at least UHD", QualityFHD, QualityUHD, false},
		{"HD is at least SD", QualityHD, QualitySD, true},
		{"HD is at least HD", QualityHD, QualityHD, true},
		{"HD is not at least FHD", QualityHD, QualityFHD, false},
		{"SD is at least SD", QualitySD, QualitySD, true},
		{"SD is not at least HD", QualitySD, QualityHD, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.quality.IsAtLeast(tt.other))
		})
	}
}

func TestVideoQuality_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		quality VideoQuality
		want    bool
	}{
		{"SD is valid", QualitySD, true},
		{"HD is valid", QualityHD, true},
		{"FHD is valid", QualityFHD, true},
		{"UHD is valid", QualityUHD, true},
		{"unknown is invalid", VideoQuality("UNKNOWN"), false},
		{"empty is invalid", VideoQuality(""), false},
		{"lowercase sd is invalid", VideoQuality("sd"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.quality.IsValid())
		})
	}
}

func TestVideoQuality_String(t *testing.T) {
	tests := []struct {
		name    string
		quality VideoQuality
		want    string
	}{
		{"SD string", QualitySD, "SD"},
		{"HD string", QualityHD, "HD"},
		{"FHD string", QualityFHD, "FHD"},
		{"UHD string", QualityUHD, "UHD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.quality.String())
		})
	}
}
