package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebCSSTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ProjectName: "testapp",
		ModuleName:  "github.com/test/testapp",
	}

	t.Run("renders app.css template", func(t *testing.T) {
		result, err := renderer.Render("web/css/app.css.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "Tailwind CSS entry point")
		assert.Contains(t, result, "body {")
		assert.Contains(t, result, "font-family: system-ui")
	})

	t.Run("renders valid CSS", func(t *testing.T) {
		result, err := renderer.Render("web/css/app.css.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "/*")
		assert.Contains(t, result, "*/")
		assert.Contains(t, result, "margin: 0;")
		assert.Contains(t, result, "padding: 0;")
	})
}

func TestWebJSTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ProjectName: "testapp",
		ModuleName:  "github.com/test/testapp",
	}

	t.Run("renders app.js template", func(t *testing.T) {
		result, err := renderer.Render("web/js/app.js.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "JavaScript entry point")
		assert.Contains(t, result, "testapp loaded")
	})

	t.Run("uses project name in console.log", func(t *testing.T) {
		result, err := renderer.Render("web/js/app.js.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "console.log('testapp loaded');")
	})

	t.Run("renders valid JavaScript", func(t *testing.T) {
		result, err := renderer.Render("web/js/app.js.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "//")
		assert.Contains(t, result, "console.log")
	})

	t.Run("handles different project names", func(t *testing.T) {
		customData := TemplateData{
			ProjectName: "my-awesome-app",
			ModuleName:  "github.com/test/my-awesome-app",
		}
		result, err := renderer.Render("web/js/app.js.tmpl", customData)
		require.NoError(t, err)
		assert.Contains(t, result, "my-awesome-app loaded")
	})
}

func TestWebImagesGitkeep(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ProjectName: "testapp",
		ModuleName:  "github.com/test/testapp",
	}

	t.Run("renders .gitkeep template", func(t *testing.T) {
		result, err := renderer.Render("web/images/.gitkeep.tmpl", data)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run(".gitkeep contains comment", func(t *testing.T) {
		result, err := renderer.Render("web/images/.gitkeep.tmpl", data)
		require.NoError(t, err)
		assert.Contains(t, result, "#")
		assert.Contains(t, result, "preserves")
	})
}
