package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()
}
