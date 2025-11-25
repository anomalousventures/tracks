package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppCSSTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAppCSSTailwindV4Import(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "@import \"tailwindcss\"", "should use Tailwind v4 @import directive")
	assert.Contains(t, result, "Tailwind CSS v4", "should reference v4 in comments")
}

func TestAppCSSTailwindV4Theme(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "@theme {", "should use Tailwind v4 @theme directive")

	themeVars := []string{
		"--font-family-sans",
		"--font-family-mono",
		"--radius-sm",
		"--radius-md",
		"--radius-lg",
	}

	for _, themeVar := range themeVars {
		assert.Contains(t, result, themeVar, "should define theme variable: %s", themeVar)
	}
}

func TestAppCSSTemplUIVariables(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	templUIVars := []string{
		"--primary",
		"--primary-foreground",
		"--background",
		"--foreground",
		"--card",
		"--card-foreground",
		"--muted",
		"--muted-foreground",
		"--border",
		"--ring",
		"--radius",
	}

	for _, cssVar := range templUIVars {
		assert.Contains(t, result, cssVar, "should define templUI CSS variable: %s", cssVar)
	}
}

func TestAppCSSComponentLayer(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "@layer components {", "should have components layer")

	components := []string{
		".btn",
		".btn-primary",
		".btn-secondary",
		".card",
	}

	for _, component := range components {
		assert.Contains(t, result, component, "should define component: %s", component)
	}
}

func TestAppCSSUtilitiesLayer(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "@layer utilities {", "should have utilities layer")

	utilities := []string{
		".container-app",
		".page-header",
	}

	for _, utility := range utilities {
		assert.Contains(t, result, utility, "should define utility: %s", utility)
	}
}

func TestAppCSSDarkModeSupport(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "@media (prefers-color-scheme: dark)", "should have system dark mode support")
	assert.Contains(t, result, ".dark {", "should have manual dark mode class support")
}

func TestAppCSSNoConfigFile(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, "tailwind.config.js", "should not reference tailwind.config.js")
	assert.NotContains(t, result, "module.exports", "should not have JavaScript config")
}

func TestAppCSSUsesCustomProperties(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	customPropertyUsages := []string{
		"var(--primary)",
		"var(--background)",
		"var(--foreground)",
		"var(--border)",
		"var(--radius",
	}

	for _, usage := range customPropertyUsages {
		assert.Contains(t, result, usage, "should use CSS custom property: %s", usage)
	}
}
