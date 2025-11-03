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

	requiredFields := []string{"with-expecter", "dir", "output", "outpkg", "all"}
	for _, field := range requiredFields {
		assert.Contains(t, config, field, "should contain required field: %s", field)
	}
}

func TestMockeryYamlWithExpecterTrue(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".mockery.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	assert.Equal(t, true, config["with-expecter"], "with-expecter should be true")
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

	assert.Equal(t, "internal/interfaces", config["dir"], "dir should be internal/interfaces")
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

	output, ok := config["output"].(string)
	require.True(t, ok, "output should be a string")
	assert.Contains(t, output, "test/mocks/", "output should include test/mocks/ directory")
	assert.Contains(t, output, ".InterfaceName", "output should include .InterfaceName placeholder")
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

	assert.Equal(t, "mocks", config["outpkg"], "outpkg should be mocks")
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

	assert.Equal(t, true, config["all"], "all should be true for auto-discovery")
}

func TestMockeryYamlIsStatic(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data1 := TemplateData{
		ModuleName:  "github.com/test/app1",
		ProjectName: "project1",
		DBDriver:    "postgres",
	}

	data2 := TemplateData{
		ModuleName:  "gitlab.com/org/app2",
		ProjectName: "project2",
		DBDriver:    "sqlite3",
	}

	result1, err := renderer.Render(".mockery.yaml.tmpl", data1)
	require.NoError(t, err)

	result2, err := renderer.Render(".mockery.yaml.tmpl", data2)
	require.NoError(t, err)

	assert.Equal(t, result1, result2, "template should produce identical output regardless of template data")
}
