package context

import (
	"context"

	"github.com/rs/zerolog"
)

// contextKey is an unexported type for context keys to prevent collisions
// with keys from other packages.
type contextKey string

const (
	loggerKey contextKey = "logger"
)

// WithLogger attaches a logger to the context for propagation through the request lifecycle.
// This enables commands and services to access logging without direct dependencies.
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves the logger from the context.
// Returns a no-op logger if none is found, enabling graceful degradation
// rather than nil pointer panics.
func GetLogger(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	return zerolog.Nop()
}
