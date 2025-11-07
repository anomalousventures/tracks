package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/tests/helpers"
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
			".PHONY: help test build dev dev-services dev-full dev-down generate mocks sqlc lint clean",
		}},
		{"help target", []string{
			"help: ## Show this help message",
			"Available targets:",
		}},
		{"test target", []string{
			"test: ## Run tests",
			"go test ./...",
		}},
		{"build target", []string{
			"build: ## Build server binary",
			"@mkdir -p bin",
			"go build -o bin/server ./cmd/server",
		}},
		{"dev target", []string{
			"dev: ## Start development server with hot reload",
			"go tool air -c .air.toml",
		}},
		{"mocks target", []string{
			"mocks: ## Generate mocks from interfaces",
			"go tool mockery",
		}},
		{"sqlc target", []string{
			"sqlc: ## Generate type-safe SQL code",
			"go tool sqlc generate",
		}},
		{"lint target", []string{
			"lint: ## Run linters",
			"go tool golangci-lint run",
		}},
		{"clean target", []string{
			"clean: ## Remove build artifacts",
			"rm -rf bin/",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertContainsAll(t, result, tt.items)
		})
	}
}

func TestMakefileHelpText(t *testing.T) {
	result := renderMakefileTemplate(t)

	helpers.AssertContainsAll(t, result, []string{
		"help         - Show this help message",
		"test         - Run all tests",
		"build        - Build the server binary",
		"dev          - Start development server with hot reload",
		"dev-services - Start docker-compose services",
		"dev-full     - Start docker services and dev server",
		"dev-down     - Stop docker-compose services",
		"generate     - Generate mocks and SQL code",
		"mocks        - Generate mocks from interfaces",
		"sqlc         - Generate type-safe SQL code",
		"lint         - Run linters",
		"clean        - Remove build artifacts",
	})
}

func TestMakefileUsesGoTool(t *testing.T) {
	result := renderMakefileTemplate(t)

	helpers.AssertContainsAll(t, result, []string{
		"go tool air",
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
