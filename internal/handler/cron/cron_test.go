package cron

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	infralog "go-arch/internal/infra/log"
	"go-arch/library/idgen"

	"github.com/rs/zerolog"
)

// Each job run must get its own request id, and every line it logs must carry
// it — the same guarantee HTTP requests get from middleware.
func TestRunJobCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(zerolog.New(&buf), idgen.NewFake())

	h.runJob(context.Background(), "nightly-settlement")

	logs := buf.String()
	if got := strings.Count(logs, "request_id"); got != 2 {
		t.Fatalf("expected request_id on both job lines, found %d:\n%s", got, logs)
	}

	// Both lines must share one id.
	ids := extractRequestIDs(logs)
	if len(ids) != 1 {
		t.Fatalf("job lines should share one id, got %v:\n%s", ids, logs)
	}
	if ids[0] == "" {
		t.Fatalf("empty request id:\n%s", logs)
	}
}

// Two separate runs must not reuse an id.
func TestRunJobUsesFreshIDEachRun(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(zerolog.New(&buf), idgen.NewFake())

	h.runJob(context.Background(), "job-a")
	h.runJob(context.Background(), "job-b")

	ids := extractRequestIDs(buf.String())
	if len(ids) != 2 {
		t.Fatalf("expected 2 distinct ids, got %v:\n%s", ids, buf.String())
	}
	if ids[0] == ids[1] {
		t.Fatalf("both runs shared id %q", ids[0])
	}
}

// extractRequestIDs pulls every distinct request_id value out of the JSON log
// lines the tests capture (the buffer is NOT wrapped in ConsoleWriter, so lines
// are raw JSON rather than the "request_id=x" console rendering).
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
		if _, dup := seen[entry.RequestID]; dup {
			continue
		}
		seen[entry.RequestID] = struct{}{}
		out = append(out, entry.RequestID)
	}
	return out
}

// The seam must be satisfied by the injected generator.
var _ infralog.IDGenerator = idgen.Real{}
