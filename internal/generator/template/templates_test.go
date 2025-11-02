package template

import (
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoModTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	coreDeps := []string{
		"github.com/go-chi/chi/v5",
		"github.com/a-h/templ",
		"github.com/pressly/goose/v3",
		"github.com/sqlc-dev/sqlc",
		"github.com/alexedwards/scs/v2",
		"github.com/rs/zerolog",
	}

	tests := []struct {
		name         string
		data         TemplateData
		wantContains []string
		wantNotContains []string
	}{
		{
			name: "go-libsql driver",
			data: TemplateData{
				ModuleName: "github.com/user/myapp",
				GoVersion:  "1.25",
				DBDriver:   "go-libsql",
			},
			wantContains: append(coreDeps,
				"module github.com/user/myapp",
				"go 1.25",
				"github.com/tursodatabase/libsql-client-go",
			),
			wantNotContains: []string{
				"github.com/mattn/go-sqlite3",
				"github.com/lib/pq",
			},
		},
		{
			name: "sqlite3 driver",
			data: TemplateData{
				ModuleName: "github.com/user/myapp",
				GoVersion:  "1.25",
				DBDriver:   "sqlite3",
			},
			wantContains: append(coreDeps,
				"module github.com/user/myapp",
				"go 1.25",
				"github.com/mattn/go-sqlite3",
			),
			wantNotContains: []string{
				"github.com/tursodatabase/libsql-client-go",
				"github.com/lib/pq",
			},
		},
		{
			name: "postgres driver",
			data: TemplateData{
				ModuleName: "github.com/user/myapp",
				GoVersion:  "1.25",
				DBDriver:   "postgres",
			},
			wantContains: append(coreDeps,
				"module github.com/user/myapp",
				"go 1.25",
				"github.com/lib/pq",
			),
			wantNotContains: []string{
				"github.com/tursodatabase/libsql-client-go",
				"github.com/mattn/go-sqlite3",
			},
		},
		{
			name: "different module path with postgres",
			data: TemplateData{
				ModuleName: "gitlab.com/org/project",
				GoVersion:  "1.23",
				DBDriver:   "postgres",
			},
			wantContains: append(coreDeps,
				"module gitlab.com/org/project",
				"go 1.23",
				"github.com/lib/pq",
			),
			wantNotContains: []string{
				"github.com/tursodatabase/libsql-client-go",
				"github.com/mattn/go-sqlite3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render("go.mod.tmpl", tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want, "expected to find: %s", want)
			}

			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, result, notWant, "should not contain: %s", notWant)
			}
		})
	}
}

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

func TestTracksYamlTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name                   string
		projectName            string
		dbDriver               string
		expectedConnection     string
		notExpectedConnections []string
		expectedMaxConn        string
		expectedMaxIdle        string
	}{
		{
			name:               "go-libsql driver",
			projectName:        "myapp",
			dbDriver:           "go-libsql",
			expectedConnection: "${DATABASE_URL:-file:./myapp.db}",
			notExpectedConnections: []string{
				"${DATABASE_URL:-./",
				"${DATABASE_URL:-postgres://",
			},
			expectedMaxConn: "max_connections: 10",
			expectedMaxIdle: "max_idle_connections: 2",
		},
		{
			name:               "sqlite3 driver",
			projectName:        "testapp",
			dbDriver:           "sqlite3",
			expectedConnection: "${DATABASE_URL:-./testapp.db}",
			notExpectedConnections: []string{
				"${DATABASE_URL:-file:./",
				"${DATABASE_URL:-postgres://",
			},
			expectedMaxConn: "max_connections: 10",
			expectedMaxIdle: "max_idle_connections: 2",
		},
		{
			name:               "postgres driver",
			projectName:        "webapp",
			dbDriver:           "postgres",
			expectedConnection: "${DATABASE_URL:-postgres://localhost/webapp?sslmode=disable}",
			notExpectedConnections: []string{
				"${DATABASE_URL:-file:./",
				"${DATABASE_URL:-./webapp.db}",
			},
			expectedMaxConn: "max_connections: 25",
			expectedMaxIdle: "max_idle_connections: 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ProjectName: tt.projectName,
				DBDriver:    tt.dbDriver,
			}

			result, err := renderer.Render("tracks.yaml.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "app:")
			assert.Contains(t, result, "environment: development")
			assert.Contains(t, result, "log_level: debug")

			assert.Contains(t, result, "server:")
			assert.Contains(t, result, "port: 8080")
			assert.Contains(t, result, "host: localhost")
			assert.Contains(t, result, "read_timeout: 10s")
			assert.Contains(t, result, "write_timeout: 10s")
			assert.Contains(t, result, "idle_timeout: 120s")

			assert.Contains(t, result, "database:")
			assert.Contains(t, result, "driver: "+tt.dbDriver)
			assert.Contains(t, result, "connection: "+tt.expectedConnection)
			assert.Contains(t, result, tt.expectedMaxConn)
			assert.Contains(t, result, tt.expectedMaxIdle)

			for _, notExpected := range tt.notExpectedConnections {
				assert.NotContains(t, result, notExpected, "should not contain %s (wrong driver)", notExpected)
			}

			if tt.dbDriver == "go-libsql" || tt.dbDriver == "sqlite3" {
				assert.Contains(t, result, "File-based databases work better with lower concurrency", "should contain WHY comment for file-based databases")
			} else if tt.dbDriver == "postgres" {
				assert.Contains(t, result, "sslmode=disable is safe for local development only", "should contain WHY comment for SSL mode")
			}

			assert.Contains(t, result, "session:")
			assert.Contains(t, result, "lifetime: 24h")
			assert.Contains(t, result, "cookie_name: session_id")
			assert.Contains(t, result, "cookie_secure: false")
			assert.Contains(t, result, "cookie_http_only: true")
			assert.Contains(t, result, "cookie_same_site: lax")

			assert.Contains(t, result, "WARNING: Set to true in production")
		})
	}
}

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
