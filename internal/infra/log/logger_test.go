package log_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	infralog "go-arch/internal/infra/log"
	"go-arch/library/idgen"
	"go-arch/library/requestid"

	"github.com/rs/zerolog"
)

// The background-work entry point: a cron tick or NSQ message calls
// NewRequestID and gets a context whose logger already carries the id.
func TestNewRequestIDAttachesIDAndLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ids := idgen.NewMockGenerator()

	ctx, id := infralog.NewRequestID(context.Background(), logger, ids)
	if id == "" {
		t.Fatal("NewRequestID returned empty id")
	}
	if got := requestid.From(ctx); got != id {
		t.Fatalf("context carries %q, returned id is %q", got, id)
	}

	infralog.ContextLogger(ctx, logger).Info().Msg("background work")
	if !strings.Contains(buf.String(), `"request_id":"`+id+`"`) {
		t.Fatalf("log line missing request_id %q: %s", id, buf.String())
	}
}

// Two units of background work must not share an id.
func TestNewRequestIDIsUniquePerUnitOfWork(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ids := idgen.NewMockGenerator()

	_, first := infralog.NewRequestID(context.Background(), logger, ids)
	_, second := infralog.NewRequestID(context.Background(), logger, ids)

	if first == second {
		t.Fatalf("two units of work shared id %q", first)
	}
}

// A layer called without the middleware (a unit test, say) must still log —
// not silently drop the line the way zerolog.Ctx would.
func TestContextLoggerFallsBackInsteadOfDroppingLines(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	infralog.ContextLogger(context.Background(), logger).Error().Msg("still logged")

	if !strings.Contains(buf.String(), "still logged") {
		t.Fatalf("fallback logger dropped the line: %q", buf.String())
	}
}

// A context carrying an id but no attached logger should still surface the id.
func TestContextLoggerFallbackIncludesRequestIDFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	ctx := requestid.With(context.Background(), "abcdef0123456789")
	infralog.ContextLogger(ctx, logger).Error().Msg("boom")

	if !strings.Contains(buf.String(), `"request_id":"abcdef0123456789"`) {
		t.Fatalf("fallback lost the id from context: %s", buf.String())
	}
}
