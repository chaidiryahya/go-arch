// Package time provides the time seam: a real implementation backed by the
// standard library and an exported mock for deterministic tests.
//
// The package is named after the standard library package it wraps, so call
// sites read as time.Time and time.MockTime. Inside this package the standard
// library is imported as stdtime to keep the two apart.
package time

import (
	"sync"
	stdtime "time"
)

// Time is the production implementation. It is stateless and its zero value is
// ready to use.
type Time struct{}

// Now returns the current time exactly as time.Now reports it. It deliberately
// does not call Round, UTC, or anything else that would strip the monotonic
// reading, so Now().Sub(start) stays equivalent to time.Since(start) and keeps
// its protection against wall-clock jumps.
func (Time) Now() stdtime.Time {
	return stdtime.Now()
}

// MockTime is a manually controlled clock. Construct it with NewMockTime; the
// zero value is not usable. The mutex makes MockTime safe for concurrent use by
// multiple goroutines.
type MockTime struct {
	mu  sync.Mutex
	now stdtime.Time
}

// NewMockTime returns a MockTime pinned to t.
func NewMockTime(t stdtime.Time) *MockTime {
	return &MockTime{now: t}
}

// Now returns the mock clock's current time. Safe for concurrent use.
func (m *MockTime) Now() stdtime.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

// Set moves the clock to t. Safe for concurrent use.
func (m *MockTime) Set(t stdtime.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = t
}

// Advance moves the clock forward by d. Safe for concurrent use.
func (m *MockTime) Advance(d stdtime.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = m.now.Add(d)
}

// Compile-time assertions that both implementations satisfy the one-method
// interface consumers declare where they use it.
var (
	_ interface{ Now() stdtime.Time } = Time{}
	_ interface{ Now() stdtime.Time } = &MockTime{}
)
