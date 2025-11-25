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
		"--color-primary-50",
		"--color-primary-100",
		"--color-primary-500",
		"--color-primary-900",
		"--font-family-sans",
		"--font-family-mono",
		"--spacing-page",
	}

	for _, themeVar := range themeVars {
		assert.Contains(t, result, themeVar, "should define theme variable: %s", themeVar)
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

	assert.Contains(t, result, "@media (prefers-color-scheme: dark)", "should have dark mode support")
	assert.Contains(t, result, "Dark mode support", "should document dark mode")
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

func TestAppCSSTailwindUtilities(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/css/app.css.tmpl", data)
	require.NoError(t, err)

	tailwindUtilities := []string{
		"@apply",
		"px-4",
		"py-2",
		"rounded-md",
		"font-medium",
		"transition-colors",
		"bg-primary-500",
		"text-white",
		"hover:bg-primary-600",
	}

	for _, utility := range tailwindUtilities {
		assert.Contains(t, result, utility, "should use Tailwind utility: %s", utility)
	}
}
