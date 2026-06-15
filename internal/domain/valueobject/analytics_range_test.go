package valueobject

import (
	"errors"
	"testing"
	"time"
)

func TestNewAnalyticsRange(t *testing.T) {
	now := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		raw      string
		wantKey  string
		wantFrom time.Time
		wantPrev bool
		wantErr  bool
	}{
		{
			name:     "7d",
			raw:      "7d",
			wantKey:  "7d",
			wantFrom: now.AddDate(0, 0, -7),
			wantPrev: true,
		},
		{
			name:     "30d",
			raw:      "30d",
			wantKey:  "30d",
			wantFrom: now.AddDate(0, 0, -30),
			wantPrev: true,
		},
		{
			name:     "90d",
			raw:      "90d",
			wantKey:  "90d",
			wantFrom: now.AddDate(0, 0, -90),
			wantPrev: true,
		},
		{
			name:     "empty defaults to 30d",
			raw:      "",
			wantKey:  "30d",
			wantFrom: now.AddDate(0, 0, -30),
			wantPrev: true,
		},
		{
			name:     "all has no previous window",
			raw:      "all",
			wantKey:  "all",
			wantFrom: time.Unix(0, 0).UTC(),
			wantPrev: false,
		},
		{
			name:    "invalid key",
			raw:     "14d",
			wantErr: true,
		},
		{
			name:    "garbage",
			raw:     "yesterday",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAnalyticsRange(tt.raw, now)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidAnalyticsRange) {
					t.Fatalf("expected ErrInvalidAnalyticsRange, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", got.Key, tt.wantKey)
			}
			if !got.From.Equal(tt.wantFrom) {
				t.Errorf("From = %v, want %v", got.From, tt.wantFrom)
			}
			if !got.To.Equal(now) {
				t.Errorf("To = %v, want %v", got.To, now)
			}
			if got.HasPrev != tt.wantPrev {
				t.Errorf("HasPrev = %v, want %v", got.HasPrev, tt.wantPrev)
			}
			if tt.wantPrev {
				if !got.PrevTo.Equal(got.From) {
					t.Errorf("PrevTo = %v, want %v (contiguous windows)", got.PrevTo, got.From)
				}
				wantPrevFrom := got.From.Add(-got.To.Sub(got.From))
				if !got.PrevFrom.Equal(wantPrevFrom) {
					t.Errorf("PrevFrom = %v, want %v (equal-length windows)", got.PrevFrom, wantPrevFrom)
				}
			}
		})
	}
}
