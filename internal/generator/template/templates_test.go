package template

import (
	"go/parser"
	"go/token"
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

func renderMainGoTemplate(t *testing.T, moduleName string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{ModuleName: moduleName}
	result, err := renderer.Render("cmd/server/main.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestMainGoTemplate(t *testing.T) {
	result := renderMainGoTemplate(t, "github.com/test/app")

	tests := []struct {
		name     string
		contains string
		message  string
	}{
		{"package declaration", "package main", "should have package main"},
		{"import block", "import", "should have import block"},
		{"main function", "func main()", "should have main() function"},
		{"run function", "func run() error", "should have run() error function"},
		{"main calls run", "if err := run(); err != nil", "main() should call run() and check error"},
		{"error to stderr", `fmt.Fprintf(os.Stderr, "error: %v\n", err)`, "should print errors to stderr"},
		{"exit on error", "os.Exit(1)", "should exit with status 1 on error"},
		{"config load", "cfg, err := config.Load()", "should load config"},
		{"config error wrap", `return fmt.Errorf("load config: %w", err)`, "should wrap config load error"},
		{"logger init", "logger := logging.NewLogger(cfg.Environment)", "should initialize logger"},
		{"server start log", `logger.Info(ctx).Msg("server starting")`, "should log server start"},
		{"db connection", "database, err := db.New(ctx, cfg.Database)", "should connect to database"},
		{"db error wrap", `return fmt.Errorf("connect to database: %w", err)`, "should wrap database connection error"},
		{"db cleanup", "database.Close()", "should close database"},
		{"health service", "healthService := health.NewService(healthRepo)", "should instantiate health service with repository"},
		{"server builder", "http.NewServer(cfg, logger)", "should use NewServer constructor"},
		{"with health", "WithHealthService(healthService)", "should chain WithHealthService"},
		{"register routes", "RegisterRoutes()", "should chain RegisterRoutes"},
		{"listen and serve", "return srv.ListenAndServe()", "run() should return srv.ListenAndServe()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, result, tt.contains, tt.message)
		})
	}

	t.Run("wrong packages excluded", func(t *testing.T) {
		assert.NotContains(t, result, "package http", "should not have 'http' package")
		assert.NotContains(t, result, "package server", "should not have 'server' package")
	})
}

func TestMainGoValidGoCode(t *testing.T) {
	result := renderMainGoTemplate(t, "github.com/test/app")

	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "main.go", result, parser.AllErrors)
	require.NoError(t, err, "generated main.go should be valid Go code")
}

func TestMainGoMarkerSections(t *testing.T) {
	result := renderMainGoTemplate(t, "github.com/test/app")

	markers := []struct {
		name  string
		begin string
		end   string
	}{
		{"DB", "// TRACKS:DB:BEGIN", "// TRACKS:DB:END"},
		{"REPOSITORIES", "// TRACKS:REPOSITORIES:BEGIN", "// TRACKS:REPOSITORIES:END"},
		{"SERVICES", "// TRACKS:SERVICES:BEGIN", "// TRACKS:SERVICES:END"},
	}

	for _, m := range markers {
		t.Run(m.name, func(t *testing.T) {
			assert.Contains(t, result, m.begin, "should have %s marker begin", m.name)
			assert.Contains(t, result, m.end, "should have %s marker end", m.name)
		})
	}

	t.Run("REPOSITORIES contains health repo", func(t *testing.T) {
		beginIdx := strings.Index(result, "// TRACKS:REPOSITORIES:BEGIN")
		endIdx := strings.Index(result, "// TRACKS:REPOSITORIES:END")
		require.NotEqual(t, -1, beginIdx, "should have REPOSITORIES begin marker")
		require.NotEqual(t, -1, endIdx, "should have REPOSITORIES end marker")

		section := result[beginIdx:endIdx]
		assert.Contains(t, section, "queries := generated.New(database)", "should create queries")
		assert.Contains(t, section, "healthRepo := health.NewRepository(queries)", "should create health repository")
	})

	t.Run("marker order", func(t *testing.T) {
		dbIdx := strings.Index(result, "// TRACKS:DB:BEGIN")
		repoIdx := strings.Index(result, "// TRACKS:REPOSITORIES:BEGIN")
		servicesIdx := strings.Index(result, "// TRACKS:SERVICES:BEGIN")

		assert.Greater(t, repoIdx, dbIdx, "REPOSITORIES marker should come after DB marker")
		assert.Greater(t, servicesIdx, repoIdx, "SERVICES marker should come after REPOSITORIES marker")
	})
}

func TestMainGoImports(t *testing.T) {
	result := renderMainGoTemplate(t, "github.com/test/app")

	imports := []string{
		`"context"`,
		`"fmt"`,
		`"os"`,
		"github.com/test/app/internal/config",
		"github.com/test/app/internal/db",
		"github.com/test/app/internal/domain/health",
		"github.com/test/app/internal/http",
		"github.com/test/app/internal/logging",
	}

	for _, imp := range imports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestMainGoModuleNameInterpolation(t *testing.T) {
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
			result := renderMainGoTemplate(t, tt.moduleName)

			imports := []string{
				tt.moduleName + "/internal/config",
				tt.moduleName + "/internal/db",
				tt.moduleName + "/internal/http",
				tt.moduleName + "/internal/logging",
				tt.moduleName + "/internal/domain/health",
			}

			for _, imp := range imports {
				assert.Contains(t, result, imp, "should interpolate module name in import")
			}
		})
	}
}

func TestTracksYamlTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name        string
		projectName string
		moduleName  string
		dbDriver    string
	}{
		{
			name:        "go-libsql driver",
			projectName: "myapp",
			moduleName:  "github.com/user/myapp",
			dbDriver:    "go-libsql",
		},
		{
			name:        "sqlite3 driver",
			projectName: "testapp",
			moduleName:  "github.com/user/testapp",
			dbDriver:    "sqlite3",
		},
		{
			name:        "postgres driver",
			projectName: "webapp",
			moduleName:  "github.com/user/webapp",
			dbDriver:    "postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ProjectName: tt.projectName,
				ModuleName:  tt.moduleName,
				DBDriver:    tt.dbDriver,
			}

			result, err := renderer.Render(".tracks.yaml.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "# Tracks CLI Project Metadata")
			assert.Contains(t, result, "schema_version: \"1.0\"")

			assert.Contains(t, result, "project:")
			assert.Contains(t, result, "name: \""+tt.projectName+"\"")
			assert.Contains(t, result, "module_path: \""+tt.moduleName+"\"")
			assert.Contains(t, result, "tracks_version: \"dev\"")
			assert.Contains(t, result, "last_upgraded_version: \"dev\"")
			assert.Contains(t, result, "database_driver: \""+tt.dbDriver+"\"")
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
		{".tracks.yaml", ".tracks.yaml.tmpl"},
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
		{".tracks.yaml", ".tracks.yaml.tmpl"},
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
