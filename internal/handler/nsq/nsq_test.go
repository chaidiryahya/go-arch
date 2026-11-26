package nsq

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go-arch/library/idgen"

	"github.com/rs/zerolog"
)

// extractRequestIDs returns every distinct request_id across JSON log lines.
func extractRequestIDs(logs string) []string {
	var out []string
	seen := make(map[string]struct{})
	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		if line == "" {
			continue
		}
		var entry struct {
			RequestID string `json:"request_id"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.RequestID == "" {
			continue
		}
		if _, dup := seen[entry.RequestID]; dup {
			continue
		}
		seen[entry.RequestID] = struct{}{}
		out = append(out, entry.RequestID)
	}
	return out
}

// Each message must get its own request id, on every line.
func TestHandleMessageCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(zerolog.New(&buf), idgen.NewMockGenerator())

	h.handleMessage(context.Background(), "orders")

	ids := extractRequestIDs(buf.String())
	if len(ids) != 1 {
		t.Fatalf("message lines should share one id, got %v:\n%s", ids, buf.String())
	}
	if n := strings.Count(buf.String(), "request_id"); n != 2 {
		t.Fatalf("expected request_id on both lines, found %d:\n%s", n, buf.String())
	}
}

// Concurrent messages must not share an id — the whole point of keying per
// message rather than per consumer connection.
func TestHandleMessageUsesFreshIDPerMessage(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(zerolog.New(&buf), idgen.NewMockGenerator())

	h.handleMessage(context.Background(), "orders")
	h.handleMessage(context.Background(), "orders")

	ids := extractRequestIDs(buf.String())
	if len(ids) != 2 {
		t.Fatalf("expected 2 distinct ids, got %v:\n%s", ids, buf.String())
	}
	if ids[0] == ids[1] {
		t.Fatalf("two messages shared id %q", ids[0])
	}
}
