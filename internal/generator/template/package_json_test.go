package template

import (
	"encoding/json"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageJSONTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("package.json.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	var pkg map[string]interface{}
	err = json.Unmarshal([]byte(result), &pkg)
	require.NoError(t, err, "should produce valid JSON")

	assert.Equal(t, "testapp", pkg["name"])
	assert.Equal(t, "1.0.0", pkg["version"])
	assert.Equal(t, true, pkg["private"])
	assert.Contains(t, pkg["description"], "Tracks")
}

func TestPackageJSONBuildScripts(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("package.json.tmpl", data)
	require.NoError(t, err)

	var pkg map[string]interface{}
	err = json.Unmarshal([]byte(result), &pkg)
	require.NoError(t, err)

	scripts, ok := pkg["scripts"].(map[string]interface{})
	require.True(t, ok, "should have scripts section")

	cssScript, ok := scripts["build:css"].(string)
	require.True(t, ok, "should have build:css script")
	assert.Contains(t, cssScript, "tailwindcss")
	assert.Contains(t, cssScript, "internal/assets/web/css/app.css")
	assert.Contains(t, cssScript, "internal/assets/dist/css/app.css")
	assert.Contains(t, cssScript, "--minify")

	jsScript, ok := scripts["build:js"].(string)
	require.True(t, ok, "should have build:js script")
	assert.Contains(t, jsScript, "esbuild")
	assert.Contains(t, jsScript, "internal/assets/web/js/*.js")
	assert.Contains(t, jsScript, "--bundle")
	assert.Contains(t, jsScript, "--minify")
	assert.Contains(t, jsScript, "--outdir=internal/assets/dist/js/")
}

func TestPackageJSONDependencies(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("package.json.tmpl", data)
	require.NoError(t, err)

	var pkg map[string]interface{}
	err = json.Unmarshal([]byte(result), &pkg)
	require.NoError(t, err)

	devDeps, ok := pkg["devDependencies"].(map[string]interface{})
	require.True(t, ok, "should have devDependencies section")

	tailwindVersion, ok := devDeps["tailwindcss"].(string)
	require.True(t, ok, "should have tailwindcss dependency")
	assert.Equal(t, "4.1.17", tailwindVersion, "should use exact Tailwind v4 version")

	esbuildVersion, ok := devDeps["esbuild"].(string)
	require.True(t, ok, "should have esbuild dependency")
	assert.Equal(t, "0.27.0", esbuildVersion, "should use exact esbuild version")

	deps, ok := pkg["dependencies"].(map[string]interface{})
	require.True(t, ok, "should have dependencies section")

	htmxVersion, ok := deps["htmx.org"].(string)
	require.True(t, ok, "should have htmx.org dependency")
	assert.Equal(t, "2.0.8", htmxVersion, "should use exact HTMX v2 version")

	headSupportVersion, ok := deps["htmx-ext-head-support"].(string)
	require.True(t, ok, "should have htmx-ext-head-support dependency")
	assert.Equal(t, "2.0.1", headSupportVersion, "should use exact head-support version")

	idiomorphVersion, ok := deps["idiomorph"].(string)
	require.True(t, ok, "should have idiomorph dependency")
	assert.Equal(t, "0.3.0", idiomorphVersion, "should use exact idiomorph version")

	responseTargetsVersion, ok := deps["htmx-ext-response-targets"].(string)
	require.True(t, ok, "should have htmx-ext-response-targets dependency")
	assert.Equal(t, "2.0.1", responseTargetsVersion, "should use exact response-targets version")
}

func TestPackageJSONExactVersions(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("package.json.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, "^", "should not use caret ranges")
	assert.NotContains(t, result, "~", "should not use tilde ranges")
}

func TestPackageJSONDifferentProjectNames(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []string{
		"myapp",
		"cool-project",
		"web-service",
	}

	for _, projectName := range tests {
		t.Run(projectName, func(t *testing.T) {
			data := TemplateData{
				ProjectName: projectName,
			}

			result, err := renderer.Render("package.json.tmpl", data)
			require.NoError(t, err)

			var pkg map[string]interface{}
			err = json.Unmarshal([]byte(result), &pkg)
			require.NoError(t, err)

			assert.Equal(t, projectName, pkg["name"])
		})
	}
}
