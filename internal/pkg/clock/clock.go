package clock

import "time"

// Clock abstracts time operations for testability.
type Clock interface {
	Now() time.Time
}

// RealClock uses the actual system time.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// MockClock returns a fixed time, useful in tests.
type MockClock struct {
	FixedTime time.Time
}

// Now returns the fixed time.
func (m MockClock) Now() time.Time {
	return m.FixedTime
}
