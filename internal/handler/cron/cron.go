package cron

import (
	"context"

	"github.com/rs/zerolog"
)

type Handler struct {
	logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) *Handler {
	return &Handler{logger: logger}
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
