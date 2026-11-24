// Package idgen provides the identifier seam: a real UUID v4 generator and an
// exported deterministic fake for tests.
package idgen

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Real is the production implementation. It is stateless and its zero value is
// ready to use.
type Real struct{}

// New returns a random (version 4) UUID string. It panics only if the system
// entropy source fails, exactly as uuid.New().String() does.
func (Real) New() string {
	return uuid.NewString()
}

// Fake is a deterministic, sequence-based generator for tests. Construct it
// with NewFake; the zero value is not usable. The mutex makes Fake safe for
// concurrent use by multiple goroutines.
type Fake struct {
	mu sync.Mutex
	n  uint64
}

// NewFake returns a Fake whose first generated value ends in 000000000001.
func NewFake() *Fake {
	return &Fake{}
}

// New returns the next identifier in the sequence, starting at 1. The value is
// shaped like a valid version 4 UUID (version nibble 4, variant nibble 8), so
// tests that assert UUID format or echo it in headers behave realistically.
// Safe for concurrent use.
func (f *Fake) New() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", f.n)
}

// Compile-time assertions that both implementations satisfy the one-method
// interface consumers declare where they use it.
var (
	_ interface{ New() string } = Real{}
	_ interface{ New() string } = &Fake{}
)
