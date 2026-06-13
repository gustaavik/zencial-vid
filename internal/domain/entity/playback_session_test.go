package entity

import "testing"

func TestNormalizePlaybackSource(t *testing.T) {
	tests := []struct {
		raw  string
		want PlaybackSource
	}{
		{"home", PlaybackSourceHome},
		{"search", PlaybackSourceSearch},
		{"series", PlaybackSourceSeries},
		{"profile", PlaybackSourceProfile},
		{"library", PlaybackSourceLibrary},
		{"direct", PlaybackSourceDirect},
		{"external", PlaybackSourceExternal},
		{"other", PlaybackSourceOther},
		{"", PlaybackSourceOther},
		{"HOME", PlaybackSourceOther},
		{"tiktok", PlaybackSourceOther},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := NormalizePlaybackSource(tt.raw); got != tt.want {
				t.Errorf("NormalizePlaybackSource(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizePlaybackPlatform(t *testing.T) {
	tests := []struct {
		raw  string
		want PlaybackPlatform
	}{
		{"web", PlaybackPlatformWeb},
		{"ios", PlaybackPlatformIOS},
		{"android", PlaybackPlatformAndroid},
		{"tvos", PlaybackPlatformTVOS},
		{"android_tv", PlaybackPlatformAndroidTV},
		{"other", PlaybackPlatformOther},
		{"", PlaybackPlatformOther},
		{"windows", PlaybackPlatformOther},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := NormalizePlaybackPlatform(tt.raw); got != tt.want {
				t.Errorf("NormalizePlaybackPlatform(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestViewThresholdSeconds(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		want     int64
	}{
		{"zero duration floors at 1", 0, 1},
		{"one second video", 1, 1},
		{"10 second video uses half", 10, 5},
		{"59 second video uses half", 59, 29},
		{"60 second video caps at 30", 60, 30},
		{"feature length caps at 30", 7200, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ViewThresholdSeconds(tt.duration); got != tt.want {
				t.Errorf("ViewThresholdSeconds(%d) = %d, want %d", tt.duration, got, tt.want)
			}
		})
	}
}
