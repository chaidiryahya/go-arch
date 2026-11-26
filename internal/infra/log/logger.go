package log

import (
	"context"
	"os"
	"time"

	"go-arch/library/requestid"

	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()
}

// IDGenerator is the subset of identifier generation this package needs.
// Declared here, where it is consumed, so any type with a New method —
// including library/idgen.Real and its Fake — can be injected.
type IDGenerator interface {
	New() string
}

// loggerKey is the context key under which WithRequestID stores the request
// logger. We keep our own key rather than relying on zerolog's WithContext so
// that ContextLogger can tell "a logger is attached" apart from "zerolog would
// hand back a disabled logger".
type loggerKey struct{}

// WithRequestID stores id on ctx (via library/requestid) and attaches a logger
// already carrying request_id to the returned context.
//
// This is what makes the id follow the process: handlers pass the returned
// context down, and every layer logs through ContextLogger(ctx, ...) to get
// request_id for free, without threading it through function signatures.
func WithRequestID(ctx context.Context, logger zerolog.Logger, id string) context.Context {
	reqLogger := logger.With().Str("request_id", id).Logger()
	ctx = requestid.With(ctx, id)
	return context.WithValue(ctx, loggerKey{}, &reqLogger)
}

// NewRequestID generates a fresh id, then delegates to WithRequestID. It is the
// entry point for units of work that have no inbound request to borrow an id
// from — a cron job firing or an NSQ message arriving. Returns the derived
// context and the generated id.
func NewRequestID(ctx context.Context, logger zerolog.Logger, gen IDGenerator) (context.Context, string) {
	id := gen.New()
	return WithRequestID(ctx, logger, id), id
}

// ContextLogger returns the logger attached to ctx by WithRequestID. When none
// is attached — a layer exercised directly in a test, say — it falls back to
// base carrying whatever request id ctx has, so a log line is never silently
// dropped (zerolog.Ctx would return a disabled logger in that case).
func ContextLogger(ctx context.Context, base zerolog.Logger) *zerolog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*zerolog.Logger); ok {
		return l
	}
	l := base.With().Str("request_id", requestid.From(ctx)).Logger()
	return &l
}
