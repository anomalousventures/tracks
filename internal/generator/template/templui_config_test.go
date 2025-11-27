package template

import (
	"encoding/json"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplUIConfigTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/testapp",
	}

	result, err := renderer.Render(".templui.json.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	var config map[string]interface{}
	err = json.Unmarshal([]byte(result), &config)
	require.NoError(t, err, "should produce valid JSON")

	assert.Equal(t, "internal/http/views/components/ui", config["componentsDir"])
	assert.Equal(t, "internal/http/views/components/utils", config["utilsDir"])
	assert.Equal(t, "github.com/user/testapp", config["moduleName"])
	assert.Equal(t, "internal/assets/web/js", config["jsDir"])
	assert.Equal(t, "/assets/js", config["jsPublicPath"])
}

func TestTemplUIConfigDirectories(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/example/myapp",
	}

	result, err := renderer.Render(".templui.json.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	componentsDir, ok := config["componentsDir"].(string)
	require.True(t, ok, "should have componentsDir field")
	assert.Equal(t, "internal/http/views/components/ui", componentsDir, "componentsDir should match expected path")

	utilsDir, ok := config["utilsDir"].(string)
	require.True(t, ok, "should have utilsDir field")
	assert.Equal(t, "internal/http/views/components/utils", utilsDir, "utilsDir should match expected path")

	jsDir, ok := config["jsDir"].(string)
	require.True(t, ok, "should have jsDir field")
	assert.Equal(t, "internal/assets/web/js", jsDir, "jsDir should be source location for templUI JS output")

	jsPublicPath, ok := config["jsPublicPath"].(string)
	require.True(t, ok, "should have jsPublicPath field")
	assert.Equal(t, "/assets/js", jsPublicPath, "jsPublicPath should be URL path served by hashfs")
}

func TestTemplUIConfigModuleName(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []string{
		"github.com/user/myapp",
		"gitlab.com/company/service",
		"example.com/org/project",
	}

	for _, moduleName := range tests {
		t.Run(moduleName, func(t *testing.T) {
			data := TemplateData{
				ModuleName: moduleName,
			}

			result, err := renderer.Render(".templui.json.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = json.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			assert.Equal(t, moduleName, config["moduleName"])
		})
	}
}

func TestTemplUIConfigAssetPipeline(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/user/webapp",
	}

	result, err := renderer.Render(".templui.json.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	// Verify jsDir is the source location where templUI writes JS files
	jsDir := config["jsDir"].(string)
	assert.Equal(t, "internal/assets/web/js", jsDir,
		"jsDir should be source location (esbuild bundles from here to dist/js/)")

	// Verify jsPublicPath is the URL path where JS is served
	jsPublicPath := config["jsPublicPath"].(string)
	assert.Equal(t, "/assets/js", jsPublicPath,
		"jsPublicPath should be URL path (hashfs serves at /assets/* after embedding dist/)")
}
