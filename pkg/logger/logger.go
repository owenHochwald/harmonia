package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(env string) zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if env == "dev" {
		logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return logger
}
