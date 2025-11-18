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
	assert.Contains(t, result, `APIPrefix = "/api"`)
	assert.Contains(t, result, `Sitemap = "/sitemap.xml"`)
	assert.Contains(t, result, `"/robots.txt"`)
	assert.Contains(t, result, `"/llms.txt"`)
	assert.Contains(t, result, `"/.well-known/security.txt"`)
	assert.Contains(t, result, `"/.well-known/change-password"`)
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

func TestRoutesSharedConstantsOnly(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, "APIHealth", "APIHealth should be in health.go, not routes.go")
	assert.Contains(t, result, "APIPrefix", "should define shared APIPrefix constant")
	assert.Contains(t, result, "Sitemap", "should define shared Sitemap constant")
	assert.Contains(t, result, "Robots", "should define shared Robots constant")
}

func TestRoutesConstBlock(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "const (", "should use const block syntax")
	assert.NotContains(t, result, "const APIPrefix", "should not use individual const declarations")
}

func TestRoutesHealthTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/health.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Contains(t, result, "package routes")
	assert.Contains(t, result, "const (")
	assert.Contains(t, result, `APIHealth = APIPrefix + "/health"`)
}

func TestRoutesHealthValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/health.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "health.go", result, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go")
}

func TestRoutesHealthAPIHealthConstant(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	result, err := renderer.Render("internal/http/routes/health.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "APIHealth", "should define APIHealth constant")
	assert.Contains(t, result, `APIPrefix + "/health"`, "APIHealth should use APIPrefix constant")
}
