package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddlewareTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

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
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLoggingMiddlewareValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
			require.NoError(t, err)

			fset := token.NewFileSet()
			_, err = parser.ParseFile(fset, "logging.go", result, parser.AllErrors)
			require.NoError(t, err, "generated logging.go should be valid Go code")
		})
	}
}

func TestLoggingMiddlewareImports(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	requiredImports := []string{
		`"net/http"`,
		`"time"`,
		`"github.com/go-chi/chi/v5/middleware"`,
	}

	for _, imp := range requiredImports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestLoggingMiddlewareModuleNameInterpolation(t *testing.T) {
	renderer := NewRenderer(templates.FS)

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
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, `"`+tt.moduleName+`/internal/interfaces"`)
			assert.Contains(t, result, `"`+tt.moduleName+`/internal/logging"`)
		})
	}
}

func TestLoggingMiddlewareResponseWriter(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "type responseWriter struct")
	assert.Contains(t, result, "http.ResponseWriter")
	assert.Contains(t, result, "status      int")
	assert.Contains(t, result, "bytesWritten int")
}

func TestLoggingMiddlewareResponseWriterMethods(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func (rw *responseWriter) WriteHeader(status int)")
	assert.Contains(t, result, "rw.status = status")
	assert.Contains(t, result, "rw.ResponseWriter.WriteHeader(status)")

	assert.Contains(t, result, "func (rw *responseWriter) Write(b []byte) (int, error)")
	assert.Contains(t, result, "n, err := rw.ResponseWriter.Write(b)")
	assert.Contains(t, result, "rw.bytesWritten += n")
}

func TestLoggingMiddlewareNewRequestLogger(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewRequestLogger(logger interfaces.Logger) func(next http.Handler) http.Handler")
	assert.Contains(t, result, "return func(next http.Handler) http.Handler")
	assert.Contains(t, result, "return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)")
}

func TestLoggingMiddlewareRequestLogging(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	requestStartChecks := []string{
		"start := time.Now()",
		"requestID := middleware.GetReqID(r.Context())",
		"ctx := logging.WithRequestID(r.Context(), requestID)",
		"logger.Info(ctx)",
		`Str("method", r.Method)`,
		`Str("path", r.URL.Path)`,
		`Str("remote_addr", r.RemoteAddr)`,
		`Msg("request started")`,
	}

	for _, check := range requestStartChecks {
		assert.Contains(t, result, check, "should log request start: %s", check)
	}

	requestCompleteChecks := []string{
		"duration := time.Since(start)",
		`Int("status", wrapped.status)`,
		`Int("bytes", wrapped.bytesWritten)`,
		`Dur("duration_ms", duration)`,
		`Msg("request completed")`,
	}

	for _, check := range requestCompleteChecks {
		assert.Contains(t, result, check, "should log request completion: %s", check)
	}
}

func TestLoggingMiddlewareResponseWriterWrapping(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}")
	assert.Contains(t, result, "next.ServeHTTP(wrapped, r.WithContext(ctx))")
}

func TestLoggingMiddlewareNewRecoverer(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewRecoverer(logger interfaces.Logger) func(next http.Handler) http.Handler")
	assert.Contains(t, result, "return func(next http.Handler) http.Handler")
	assert.Contains(t, result, "return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)")
}

func TestLoggingMiddlewarePanicRecovery(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	panicChecks := []string{
		"defer func()",
		"if rvr := recover(); rvr != nil",
		"logger.Error(r.Context())",
		`Interface("panic", rvr)`,
		`Str("method", r.Method)`,
		`Str("path", r.URL.Path)`,
		`Stack()`,
		`Msg("panic recovered")`,
		"w.WriteHeader(http.StatusInternalServerError)",
	}

	for _, check := range panicChecks {
		assert.Contains(t, result, check, "should handle panic recovery: %s", check)
	}

	assert.Contains(t, result, "next.ServeHTTP(w, r)")
}

func TestLoggingMiddlewarePackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package middleware")
}

func TestLoggingMiddlewareInterfacesUsage(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/project",
	}

	result, err := renderer.Render("internal/http/middleware/logging.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "logger interfaces.Logger", "should use interfaces.Logger type")
	assert.Contains(t, result, "logger.Info(ctx)", "should call logger methods")
	assert.Contains(t, result, "logger.Error(r.Context())", "should call logger methods")
}
