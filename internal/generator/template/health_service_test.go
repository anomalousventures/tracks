package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthServiceTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name       string
		moduleName string
		wantImport string
	}{
		{
			name:       "standard module path",
			moduleName: "github.com/user/myapp",
			wantImport: `"github.com/user/myapp/internal/interfaces"`,
		},
		{
			name:       "gitlab module path",
			moduleName: "gitlab.com/org/project",
			wantImport: `"gitlab.com/org/project/internal/interfaces"`,
		},
		{
			name:       "simple module name",
			moduleName: "myapp",
			wantImport: `"myapp/internal/interfaces"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/domain/health/service.go.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "package health")
			assert.Contains(t, result, tt.wantImport, "should import interfaces package from module")
			assert.Contains(t, result, "type service struct")
			assert.Contains(t, result, "func NewService() interfaces.HealthService")
			assert.Contains(t, result, "func (s *service) Check(ctx context.Context) interfaces.HealthStatus")
			assert.Contains(t, result, `Status:    "ok"`)
			assert.Contains(t, result, "time.Now()")
		})
	}
}

func TestHealthServiceValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/domain/health/service.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "service.go", result, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go")
}

func TestHealthServicePackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/domain/health/service.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package health", "package declaration should be 'health'")
	assert.NotContains(t, result, "package interfaces", "should not have 'interfaces' package")
	assert.NotContains(t, result, "package main", "should not have 'main' package")
}

func TestHealthServiceImportsInterfaces(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name       string
		moduleName string
		wantImport string
	}{
		{
			name:       "github module",
			moduleName: "github.com/user/myapp",
			wantImport: "github.com/user/myapp/internal/interfaces",
		},
		{
			name:       "gitlab module",
			moduleName: "gitlab.com/org/project",
			wantImport: "gitlab.com/org/project/internal/interfaces",
		},
		{
			name:       "simple name",
			moduleName: "myapp",
			wantImport: "myapp/internal/interfaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ModuleName: tt.moduleName,
			}

			result, err := renderer.Render("internal/domain/health/service.go.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, tt.wantImport, "should import interfaces package with correct module path")
		})
	}
}

func TestHealthServiceImplementsInterface(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/domain/health/service.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewService() interfaces.HealthService", "NewService should return HealthService interface")
	assert.Contains(t, result, "return &service{}", "NewService should return pointer to service")
	assert.Contains(t, result, "func (s *service) Check(ctx context.Context) interfaces.HealthStatus", "Check should match interface signature")
}
