// Package clock provides the time seam: a real implementation backed by the
// standard library and an exported fake for deterministic tests.
package clock

import (
	"sync"
	"time"
)

// Real is the production implementation. It is stateless and its zero value is
// ready to use.
type Real struct{}

// Now returns the current time exactly as time.Now reports it. It deliberately
// does not call Round, UTC, or anything else that would strip the monotonic
// reading, so Now().Sub(start) stays equivalent to time.Since(start) and keeps
// its protection against wall-clock jumps.
func (Real) Now() time.Time {
	return time.Now()
}

// Fake is a manually controlled clock. Construct it with NewFake; the zero
// value is not usable. The mutex makes Fake safe for concurrent use by
// multiple goroutines.
type Fake struct {
	mu  sync.Mutex
	now time.Time
}

// NewFake returns a Fake pinned to t.
func NewFake(t time.Time) *Fake {
	return &Fake{now: t}
}

// Now returns the clock's current time. Safe for concurrent use.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Set moves the clock to t. Safe for concurrent use.
func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t
}

// Advance moves the clock forward by d. Safe for concurrent use.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

// Compile-time assertions that both implementations satisfy the one-method
// interface consumers declare where they use it.
var (
	_ interface{ Now() time.Time } = Real{}
	_ interface{ Now() time.Time } = &Fake{}
)
