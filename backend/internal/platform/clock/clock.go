// Package clock provides a thin abstraction over time.Now() so that
// tests can control what "now" means without monkey-patching globals.
package clock

import "time"

// Clock provides the current time. Use RealClock in production
// and MockClock in tests.
type Clock interface {
	Now() time.Time
}

// RealClock returns the actual current time.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// MockClock returns a fixed time, useful for tests.
type MockClock struct {
	T time.Time
}

func (m MockClock) Now() time.Time { return m.T }
