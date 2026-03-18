package http

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	startTimeKey contextKey = "start_time"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
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
				Dur("duration", time.Since(start)).
				Msg("request completed")
		})
	}
}

func RecoveryMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error().
						Interface("panic", rec).
						Str("path", r.URL.Path).
						Msg("panic recovered")
					WriteError(w, time.Now(), http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func GetStartTime(ctx context.Context) time.Time {
	if t, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return t
	}
	return time.Now()
}
