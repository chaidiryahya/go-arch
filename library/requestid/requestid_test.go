package requestid_test

import (
	"context"
	"testing"

	"go-arch/library/requestid"
)

func TestWithThenFrom(t *testing.T) {
	ctx := requestid.With(context.Background(), "01a218474c8e7e7eab9d27b260635b64")
	if got := requestid.From(ctx); got != "01a218474c8e7e7eab9d27b260635b64" {
		t.Fatalf("From = %q", got)
	}
}

func TestFromWithoutIDReturnsEmpty(t *testing.T) {
	if got := requestid.From(context.Background()); got != "" {
		t.Fatalf("From on bare context = %q, want empty", got)
	}
}

// Nested contexts must not lose the outer id.
func TestWithSurvivesChildContext(t *testing.T) {
	parent := requestid.With(context.Background(), "abcdef0123456789")
	child, cancel := context.WithCancel(parent)
	defer cancel()

	if got := requestid.From(child); got != "abcdef0123456789" {
		t.Fatalf("From on child = %q", got)
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"32 hex", "01a218474c8e7e7eab9d27b260635b64", true},
		{"mixed alnum", "AbC123xy", true},
		{"64 chars", string(make([]byte, 0)) + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", true},
		{"empty", "", false},
		{"too short", "1234567", false},
		{"too long", "0123456789012345678901234567890123456789012345678901234567890123456789", false},
		{"has hyphen", "01a21847-4c8e-7e7e-ab9d-27b260635b64", false},
		{"newline injection", "0123456789abcdef\nFAKE LOG LINE", false},
		{"space", "0123456789abcde ", false},
		{"unicode", "012345678äöü", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := requestid.Valid(tc.id); got != tc.want {
				t.Fatalf("Valid(%q) = %v, want %v", tc.id, got, tc.want)
			}
		})
	}
}
