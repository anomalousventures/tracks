package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestLoggerValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "logger.go", result, parser.AllErrors)
	require.NoError(t, err, "generated logger.go should be valid Go code")
}

func TestLoggerImports(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	requiredImports := []string{
		`"context"`,
		`"os"`,
		`"github.com/rs/zerolog"`,
	}

	for _, imp := range requiredImports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestLoggerStruct(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "type Logger struct")
	assert.Contains(t, result, "logger zerolog.Logger")
}

func TestLoggerNewLoggerFunction(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewLogger(environment string) *Logger")
	assert.Contains(t, result, "zerolog.TimeFieldFormat = zerolog.TimeFormatUnix")
	assert.Contains(t, result, "return &Logger{logger: logger}")
}

func TestLoggerEnvironmentLevels(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	developmentChecks := []string{
		`if environment == "development"`,
		"zerolog.ConsoleWriter{Out: os.Stderr}",
		"zerolog.SetGlobalLevel(zerolog.DebugLevel)",
	}

	for _, check := range developmentChecks {
		assert.Contains(t, result, check, "should configure development logger: %s", check)
	}

	productionChecks := []string{
		"zerolog.New(os.Stdout)",
		"zerolog.SetGlobalLevel(zerolog.InfoLevel)",
	}

	for _, check := range productionChecks {
		assert.Contains(t, result, check, "should configure production logger: %s", check)
	}
}

func TestLoggerContextMethods(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	contextMethods := []string{
		"func (l *Logger) Debug(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Info(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Warn(ctx context.Context) *zerolog.Event",
		"func (l *Logger) Error(ctx context.Context) *zerolog.Event",
	}

	for _, method := range contextMethods {
		assert.Contains(t, result, method, "should have method: %s", method)
	}

	assert.Contains(t, result, "return l.loggerWithContext(ctx).Debug()")
	assert.Contains(t, result, "return l.loggerWithContext(ctx).Info()")
	assert.Contains(t, result, "return l.loggerWithContext(ctx).Warn()")
	assert.Contains(t, result, "return l.loggerWithContext(ctx).Error()")
}

func TestLoggerContextKey(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "type contextKey string")
	assert.Contains(t, result, `const requestIDKey contextKey = "request_id"`)
}

func TestLoggerWithRequestID(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func WithRequestID(ctx context.Context, requestID string) context.Context")
	assert.Contains(t, result, "return context.WithValue(ctx, requestIDKey, requestID)")
}

func TestLoggerWithContext(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func (l *Logger) loggerWithContext(ctx context.Context) *zerolog.Logger")
	assert.Contains(t, result, "if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != \"\"")
	assert.Contains(t, result, `logger = logger.With().Str("request_id", requestID).Logger()`)
}

func TestLoggerTimestamp(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, ".With().Timestamp().Logger()")
}

func TestLoggerPackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/logging/logger.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package logging")
}
