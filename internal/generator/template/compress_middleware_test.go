package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderCompressMiddlewareTemplate(t *testing.T, moduleName string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{ModuleName: moduleName}
	result, err := renderer.Render("internal/http/middleware/compress.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestCompressMiddlewareTemplate(t *testing.T) {
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
			result := renderCompressMiddlewareTemplate(t, tt.moduleName)
			assert.NotEmpty(t, result)
		})
	}
}

func TestCompressMiddlewareValidGoCode(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderCompressMiddlewareTemplate(t, tt.moduleName)
			testutil.AssertValidGoCode(t, result, "compress.go")
		})
	}
}

func TestCompressMiddlewarePackageDeclaration(t *testing.T) {
	result := renderCompressMiddlewareTemplate(t, "github.com/user/project")
	assert.Contains(t, result, "package middleware")
}

func TestCompressMiddlewareImports(t *testing.T) {
	result := renderCompressMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		`"net/http"`,
		`"github.com/go-chi/chi/v5/middleware"`,
	})
}

func TestCompressMiddlewareNewCompressMiddleware(t *testing.T) {
	result := renderCompressMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		"func NewCompressMiddleware() func(next http.Handler) http.Handler",
		"return middleware.Compress(5,",
	})
}

func TestCompressMiddlewareContentTypes(t *testing.T) {
	result := renderCompressMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		`"text/html"`,
		`"text/css"`,
		`"text/plain"`,
		`"text/javascript"`,
		`"application/javascript"`,
		`"application/json"`,
		`"application/xml"`,
		`"image/svg+xml"`,
	})
}

func TestCompressMiddlewareCompressionLevel(t *testing.T) {
	result := renderCompressMiddlewareTemplate(t, "github.com/user/project")
	assert.Contains(t, result, "middleware.Compress(5,")
}
