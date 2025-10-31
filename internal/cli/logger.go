package cli

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(logLevel string) zerolog.Logger {
	level := parseLogLevel(logLevel)

	zerolog.SetGlobalLevel(level)

	output := io.Discard
	if level != zerolog.Disabled {
		output = zerolog.ConsoleWriter{Out: os.Stderr}
	}

	return zerolog.New(output).With().Timestamp().Logger()
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "off":
		return zerolog.Disabled
	default:
		return zerolog.Disabled
	}
}
