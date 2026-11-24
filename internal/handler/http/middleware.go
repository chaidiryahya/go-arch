package http

import (
	"context"
	"net/http"
	"time"

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

const (
	requestIDKey contextKey = "request_id"
	startTimeKey contextKey = "start_time"
)

func RequestIDMiddleware(ids idGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = ids.New()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, reqID)
			w.Header().Set("X-Request-ID", reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggingMiddleware(logger zerolog.Logger, clk clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := clk.Now()
			ctx := context.WithValue(r.Context(), startTimeKey, start)

			reqID, _ := r.Context().Value(requestIDKey).(string)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("request_id", reqID).
				Msg("request started")

			next.ServeHTTP(w, r.WithContext(ctx))

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("request_id", reqID).
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
					logger.Error().
						Interface("panic", rec).
						Str("path", r.URL.Path).
						Msg("panic recovered")
					WriteError(w, clk, clk.Now(), http.StatusInternalServerError, "internal server error")
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
