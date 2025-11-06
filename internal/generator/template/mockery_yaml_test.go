package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMockeryYamlTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestMockeryYamlValidYAML(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err, "generated YAML should be valid")
}

func TestMockeryYamlRequiredFields(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	requiredFields := []string{"dir", "filename", "pkgname", "packages"}
	for _, field := range requiredFields {
		assert.Contains(t, config, field, "should contain required field: %s", field)
	}
}

func TestMockeryYamlWithExpecterTrue(t *testing.T) {
	t.Skip("mockery v3 always generates expecter methods, no config needed")
}

func TestMockeryYamlDirectoryPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	assert.Equal(t, "tests/mocks", config["dir"], "dir should be tests/mocks")
}

func TestMockeryYamlOutputPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	filename, ok := config["filename"].(string)
	require.True(t, ok, "filename should be a string")
	assert.Contains(t, filename, "mock_", "filename should include mock_ prefix")
	assert.Contains(t, filename, ".InterfaceName", "filename should include .InterfaceName placeholder")
}

func TestMockeryYamlPackageName(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	assert.Equal(t, "mocks", config["pkgname"], "pkgname should be mocks")
}

func TestMockeryYamlAutoDiscovery(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	packages, ok := config["packages"].(map[string]interface{})
	require.True(t, ok, "packages should exist")

	packageKey := "github.com/test/app/internal/interfaces"
	packageCfg, ok := packages[packageKey].(map[string]interface{})
	require.True(t, ok, "should have package config for %s", packageKey)

	cfg, ok := packageCfg["config"].(map[string]interface{})
	require.True(t, ok, "should have config section")

	assert.Equal(t, true, cfg["all"], "all should be true for auto-discovery")
}

func TestMockeryYamlModuleNameInterpolation(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data1 := TemplateData{
		ModuleName: "github.com/test/app1",
	}

	data2 := TemplateData{
		ModuleName: "gitlab.com/org/app2",
	}

	result1, err := renderer.Render(".mockery.yaml.tmpl", data1)
	require.NoError(t, err)

	result2, err := renderer.Render(".mockery.yaml.tmpl", data2)
	require.NoError(t, err)

	assert.NotEqual(t, result1, result2, "template should interpolate module name")
	assert.Contains(t, result1, "github.com/test/app1/internal/interfaces", "should contain first module name")
	assert.Contains(t, result2, "gitlab.com/org/app2/internal/interfaces", "should contain second module name")
}
