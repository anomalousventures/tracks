package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderLoggerTemplate(t *testing.T) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{}
	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestLoggerTemplate(t *testing.T) {
	result := renderLoggerTemplate(t)
	assert.NotEmpty(t, result)
}

func TestLoggerValidGoCode(t *testing.T) {
	result := renderLoggerTemplate(t)
	testutil.AssertValidGoCode(t, result, "logger.go")
}

func TestLoggerImports(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		`"context"`,
		`"os"`,
		`"github.com/rs/zerolog"`,
	})
}

func TestLoggerStruct(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"type Logger struct",
		"logger zerolog.Logger",
	})
}

func TestLoggerNewLoggerFunction(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"func NewLogger(environment string) *Logger",
		"zerolog.TimeFieldFormat = zerolog.TimeFormatUnix",
		"return &Logger{logger: logger}",
	})
}

func TestLoggerEnvironmentLevels(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		`if environment == "development"`,
		"zerolog.ConsoleWriter{Out: os.Stderr}",
		"zerolog.SetGlobalLevel(zerolog.DebugLevel)",
		"zerolog.New(os.Stdout)",
		"zerolog.SetGlobalLevel(zerolog.InfoLevel)",
	})
}

func TestLoggerContextMethods(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"func (l *Logger) Debug(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Info(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Warn(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Error(ctx context.Context) *zerolog.Event",
		"return l.loggerWithContext(ctx).Debug()",
		"return l.loggerWithContext(ctx).Info()",
		"return l.loggerWithContext(ctx).Warn()",
		"return l.loggerWithContext(ctx).Error()",
	})
}

func TestLoggerContextKey(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"type contextKey string",
		`const requestIDKey contextKey = "request_id"`,
	})
}

func TestLoggerWithRequestID(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"func WithRequestID(ctx context.Context, requestID string) context.Context",
		"return context.WithValue(ctx, requestIDKey, requestID)",
	})
}

func TestLoggerWithContext(t *testing.T) {
	result := renderLoggerTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"func (l *Logger) loggerWithContext(ctx context.Context) *zerolog.Logger",
		"if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != \"\"",
		`logger = logger.With().Str("request_id", requestID).Logger()`,
	})
}

func TestLoggerTimestamp(t *testing.T) {
	result := renderLoggerTemplate(t)

	assert.Contains(t, result, ".With().Timestamp().Logger()")
}

func TestLoggerPackageDeclaration(t *testing.T) {
	result := renderLoggerTemplate(t)

	assert.Contains(t, result, "package logging")
}
