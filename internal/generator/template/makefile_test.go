package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderMakefileTemplate(t *testing.T) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{}
	result, err := renderer.Render("Makefile.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestMakefileTemplate(t *testing.T) {
	result := renderMakefileTemplate(t)
	assert.NotEmpty(t, result)
}

func TestMakefileTargets(t *testing.T) {
	result := renderMakefileTemplate(t)

	tests := []struct {
		name  string
		items []string
	}{
		{"phony declarations", []string{
			".PHONY: assets build clean css dev dev-down dev-services generate help js lint mocks sqlc templ test",
		}},
		{"help target", []string{
			"help: ## Show this help message",
			"Available targets:",
		}},
		{"build target", []string{
			"build: generate assets ## Build server binary (includes code generation and assets)",
			"@mkdir -p bin",
			"go build -o bin/server ./cmd/server",
		}},
		{"assets target", []string{
			"assets: css js ## Build all assets (CSS and JS)",
		}},
		{"css target", []string{
			"css: ## Compile TailwindCSS with minification",
			"@mkdir -p internal/assets/dist/css",
			"npx @tailwindcss/cli -i internal/assets/web/css/app.css -o internal/assets/dist/css/app.css --minify",
		}},
		{"js target", []string{
			"js: ## Bundle JavaScript with esbuild and minification",
			"@mkdir -p internal/assets/dist/js",
			"npx esbuild internal/assets/web/js/*.js --bundle --minify --outdir=internal/assets/dist/js/",
		}},
		{"clean target", []string{
			"clean: ## Remove build artifacts",
			"rm -rf bin/",
		}},
		{"dev target", []string{
			"dev: ## Start development server (auto-starts services if needed)",
			"grep -q '^  [a-z]' docker-compose.yml",
			"docker-compose up -d",
			"go tool air -c .air.toml",
		}},
		{"lint target", []string{
			"lint: ## Run linters",
			"go tool golangci-lint run",
		}},
		{"mocks target", []string{
			"mocks: ## Generate mocks from interfaces",
			"go tool mockery",
		}},
		{"sqlc target", []string{
			"sqlc: ## Generate type-safe SQL code",
			"go tool sqlc generate",
		}},
		{"templ target", []string{
			"templ: ## Generate templ templates",
			"go tool templ generate",
		}},
		{"test target", []string{
			"test: ## Run tests",
			"go test ./...",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.AssertContainsAll(t, result, tt.items)
		})
	}
}

func TestMakefileHelpText(t *testing.T) {
	result := renderMakefileTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"assets       - Build all assets (CSS and JS)",
		"build        - Build the server binary (includes assets)",
		"clean        - Remove build artifacts",
		"css          - Compile TailwindCSS",
		"dev          - Start development server (auto-starts services if needed)",
		"dev-down     - Stop docker-compose services",
		"dev-services - Start docker-compose services",
		"generate     - Generate all code (templ, mocks, SQL)",
		"help         - Show this help message",
		"js           - Bundle JavaScript with esbuild",
		"lint         - Run linters",
		"mocks        - Generate mocks from interfaces",
		"sqlc         - Generate type-safe SQL code",
		"templ        - Generate templ templates",
		"test         - Run all tests",
	})
}

func TestMakefileUsesGoTool(t *testing.T) {
	result := renderMakefileTemplate(t)

	testutil.AssertContainsAll(t, result, []string{
		"go tool air",
		"go tool templ",
		"go tool mockery",
		"go tool sqlc",
		"go tool golangci-lint",
	})
}

func TestMakefileNoHardcodedPaths(t *testing.T) {
	result := renderMakefileTemplate(t)

	assert.Contains(t, result, "./cmd/server")
	assert.NotContains(t, result, "/myproject/")
	assert.NotContains(t, result, "/usr/local/")
}
