package cli

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type contextKey int

const loggerKey contextKey = iota

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

func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	return zerolog.Nop()
}
