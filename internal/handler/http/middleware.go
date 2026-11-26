package http

import (
	"context"
	"net/http"
	"time"

	infralog "go-arch/internal/infra/log"
	"go-arch/library/requestid"

	"github.com/rs/zerolog"
)

// The clock and idGenerator seams are declared here, where they are consumed,
// rather than in the provider packages (see "Define Interfaces Where They're
// Used"). Consumers only need the one method they call, so any type with a
// matching method can be injected — including library/clock and library/idgen.

// clock is the subset of time behavior this package needs.
type clock interface {
	Now() time.Time
}

// idGenerator is the subset of identifier generation this package needs.
type idGenerator interface {
	New() string
}

type contextKey string

const startTimeKey contextKey = "start_time"

// RequestIDMiddleware establishes the correlation id for the request.
//
// A client-supplied X-Request-ID is honoured only when it matches the accepted
// shape (alphanumeric, 8-64 chars); anything else — missing, too short, or
// containing control characters — is replaced with a freshly generated id, so
// untrusted input never reaches logs verbatim.
//
// The id is deliberately never written back to the client: no response header
// and no response body field. It is an internal correlation key for logs only,
// so an outside caller cannot learn, forge, or reason about how the service is
// instrumented.
func RequestIDMiddleware(ids idGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if !requestid.Valid(reqID) {
				reqID = ids.New()
			}

			ctx := requestid.With(r.Context(), reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware logs request start and completion and attaches a
// request-scoped logger to the context.
//
// The logging is done through the request logger — never the base logger — so
// request_id is present on every line. RecoveryMiddleware and the layers below
// reuse it via infralog.ContextLogger.
func LoggingMiddleware(logger zerolog.Logger, clk clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := clk.Now()

			ctx := infralog.WithRequestID(r.Context(), logger, requestid.From(r.Context()))
			ctx = context.WithValue(ctx, startTimeKey, start)
			reqLogger := infralog.ContextLogger(ctx, logger)

			reqLogger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("request started")

			next.ServeHTTP(w, r.WithContext(ctx))

			reqLogger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", clk.Now().Sub(start)).
				Msg("request completed")
		})
	}
}

func RecoveryMiddleware(logger zerolog.Logger, clk clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Already carrying request_id because LoggingMiddleware
					// attached it; the base logger is the fallback for a chain
					// that runs without it.
					infralog.ContextLogger(r.Context(), logger).Error().
						Interface("panic", rec).
						Str("path", r.URL.Path).
						Msg("panic recovered")
					WriteError(w, r, clk, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func GetStartTime(ctx context.Context, clk clock) time.Time {
	if t, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return t
	}
	return clk.Now()
}
