package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthInterfacesTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name       string
		moduleName string
	}{
		{
			name:       "standard module path",
			moduleName: "github.com/user/myapp",
		},
		{
			name:       "gitlab module path",
			moduleName: "gitlab.com/org/project",
		},
		{
			name:       "simple module name",
			moduleName: "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/interfaces/health.go.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "package interfaces")
			assert.Contains(t, result, "type HealthService interface")
			assert.Contains(t, result, "type HealthStatus struct")
			assert.Contains(t, result, "Check(ctx context.Context) HealthStatus")
			assert.Contains(t, result, `Status    string    `+"`json:\"status\"`")
			assert.Contains(t, result, `Timestamp time.Time `+"`json:\"timestamp\"`")

			assert.NotContains(t, result, "internal/", "should not import any internal packages")
		})
	}
}

func TestHealthInterfacesValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/interfaces/health.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "health.go", result, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go")
}

func TestHealthInterfacesPackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/interfaces/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package interfaces", "package declaration should be 'interfaces'")
	assert.NotContains(t, result, "package health", "should not have 'health' package")
	assert.NotContains(t, result, "package main", "should not have 'main' package")
}

func TestHealthInterfacesOnlyStdlibImports(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/interfaces/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `"context"`, "should import context")
	assert.Contains(t, result, `"time"`, "should import time")

	assert.NotContains(t, result, "github.com/test/app/internal", "should not import project internal packages")
	assert.NotContains(t, result, "internal/domain", "should not import internal/domain")
	assert.NotContains(t, result, "internal/handlers", "should not import internal/handlers")
}

func TestHealthInterfacesHasGodocComments(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/interfaces/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "// HealthService", "HealthService should have godoc comment")
	assert.Contains(t, result, "// Check", "Check method should have godoc comment")
	assert.Contains(t, result, "// HealthStatus", "HealthStatus should have godoc comment")
}
