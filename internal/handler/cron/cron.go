package cron

import (
	"context"

	infralog "go-arch/internal/infra/log"

	"github.com/rs/zerolog"
)

type Handler struct {
	logger zerolog.Logger
	ids    infralog.IDGenerator
}

// NewHandler takes the id generator so each job run can be correlated in logs.
// Every tick is a separate unit of work and gets its own request id, exactly as
// an inbound HTTP request does.
func NewHandler(logger zerolog.Logger, ids infralog.IDGenerator) *Handler {
	return &Handler{logger: logger, ids: ids}
}

func (h *Handler) Start(ctx context.Context) error {
	h.logger.Info().Msg("cron handler started")

	// Register cron jobs here
	// Example:
	// ticker := time.NewTicker(1 * time.Minute)
	// defer ticker.Stop()
	// for {
	//     select {
	//     case <-ctx.Done():
	//         return nil
	//     case <-ticker.C:
	//         h.runJob(ctx)
	//     }
	// }

	<-ctx.Done()
	h.logger.Info().Msg("cron handler stopped")
	return nil
}

// runJob is the pattern every scheduled job follows: derive a context carrying
// a fresh request id and a logger bound to it, then hand that context to the
// usecase. Everything downstream logs the same id without further plumbing.
//
// The name is kept deliberately concrete in this skeleton — when real jobs are
// added, copy this shape per job.
func (h *Handler) runJob(ctx context.Context, name string) {
	ctx, _ = infralog.NewRequestID(ctx, h.logger, h.ids)
	log := infralog.ContextLogger(ctx, h.logger)

	log.Info().Str("job", name).Msg("job started")
	defer log.Info().Str("job", name).Msg("job finished")

	// Call the usecase with ctx here:
	//   if err := h.exampleUsecase.DoSomething(ctx); err != nil { ... }
}
