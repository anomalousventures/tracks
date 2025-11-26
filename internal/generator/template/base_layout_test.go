package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseLayoutTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestBaseLayoutPackage(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package layouts", "should be in layouts package")
}

func TestBaseLayoutAssetsImport(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `"github.com/example/testapp/internal/assets"`, "should import assets package")
}

func TestBaseLayoutCSSURL(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `href={ "/assets/" + assets.CSSURL() }`, "should use assets.CSSURL() for stylesheet")
}

func TestBaseLayoutJSURL(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `src={ "/assets/" + assets.JSURL() }`, "should use assets.JSURL() for script")
}

func TestBaseLayoutNoHardcodedAssetPaths(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/example/testapp",
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, `"/assets/css/app.css"`, "should not have hardcoded CSS path")
	assert.NotContains(t, result, `"/assets/js/app.js"`, "should not have hardcoded JS path")
}

func TestBaseLayoutModuleNameInterpolation(t *testing.T) {
	testCases := []struct {
		name       string
		moduleName string
	}{
		{
			name:       "simple module",
			moduleName: "myapp",
		},
		{
			name:       "github module",
			moduleName: "github.com/user/project",
		},
		{
			name:       "nested module",
			moduleName: "example.com/org/team/service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewRenderer(templates.FS)
			data := TemplateData{
				ModuleName:  tc.moduleName,
				ProjectName: "testproject",
			}

			output, err := renderer.Render("internal/http/views/layouts/base.templ.tmpl", data)
			require.NoError(t, err)

			expectedAssetsImport := tc.moduleName + "/internal/assets"
			assert.Contains(t, output, expectedAssetsImport, "should interpolate module name in assets import")
		})
	}
}
