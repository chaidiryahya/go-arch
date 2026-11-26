// Package requestid carries a request's correlation identifier through
// context.Context.
//
// It lives in library/ rather than in handler/http so that every layer can read
// the same value: handler, usecase, service, and repository all import From and
// see the id the HTTP middleware attached, with no upward dependency and no
// signature changes.
package requestid

import (
	"context"
	"regexp"
)

// contextKey is unexported so no other package can collide with or overwrite
// the entry. Zero-size struct avoids an allocation for the key itself.
type contextKey struct{}

// With returns a copy of ctx carrying id. An empty id is stored as-is and is
// indistinguishable from "no id"; callers should only pass a real one.
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// From returns the id stored in ctx, or "" when none is present, so callers can
// use it unconditionally in log fields.
func From(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

// validID is the accepted shape for a caller-supplied correlation id. It must
// be alphanumeric — matching what this project generates — and bounded so a
// client cannot flood logs with an enormous value. Hex ids are 32 chars; 8 is a
// floor that rejects trivially colliding values like "1".
var validID = regexp.MustCompile(`^[A-Za-z0-9]{8,64}$`)

// Valid reports whether id is an acceptable caller-supplied correlation id.
func Valid(id string) bool {
	return validID.MatchString(id)
}
