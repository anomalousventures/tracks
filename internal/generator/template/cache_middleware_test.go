package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderCacheMiddlewareTemplate(t *testing.T, moduleName string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{ModuleName: moduleName}
	result, err := renderer.Render("internal/http/middleware/cache.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestCacheMiddlewareTemplate(t *testing.T) {
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
			result := renderCacheMiddlewareTemplate(t, tt.moduleName)
			assert.NotEmpty(t, result)
		})
	}
}

func TestCacheMiddlewareValidGoCode(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderCacheMiddlewareTemplate(t, tt.moduleName)
			testutil.AssertValidGoCode(t, result, "cache.go")
		})
	}
}

func TestCacheMiddlewarePackageDeclaration(t *testing.T) {
	result := renderCacheMiddlewareTemplate(t, "github.com/user/project")
	assert.Contains(t, result, "package middleware")
}

func TestCacheMiddlewareImports(t *testing.T) {
	result := renderCacheMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		`"net/http"`,
		`"regexp"`,
	})
}

func TestCacheMiddlewareNewCacheMiddleware(t *testing.T) {
	result := renderCacheMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		"func NewCacheMiddleware() func(next http.Handler) http.Handler",
		"return func(next http.Handler) http.Handler",
	})
}

func TestCacheMiddlewareHashPattern(t *testing.T) {
	result := renderCacheMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		"hashPattern",
		"regexp.MustCompile",
		`[a-f0-9]{8,}`,
	})
}

func TestCacheMiddlewareImmutableCacheControl(t *testing.T) {
	result := renderCacheMiddlewareTemplate(t, "github.com/user/project")

	testutil.AssertContainsAll(t, result, []string{
		`"Cache-Control"`,
		`"public, max-age=31536000, immutable"`,
	})
}
