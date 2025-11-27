package template

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderAirConfigTemplate(t *testing.T, moduleName string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{ModuleName: moduleName}
	result, err := renderer.Render(".air.toml.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestAirConfigTemplate(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
		{"simple name", "myapp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderAirConfigTemplate(t, tt.moduleName)
			assert.NotEmpty(t, result)
		})
	}
}

func TestAirConfigValidTOML(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
	}{
		{"github module", "github.com/user/project"},
		{"gitlab module", "gitlab.com/org/service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderAirConfigTemplate(t, tt.moduleName)
			var cfg map[string]any
			err := toml.Unmarshal([]byte(result), &cfg)
			require.NoError(t, err)
		})
	}
}

func TestAirConfigWatchesAssetFiles(t *testing.T) {
	result := renderAirConfigTemplate(t, "github.com/user/project")

	assert.Contains(t, result, `"css"`)
	assert.Contains(t, result, `"js"`)
	assert.Contains(t, result, `"templ"`)
}

func TestAirConfigExcludesGeneratedFiles(t *testing.T) {
	result := renderAirConfigTemplate(t, "github.com/user/project")

	assert.Contains(t, result, "_templ.go")
	assert.Contains(t, result, "internal/assets/dist")
	assert.Contains(t, result, "internal/db/generated")
	assert.Contains(t, result, "tests/mocks")
}

func TestAirConfigHasPreCmd(t *testing.T) {
	result := renderAirConfigTemplate(t, "github.com/user/project")

	assert.Contains(t, result, "pre_cmd")
	assert.Contains(t, result, "make generate assets")
}

func TestAirConfigBuildSection(t *testing.T) {
	result := renderAirConfigTemplate(t, "github.com/user/project")
	var cfg map[string]any
	err := toml.Unmarshal([]byte(result), &cfg)
	require.NoError(t, err)

	build, ok := cfg["build"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "./tmp/main", build["bin"])
	assert.Equal(t, "go build -o ./tmp/main ./cmd/server", build["cmd"])
}
