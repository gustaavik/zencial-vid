package valueobject

import (
	"errors"
	"time"
)

// ErrInvalidAnalyticsRange is returned when a range key is not one of 7d, 30d, 90d, all.
var ErrInvalidAnalyticsRange = errors.New("invalid analytics range")

// DefaultAnalyticsRange is the range used when no range is requested.
const DefaultAnalyticsRange = "30d"

// AnalyticsRange is a resolved reporting window plus the equal-length window
// immediately preceding it, used for period-over-period deltas. The "all"
// range has no previous window (HasPrev is false).
type AnalyticsRange struct {
	Key      string
	From     time.Time
	To       time.Time
	PrevFrom time.Time
	PrevTo   time.Time
	HasPrev  bool
}

// NewAnalyticsRange resolves a range key ("7d", "30d", "90d", "all"; empty
// defaults to 30d) against the given current time. The window is half-open:
// [From, To).
func NewAnalyticsRange(raw string, now time.Time) (AnalyticsRange, error) {
	if raw == "" {
		raw = DefaultAnalyticsRange
	}

	var days int
	switch raw {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	case "all":
		return AnalyticsRange{Key: raw, From: time.Unix(0, 0).UTC(), To: now}, nil
	default:
		return AnalyticsRange{}, ErrInvalidAnalyticsRange
	}

	window := time.Duration(days) * 24 * time.Hour
	from := now.Add(-window)
	return AnalyticsRange{
		Key:      raw,
		From:     from,
		To:       now,
		PrevFrom: from.Add(-window),
		PrevTo:   from,
		HasPrev:  true,
	}, nil
}
