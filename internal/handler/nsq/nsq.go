package nsq

import (
	"context"

	infralog "go-arch/internal/infra/log"

	"github.com/rs/zerolog"
)

type Handler struct {
	logger zerolog.Logger
	ids    infralog.IDGenerator
}

// NewHandler takes the id generator so each consumed message can be correlated
// in logs. Every message is a separate unit of work and gets its own request
// id, exactly as an inbound HTTP request does.
func NewHandler(logger zerolog.Logger, ids infralog.IDGenerator) *Handler {
	return &Handler{logger: logger, ids: ids}
}

func (h *Handler) Start(ctx context.Context) error {
	h.logger.Info().Msg("nsq consumer started")

	// Register NSQ consumers here
	// Example:
	// consumer, _ := nsq.NewConsumer("topic", "channel", nsq.NewConfig())
	// consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
	//     return h.handleMessage(ctx, msg)
	// }))
	// consumer.ConnectToNSQLookupd("localhost:4161")

	<-ctx.Done()
	h.logger.Info().Msg("nsq consumer stopped")
	return nil
}

// handleMessage is the pattern every consumer follows: derive a context
// carrying a fresh request id and a logger bound to it, then hand that context
// to the usecase. Everything downstream logs the same id with no further
// plumbing.
//
// Keying the id per message (not per consumer connection) is what lets logs
// from concurrent messages be told apart.
func (h *Handler) handleMessage(ctx context.Context, topic string) {
	ctx, _ = infralog.NewRequestID(ctx, h.logger, h.ids)
	log := infralog.ContextLogger(ctx, h.logger)

	log.Info().Str("topic", topic).Msg("message received")
	defer log.Info().Str("topic", topic).Msg("message handled")

	// Call the usecase with ctx here:
	//   if err := h.exampleUsecase.DoSomething(ctx); err != nil { ... }
}
