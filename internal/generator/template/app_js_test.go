package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppJSTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAppJSModuleStructure(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "(function() {", "should use IIFE pattern")
	assert.Contains(t, result, "'use strict';", "should use strict mode")
	assert.Contains(t, result, "const App = {", "should define App object")
	assert.Contains(t, result, "window.App = App;", "should export to global scope")
}

func TestAppJSProjectName(t *testing.T) {
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

			result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, "name: '"+projectName+"'", "should use project name")
			assert.Contains(t, result, projectName+" - Main JavaScript Application", "should reference project in header")
		})
	}
}

func TestAppJSInitialization(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "init()", "should have init method")
	assert.Contains(t, result, "App.init();", "should auto-initialize")
	assert.Contains(t, result, "setupEventListeners()", "should setup event listeners")
	assert.Contains(t, result, "onReady(", "should have DOM ready handler")
}

func TestAppJSUtilityMethods(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	methods := []string{
		"isDevelopment()",
		"log(",
		"error(",
		"onReady(",
		"setupEventListeners()",
	}

	for _, method := range methods {
		assert.Contains(t, result, method, "should have utility method: %s", method)
	}
}

func TestAppJSDevelopmentMode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "isDevelopment()", "should detect development mode")
	assert.Contains(t, result, "window.location.hostname === 'localhost'", "should check for localhost")
	assert.Contains(t, result, "if (this.debug)", "should have debug flag")
}

func TestAppJSEventListeners(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "setupEventListeners()", "should setup event listeners")
	assert.Contains(t, result, "form[data-validate]", "should handle form validation")
	assert.Contains(t, result, "a[href^=\"http\"]", "should handle external links")
	assert.Contains(t, result, "addEventListener", "should use addEventListener")
}

func TestAppJSFormValidation(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "form.checkValidity()", "should check form validity")
	assert.Contains(t, result, "was-validated", "should add validation class")
	assert.Contains(t, result, "preventDefault()", "should prevent invalid submission")
}

func TestAppJSExternalLinks(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "target", "_blank", "should open external links in new tab")
	assert.Contains(t, result, "rel", "noopener noreferrer", "should add security attributes")
}

func TestAppJSDOMReady(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "document.readyState === 'loading'", "should check ready state")
	assert.Contains(t, result, "DOMContentLoaded", "should listen for DOMContentLoaded")
}

func TestAppJSNoHTMX(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
	}

	result, err := renderer.Render("internal/assets/web/js/app.js.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, "htmx", "should not include HTMX in Phase 2")
	assert.NotContains(t, result, "hx-", "should not reference hx- attributes in Phase 2")
}
