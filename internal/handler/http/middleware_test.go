package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-arch/library/idgen"
	"go-arch/library/requestid"

	"github.com/rs/zerolog"

	// The package declares its own `clock` interface, so the provider package
	// is imported under an alias to avoid shadowing it in this file.
	clockpkg "go-arch/library/clock"
)

// newTestRouter builds the same middleware chain NewRouter uses, without the
// health handler, so tests exercise only the correlation-id plumbing.
//
// The probe handler stashes the id it saw on the request context into a side
// channel rather than a response header, since the id must never reach the
// client. Tests read `seen` after ServeHTTP returns.
func newTestRouter(t *testing.T, capture *bytes.Buffer, seen *string) http.Handler {
	t.Helper()

	logger := zerolog.New(capture)
	clk := clockpkg.NewFake(clockTestTime)
	ids := idgen.NewFake()

	mux := http.NewServeMux()
	mux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		*seen = requestid.From(r.Context())
		WriteSuccess(w, r, clk, map[string]string{"status": "ok"})
	})

	return RequestIDMiddleware(ids)(
		LoggingMiddleware(logger, clk)(
			RecoveryMiddleware(logger, clk)(mux)))
}

func TestRequestIDGeneratedButNeverReturned(t *testing.T) {
	var (
		buf  bytes.Buffer
		seen string
	)
	h := newTestRouter(t, &buf, &seen)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen == "" {
		t.Fatal("no request id was generated for the handler")
	}
	if len(seen) != 32 || strings.Contains(seen, "-") {
		t.Fatalf("generated id %q is not 32 dashless chars", seen)
	}

	// The id must not leak through any response channel.
	if got := rec.Header().Get("X-Request-ID"); got != "" {
		t.Fatalf("X-Request-ID response header must be absent, got %q", got)
	}
	if body := rec.Body.String(); strings.Contains(body, seen) {
		t.Fatalf("response body must not contain the request id %q: %s", seen, body)
	}
	if body := rec.Body.String(); strings.Contains(body, "request_id") {
		t.Fatalf("response body must not have a request_id field: %s", body)
	}
}

func TestRequestIDAcceptedFromClientButNotEchoed(t *testing.T) {
	var (
		buf  bytes.Buffer
		seen string
	)
	h := newTestRouter(t, &buf, &seen)

	const clientID = "abcdef0123456789"
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Request-ID", clientID)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// A validated client id is used internally...
	if seen != clientID {
		t.Fatalf("handler saw %q, want client id %q", seen, clientID)
	}
	// ...but still never reflected back.
	if got := rec.Header().Get("X-Request-ID"); got != "" {
		t.Fatalf("client id must not be echoed back, got header %q", got)
	}
	if body := rec.Body.String(); strings.Contains(body, clientID) {
		t.Fatalf("response body must not echo client id: %s", body)
	}
}

func TestRequestIDRejectsUnsafeClientValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"too short", "abc"},
		{"newline injection", "0123456789abcdef\nINJECTED"},
		{"hyphenated uuid", "01a21847-4c8e-7e7e-ab9d-27b260635b64"},
		{"overlong", strings.Repeat("a", 200)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var (
				buf  bytes.Buffer
				seen string
			)
			h := newTestRouter(t, &buf, &seen)

			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			req.Header.Set("X-Request-ID", tc.value)

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if seen == tc.value {
				t.Fatalf("unsafe client id %q was accepted verbatim", tc.value)
			}
			if len(seen) != 32 {
				t.Fatalf("replacement id %q is not the generated 32-char shape", seen)
			}
		})
	}
}

func TestRequestIDReachesEveryLogLine(t *testing.T) {
	var (
		buf  bytes.Buffer
		seen string
	)
	h := newTestRouter(t, &buf, &seen)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	logs := buf.String()
	lines := strings.Split(strings.TrimSpace(logs), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d:\n%s", len(lines), logs)
	}

	// Every line — not just some — must carry the id.
	for _, line := range lines {
		if !strings.Contains(line, `"request_id":"`+seen+`"`) {
			t.Fatalf("log line missing request_id %q:\n%s", seen, line)
		}
	}
}

func TestRequestIDAvailableDuringPanicRecovery(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	clk := clockpkg.NewFake(clockTestTime)

	var seen string
	panicky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = requestid.From(r.Context())
		panic("boom")
	})

	handler := RequestIDMiddleware(idgen.NewFake())(
		LoggingMiddleware(logger, clk)(
			RecoveryMiddleware(logger, clk)(panicky)))

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if seen == "" {
		t.Fatal("no request id generated on panic path")
	}
	if !strings.Contains(buf.String(), `"request_id":"`+seen+`"`) {
		t.Fatalf("panic log missing request_id %q:\n%s", seen, buf.String())
	}
}
