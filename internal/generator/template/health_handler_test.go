package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandlerTemplate(t *testing.T) {
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

			result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "package handlers")
			assert.Contains(t, result, tt.wantImport, "should import interfaces package from module")
			assert.Contains(t, result, "type HealthHandler struct")
			assert.Contains(t, result, "func NewHealthHandler(svc interfaces.HealthService, logger interfaces.Logger) *HealthHandler")
			assert.Contains(t, result, "func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request)")
			assert.Contains(t, result, `w.Header().Set("Content-Type", "application/json")`)
			assert.Contains(t, result, "json.NewEncoder(w).Encode(status)")
		})
	}
}

func TestHealthHandlerValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "health.go", result, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go")
}

func TestHealthHandlerPackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package handlers", "package declaration should be 'handlers'")
	assert.NotContains(t, result, "package http", "should not have 'http' package")
	assert.NotContains(t, result, "package main", "should not have 'main' package")
}

func TestHealthHandlerImportsInterfaces(t *testing.T) {
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

			result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, tt.wantImport, "should import interfaces package with correct module path")
		})
	}
}

func TestHealthHandlerStructDefinition(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "type HealthHandler struct", "should define HealthHandler struct")
	assert.Contains(t, result, "healthService interfaces.HealthService", "should have healthService field with interface type")
}

func TestHealthHandlerConstructor(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewHealthHandler(svc interfaces.HealthService, logger interfaces.Logger) *HealthHandler", "NewHealthHandler should accept interface, logger and return pointer")
	assert.Contains(t, result, "return &HealthHandler{", "should return pointer to struct")
	assert.Contains(t, result, "healthService: svc", "should initialize healthService field")
	assert.Contains(t, result, "logger:        logger", "should initialize logger field")
}

func TestHealthHandlerCheckMethod(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request)", "Check should match http.HandlerFunc signature")
	assert.Contains(t, result, "ctx := r.Context()", "should extract context")
	assert.Contains(t, result, "h.healthService.Check(ctx)", "should call service with context")
}

func TestHealthHandlerSetsContentType(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `w.Header().Set("Content-Type", "application/json")`, "should set JSON content type header")
}

func TestHealthHandlerErrorHandling(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("internal/http/handlers/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "if err := json.NewEncoder(w).Encode(status); err != nil", "should check for encoding errors")
	assert.Contains(t, result, "h.logger.Error(ctx).Err(err).Msg", "should log error using handler logger")
	assert.Contains(t, result, `http.Error(w, "Failed to encode response", http.StatusInternalServerError)`, "should handle encoding errors properly")
	assert.Contains(t, result, "return", "should return after error")
}
