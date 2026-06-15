package analytics

import (
	"time"

	"github.com/zenfulcode/zencial/internal/domain/repository"
)

// Deltas holds period-over-period changes. Count metrics are relative percent
// changes; rate metrics are percentage-point differences.
type Deltas struct {
	ViewsPct             float64
	WatchTimePct         float64
	UniqueViewersPct     float64
	AvgPercentWatchedPts float64
	FinishRatePts        float64
}

// computeDeltas compares the current window against the previous one. It
// returns nil when there is no meaningful baseline (no previous window, or a
// previous window with zero views).
func computeDeltas(cur, prev *repository.PlaybackTotals) *Deltas {
	if cur == nil || prev == nil || prev.Views == 0 {
		return nil
	}
	return &Deltas{
		ViewsPct:             pctChange(cur.Views, prev.Views),
		WatchTimePct:         pctChange(cur.WatchedSeconds, prev.WatchedSeconds),
		UniqueViewersPct:     pctChange(cur.UniqueViewers, prev.UniqueViewers),
		AvgPercentWatchedPts: cur.AvgPercentWatched - prev.AvgPercentWatched,
		FinishRatePts:        cur.FinishRate - prev.FinishRate,
	}
}

func pctChange(cur, prev int64) float64 {
	if prev == 0 {
		return 0
	}
	return (float64(cur) - float64(prev)) / float64(prev) * 100
}

// fillDailyGaps returns one entry per UTC day in [from, to], inserting
// zero-valued days where the series has no data. Callers should skip this for
// unbounded ("all") ranges.
func fillDailyGaps(series []repository.DailyStat, from, to time.Time) []repository.DailyStat {
	byDay := make(map[string]repository.DailyStat, len(series))
	for _, d := range series {
		byDay[d.Day.UTC().Format("2006-01-02")] = d
	}

	start := truncateToDay(from)
	end := truncateToDay(to)

	filled := make([]repository.DailyStat, 0, int(end.Sub(start).Hours()/24)+1)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		if d, ok := byDay[key]; ok {
			d.Day = day
			filled = append(filled, d)
			continue
		}
		filled = append(filled, repository.DailyStat{Day: day})
	}
	return filled
}

func truncateToDay(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
