package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMXConfigTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestHTMXConfigPackage(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package components", "should be in components package")
}

func TestHTMXConfigUsesTemplGetNonce(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "templ HTMXConfig()", "should have no parameters")
	assert.NotContains(t, result, "import", "should not have explicit import - templ provides it automatically")
	assert.Contains(t, result, "templ.GetNonce(ctx)", "should use templ.GetNonce(ctx) directly")
}

func TestHTMXConfigMetaTag(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `<meta name="htmx-config"`, "should render htmx-config meta tag")
}

func TestHTMXConfigOptions(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `selfRequestsOnly`, "should include selfRequestsOnly option")
	assert.Contains(t, result, `inlineScriptNonce`, "should include inlineScriptNonce option")
	assert.Contains(t, result, `useTemplateFragments`, "should include useTemplateFragments option")
	assert.Contains(t, result, `scrollBehavior`, "should include scrollBehavior option")
}

func TestHTMXConfigNonceInterpolation(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/http/views/components/htmx_config.templ.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `templ.GetNonce(ctx)`, "should interpolate nonce via templ.GetNonce(ctx)")
}
