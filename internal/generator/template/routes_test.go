package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutesTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Contains(t, result, "package routes")
	assert.Contains(t, result, "const (")
	assert.Contains(t, result, `APIHealth = "/api/health"`)
}

func TestRoutesValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "routes.go", result, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go")
}

func TestRoutesPackageDeclaration(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package routes", "package declaration should be 'routes'")
	assert.NotContains(t, result, "package http", "should not have 'http' package")
	assert.NotContains(t, result, "package main", "should not have 'main' package")
}

func TestRoutesAPIHealthConstant(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "APIHealth", "should define APIHealth constant")
	assert.Contains(t, result, `"/api/health"`, "APIHealth should have value /api/health")
}

func TestRoutesConstBlock(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "const (", "should use const block syntax")
	assert.NotContains(t, result, "const APIHealth", "should not use individual const declarations")
}
