package valueobject

import "fmt"

// Duration represents a video duration in seconds.
type Duration struct {
	Seconds int64
}

// NewDuration creates a Duration from seconds.
func NewDuration(seconds int64) Duration {
	if seconds < 0 {
		seconds = 0
	}
	return Duration{Seconds: seconds}
}

// Minutes returns the duration in minutes.
func (d Duration) Minutes() float64 {
	return float64(d.Seconds) / 60.0
}

// String returns a human-readable representation (e.g., "1h 30m").
func (d Duration) String() string {
	hours := d.Seconds / 3600
	minutes := (d.Seconds % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// IsZero reports whether the duration is zero.
func (d Duration) IsZero() bool {
	return d.Seconds == 0
}
