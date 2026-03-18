package nsq

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
	h.logger.Info().Msg("nsq consumer started")

	// Register NSQ consumers here
	// Example:
	// consumer, _ := nsq.NewConsumer("topic", "channel", nsq.NewConfig())
	// consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error { ... }))
	// consumer.ConnectToNSQLookupd("localhost:4161")

	<-ctx.Done()
	h.logger.Info().Msg("nsq consumer stopped")
	return nil
}
