package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func renderGolangciTemplate(t *testing.T) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{}
	result, err := renderer.Render(".golangci.yml.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestGolangciTemplate(t *testing.T) {
	result := renderGolangciTemplate(t)
	assert.NotEmpty(t, result)
}

func TestGolangciValidYAML(t *testing.T) {
	result := renderGolangciTemplate(t)

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)
}

func TestGolangciConfiguration(t *testing.T) {
	result := renderGolangciTemplate(t)

	tests := []struct {
		name  string
		items []string
	}{
		{"required linters", []string{
			"errcheck",
			"govet",
			"staticcheck",
			"unused",
			"gosimple",
			"ineffassign",
			"contextcheck",
			"gofmt",
			"goimports",
		}},
		{"generated code exclusions", []string{
			"test/mocks",
			"db/generated",
			".*_templ",
		}},
		{"configuration settings", []string{
			"disable-all: true",
			"timeout: 5m",
			"max-issues-per-linter: 0",
			"max-same-issues: 0",
			"check-blank: true",
			"tests: true",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertContainsAll(t, result, tt.items)
		})
	}
}
