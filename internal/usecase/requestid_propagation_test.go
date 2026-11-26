package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-arch/internal/entity"
	apphttp "go-arch/internal/handler/http"
	"go-arch/internal/repository"
	"go-arch/internal/service"
	"go-arch/internal/usecase"
	"go-arch/library/idgen"
	"go-arch/library/requestid"
	clockpkg "go-arch/library/time"

	"github.com/rs/zerolog"
)

// errRepo is the boundary fake. It records the request id it saw, then fails,
// driving the error path through repository → service → usecase so every
// layer's log line is produced in one call.
type errRepo struct{ sawRequestID string }

func (r *errRepo) GetByID(ctx context.Context, id int64) (*entity.Example, error) {
	r.sawRequestID = requestid.From(ctx)
	return nil, errors.New("connection refused")
}

// The whole point of the change: one request produces log lines from the
// handler chain down to the repository, all carrying the same request_id.
func TestRequestIDFollowsEveryLayer(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	clk := clockpkg.NewMockTime(time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC))

	repo := &errRepo{}
	svc := service.NewExampleService(repo, logger)
	uc := usecase.NewExampleUsecase(svc, logger)

	// Exercise the usecase as a handler would, inside the real middleware chain.
	// The probe records the id the usecase received, since it must never be
	// written to the response.
	mux := http.NewServeMux()
	mux.HandleFunc("/examples", func(w http.ResponseWriter, r *http.Request) {
		if _, err := uc.GetByID(r.Context(), 42); err != nil {
			apphttp.WriteError(w, r, clk, http.StatusInternalServerError, "failed")
			return
		}
	})

	handler := apphttp.RequestIDMiddleware(idgen.NewMockGenerator())(
		apphttp.LoggingMiddleware(logger, clk)(
			apphttp.RecoveryMiddleware(logger, clk)(mux)))

	req := httptest.NewRequest(http.MethodGet, "/examples", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// The repository's view of the id is the authoritative one for assertions.
	reqID := repo.sawRequestID
	if reqID == "" {
		t.Fatal("no request id reached the repository")
	}

	logs := buf.String()
	for _, layer := range []string{"usecase", "service"} {
		// request_id now rides on the base logger, so it appears before the
		// per-line fields. Assert presence on the same line, not ordering.
		var found bool
		for _, line := range strings.Split(logs, "\n") {
			if strings.Contains(line, `"layer":"`+layer+`"`) &&
				strings.Contains(line, `"request_id":"`+reqID+`"`) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s log line missing request id %q:\n%s", layer, reqID, logs)
		}
	}

	// And it must not reach the client.
	if got := rec.Header().Get("X-Request-ID"); got != "" {
		t.Fatalf("X-Request-ID header must be absent, got %q", got)
	}
	if got := rec.Body.String(); strings.Contains(got, reqID) || strings.Contains(got, "request_id") {
		t.Fatalf("response leaked the request id: %s", got)
	}
}

// A fake repository implementation must satisfy the interface the service
// depends on; this mirrors how a real one would be swapped in for tests.
var _ repository.ExampleRepository = (*errRepo)(nil)
