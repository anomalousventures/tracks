package template

import (
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoModTemplate tests rendering of go.mod.tmpl
func TestGoModTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		data         TemplateData
		wantContains []string
	}{
		{
			name: "basic go.mod",
			data: TemplateData{
				ModuleName: "github.com/user/myapp",
				GoVersion:  "1.25",
			},
			wantContains: []string{
				"module github.com/user/myapp",
				"go 1.25",
			},
		},
		{
			name: "different module path",
			data: TemplateData{
				ModuleName: "gitlab.com/org/project",
				GoVersion:  "1.23",
			},
			wantContains: []string{
				"module gitlab.com/org/project",
				"go 1.23",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render("go.mod.tmpl", tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestGitignoreTemplate tests rendering of .gitignore.tmpl
func TestGitignoreTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify it includes expected patterns
	expectedPatterns := []string{
		"bin/",
		"*.exe",
		"coverage.out",
		".env",
		"!.env.example",
		".DS_Store",
		".vscode/",
		".idea/",
	}

	for _, pattern := range expectedPatterns {
		assert.Contains(t, result, pattern, "should contain pattern: %s", pattern)
	}
}

// TestMainGoTemplate tests rendering of cmd/server/main.go.tmpl
func TestMainGoTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		data         TemplateData
		wantContains []string
	}{
		{
			name: "basic main.go",
			data: TemplateData{
				ProjectName: "myapp",
			},
			wantContains: []string{
				"package main",
				"import",
				"func main()",
				"myapp server starting...",
			},
		},
		{
			name: "different project name",
			data: TemplateData{
				ProjectName: "awesome-service",
			},
			wantContains: []string{
				"package main",
				"awesome-service server starting...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render("cmd/server/main.go.tmpl", tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestTracksYamlTemplate tests rendering of tracks.yaml.tmpl with different DB drivers
func TestTracksYamlTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name     string
		dbDriver string
	}{
		{"go-libsql driver", "go-libsql"},
		{"sqlite3 driver", "sqlite3"},
		{"postgres driver", "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				DBDriver: tt.dbDriver,
			}

			result, err := renderer.Render("tracks.yaml.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "database:")
			assert.Contains(t, result, "driver: "+tt.dbDriver)
			assert.Contains(t, result, "connection: ${DATABASE_URL}")
			assert.Contains(t, result, "server:")
			assert.Contains(t, result, "port: 8080")
			assert.Contains(t, result, "host: localhost")
		})
	}
}

// TestEnvExampleTemplate tests rendering of .env.example.tmpl
func TestEnvExampleTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".env.example.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify security warnings
	assert.Contains(t, result, "WARNING")
	assert.Contains(t, result, "NEVER commit .env")

	// Verify expected environment variables
	assert.Contains(t, result, "DATABASE_URL")
	assert.Contains(t, result, "PORT")

	// Verify placeholder values
	assert.Contains(t, result, "sqlite://data/app.db")
	assert.Contains(t, result, "8080")
}

// TestReadmeTemplate tests rendering of README.md.tmpl
func TestReadmeTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		data         TemplateData
		wantContains []string
	}{
		{
			name: "basic README",
			data: TemplateData{
				ProjectName: "myapp",
			},
			wantContains: []string{
				"# myapp",
				"Generated with [Tracks]",
				"## Setup",
				"## Development",
				"## Configuration",
				"make build",
				"make test",
				"make run",
			},
		},
		{
			name: "different project name",
			data: TemplateData{
				ProjectName: "cool-project",
			},
			wantContains: []string{
				"# cool-project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render("README.md.tmpl", tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestAllTemplatesRenderWithFullData tests all templates with complete TemplateData
func TestAllTemplatesRenderWithFullData(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/org/repo",
		ProjectName: "repo",
		DBDriver:    "postgres",
		GoVersion:   "1.25",
		Year:        2025,
	}

	templates := []struct {
		name     string
		template string
	}{
		{"go.mod", "go.mod.tmpl"},
		{".gitignore", ".gitignore.tmpl"},
		{"main.go", "cmd/server/main.go.tmpl"},
		{"tracks.yaml", "tracks.yaml.tmpl"},
		{".env.example", ".env.example.tmpl"},
		{"README.md", "README.md.tmpl"},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.name, func(t *testing.T) {
			result, err := renderer.Render(tmpl.template, data)
			require.NoError(t, err, "rendering %s should not fail", tmpl.name)
			assert.NotEmpty(t, result, "%s result should not be empty", tmpl.name)
		})
	}
}

// TestTemplatesWithEmptyData tests graceful handling of minimal TemplateData
func TestTemplatesWithEmptyData(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{}

	templates := []struct {
		name     string
		template string
	}{
		{"go.mod", "go.mod.tmpl"},
		{".gitignore", ".gitignore.tmpl"},
		{"main.go", "cmd/server/main.go.tmpl"},
		{"tracks.yaml", "tracks.yaml.tmpl"},
		{".env.example", ".env.example.tmpl"},
		{"README.md", "README.md.tmpl"},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.name, func(t *testing.T) {
			result, err := renderer.Render(tmpl.template, data)
			require.NoError(t, err, "rendering %s with empty data should not fail", tmpl.name)
			assert.NotEmpty(t, result, "%s result should not be empty", tmpl.name)
		})
	}
}

// TestGoModValidGoSyntax verifies go.mod output is valid
func TestGoModValidGoSyntax(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "example.com/test/module",
		GoVersion:  "1.25",
	}

	result, err := renderer.Render("go.mod.tmpl", data)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	require.GreaterOrEqual(t, len(lines), 2, "go.mod should have at least 2 lines")

	assert.True(t, strings.HasPrefix(lines[0], "module "), "first line should start with 'module'")
	assert.True(t, strings.HasPrefix(lines[2], "go "), "third line should start with 'go'")
}

// TestMainGoValidGoSyntax verifies main.go output is valid
func TestMainGoValidGoSyntax(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("cmd/server/main.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package main", "should have package main")
	assert.Contains(t, result, "func main()", "should have main function")
	assert.Contains(t, result, "import", "should have imports")
}
