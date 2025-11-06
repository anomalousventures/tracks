package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderLoggingMiddlewareTemplate(t *testing.T, moduleName string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{ModuleName: moduleName}
	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestLoggingMiddlewareTemplate(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
		{"simple name", "myapp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderLoggingMiddlewareTemplate(t, tt.moduleName)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLoggingMiddlewareValidGoCode(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderLoggingMiddlewareTemplate(t, tt.moduleName)
			helpers.AssertValidGoCode(t, result, "logging.go")
		})
	}
}

func TestLoggingMiddlewareImports(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		`"net/http"`,
		`"time"`,
		`"github.com/go-chi/chi/v5/middleware"`,
	})
}

func TestLoggingMiddlewareModuleNameInterpolation(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
		{"simple name", "myapp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderLoggingMiddlewareTemplate(t, tt.moduleName)

			assert.Contains(t, result, `"`+tt.moduleName+`/internal/interfaces"`)
			assert.Contains(t, result, `"`+tt.moduleName+`/internal/logging"`)
		})
	}
}

func TestLoggingMiddlewareResponseWriter(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"type responseWriter struct",
		"http.ResponseWriter",
		"status       int",
		"bytesWritten int",
	})
}

func TestLoggingMiddlewareResponseWriterMethods(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"func (rw *responseWriter) WriteHeader(status int)",
		"rw.status = status",
		"rw.ResponseWriter.WriteHeader(status)",
		"func (rw *responseWriter) Write(b []byte) (int, error)",
		"n, err := rw.ResponseWriter.Write(b)",
		"rw.bytesWritten += n",
	})
}

func TestLoggingMiddlewareNewRequestLogger(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"func NewRequestLogger(logger interfaces.Logger) func(next http.Handler) http.Handler",
		"return func(next http.Handler) http.Handler",
		"return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)",
	})
}

func TestLoggingMiddlewareRequestLogging(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"start := time.Now()",
		"requestID := middleware.GetReqID(r.Context())",
		"ctx := logging.WithRequestID(r.Context(), requestID)",
		"logger.Info(ctx)",
		`Str("method", r.Method)`,
		`Str("path", r.URL.Path)`,
		`Str("remote_addr", r.RemoteAddr)`,
		`Msg("request started")`,
		"duration := time.Since(start)",
		`Int("status", wrapped.status)`,
		`Int("bytes", wrapped.bytesWritten)`,
		`Dur("duration_ms", duration)`,
		`Msg("request completed")`,
	})
}

func TestLoggingMiddlewareWrapsResponseWriter(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}",
		"next.ServeHTTP(wrapped, r.WithContext(ctx))",
	})
}

func TestLoggingMiddlewareNewRecoverer(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"func NewRecoverer(logger interfaces.Logger) func(next http.Handler) http.Handler",
		"return func(next http.Handler) http.Handler",
		"return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)",
	})
}

func TestLoggingMiddlewarePanicRecovery(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"defer func()",
		"if rvr := recover(); rvr != nil",
		"logger.Error(r.Context())",
		`Interface("panic", rvr)`,
		`Str("method", r.Method)`,
		`Str("path", r.URL.Path)`,
		`Stack()`,
		`Msg("panic recovered")`,
		"w.WriteHeader(http.StatusInternalServerError)",
		"next.ServeHTTP(w, r)",
	})
}

func TestLoggingMiddlewarePackageDeclaration(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	assert.Contains(t, result, "package middleware")
}

func TestLoggingMiddlewareUsesInterfacesLogger(t *testing.T) {
	result := renderLoggingMiddlewareTemplate(t, "github.com/user/project")

	helpers.AssertContainsAll(t, result, []string{
		"logger interfaces.Logger",
		"logger.Info(ctx)",
		"logger.Error(r.Context())",
	})
}
