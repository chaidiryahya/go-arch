// Package idgen provides the identifier seam: a real UUID v7 generator and an
// exported deterministic fake for tests.
package idgen

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Real is the production implementation. It is stateless and its zero value is
// ready to use.
type Real struct{}

// New returns a version 7 (time-ordered) UUID as 32 lowercase hexadecimal
// characters, without hyphens — e.g. "01a218474c8e7e7eab9d27b260635b64".
//
// The 48-bit timestamp prefix means the strings sort lexicographically in
// creation order, which keeps them cheap to index and pleasant to read in logs.
// Hex is a subset of [0-9a-z], so the result is alphanumeric and URL-safe.
//
// New panics only if the system entropy source fails, exactly as the previous
// uuid.New().String() did.
func (Real) New() string {
	u, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("idgen: generating uuid v7: %v", err))
	}
	return hex.EncodeToString(u[:])
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

// New returns the next identifier in the sequence, starting at 1, rendered as
// 32 lowercase hexadecimal characters to match Real's shape exactly, so tests
// that assert format or length behave realistically. Safe for concurrent use.
func (f *Fake) New() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return fmt.Sprintf("%032x", f.n)
}

// Compile-time assertions that both implementations satisfy the one-method
// interface consumers declare where they use it.
var (
	_ interface{ New() string } = Real{}
	_ interface{ New() string } = &Fake{}
)
