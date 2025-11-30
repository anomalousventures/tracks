package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderMigrateTemplate(t *testing.T, driver string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/test/app",
		DBDriver:   driver,
	}
	result, err := renderer.Render("internal/db/migrate.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestMigrateTemplateRenders(t *testing.T) {
	drivers := []struct {
		name   string
		driver string
	}{
		{"go-libsql", "go-libsql"},
		{"sqlite3", "sqlite3"},
		{"postgres", "postgres"},
	}

	for _, tt := range drivers {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderMigrateTemplate(t, tt.driver)

			assert.Contains(t, result, "package db", "should have package db")
			assert.NotEmpty(t, result, "template should render")
		})
	}
}

func TestMigrateTemplateValidGoCode(t *testing.T) {
	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			result := renderMigrateTemplate(t, driver)

			fset := token.NewFileSet()
			_, err := parser.ParseFile(fset, "migrate.go", result, parser.AllErrors)
			require.NoError(t, err, "generated migrate.go for %s should be valid Go code", driver)
		})
	}
}

func TestMigrateTemplateDialectMapping(t *testing.T) {
	tests := []struct {
		driver      string
		wantDialect string
		wantDir     string
	}{
		{"go-libsql", `dialect       = "sqlite3"`, "migrations/sqlite"},
		{"sqlite3", `dialect       = "sqlite3"`, "migrations/sqlite"},
		{"postgres", `dialect       = "postgres"`, "migrations/postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderMigrateTemplate(t, tt.driver)

			assert.Contains(t, result, tt.wantDialect, "should use correct dialect for %s", tt.driver)
			assert.Contains(t, result, tt.wantDir, "should use correct migrations directory for %s", tt.driver)
		})
	}
}

func TestMigrateTemplateFunctions(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	functions := []struct {
		name      string
		signature string
	}{
		{"MigrateUp", "func MigrateUp(ctx context.Context, db *sql.DB) (*MigrationResult, error)"},
		{"MigrateDown", "func MigrateDown(ctx context.Context, db *sql.DB) (*MigrationResult, error)"},
		{"MigrateStatus", "func MigrateStatus(ctx context.Context, db *sql.DB) ([]MigrationStatus, error)"},
		{"MigrateTo", "func MigrateTo(ctx context.Context, db *sql.DB, version int64) (*MigrationResult, error)"},
		{"GetDialect", "func GetDialect() string"},
	}

	for _, fn := range functions {
		t.Run(fn.name, func(t *testing.T) {
			assert.Contains(t, result, fn.signature, "should have %s function with correct signature", fn.name)
		})
	}
}

func TestMigrateTemplateTypes(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	assert.Contains(t, result, "type MigrationStatus struct", "should have MigrationStatus type")
	assert.Contains(t, result, "type MigrationResult struct", "should have MigrationResult type")

	statusFields := []string{
		"Version   int64",
		"Name      string",
		"AppliedAt *time.Time",
		"IsPending bool",
	}
	for _, field := range statusFields {
		assert.Contains(t, result, field, "MigrationStatus should have field %s", field)
	}

	resultFields := []string{
		`Direction   string`,
		"FromVersion int64",
		"ToVersion   int64",
		"Applied     []MigrationStatus",
	}
	for _, field := range resultFields {
		assert.Contains(t, result, field, "MigrationResult should have field %s", field)
	}
}

func TestMigrateTemplateEmbedDirectives(t *testing.T) {
	tests := []struct {
		driver    string
		wantEmbed string
		excludes  []string
	}{
		{
			driver:    "go-libsql",
			wantEmbed: "//go:embed migrations/sqlite/*.sql",
			excludes:  []string{"//go:embed migrations/postgres/*.sql"},
		},
		{
			driver:    "sqlite3",
			wantEmbed: "//go:embed migrations/sqlite/*.sql",
			excludes:  []string{"//go:embed migrations/postgres/*.sql"},
		},
		{
			driver:    "postgres",
			wantEmbed: "//go:embed migrations/postgres/*.sql",
			excludes:  []string{"//go:embed migrations/sqlite/*.sql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderMigrateTemplate(t, tt.driver)

			assert.Contains(t, result, tt.wantEmbed, "should have correct embed directive for %s", tt.driver)

			for _, exclude := range tt.excludes {
				assert.NotContains(t, result, exclude, "should not have %s for driver %s", exclude, tt.driver)
			}
		})
	}
}

func TestMigrateTemplateImports(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	imports := []string{
		`"context"`,
		`"database/sql"`,
		`"embed"`,
		`"fmt"`,
		`"io/fs"`,
		`"time"`,
		`"github.com/pressly/goose/v3"`,
	}

	for _, imp := range imports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestMigrateTemplateGooseProviderAPI(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	assert.Contains(t, result, "goose.NewProvider", "should use Goose Provider API")
	assert.Contains(t, result, "goose.Dialect(dialect)", "should pass dialect to provider")
	assert.Contains(t, result, "provider.Up(ctx)", "should use provider for Up")
	assert.Contains(t, result, "provider.Down(ctx)", "should use provider for Down")
	assert.Contains(t, result, "provider.Status(ctx)", "should use provider for Status")
	assert.Contains(t, result, "provider.UpTo(ctx, version)", "should use provider for UpTo")
	assert.Contains(t, result, "provider.GetDBVersion(ctx)", "should use provider to get version")
}

func TestMigrateTemplatePackageDeclaration(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	assert.Contains(t, result, "package db", "should have package db declaration")
	assert.NotContains(t, result, "package migrate", "should not have package migrate")
	assert.NotContains(t, result, "package main", "should not have package main")
}

func TestMigrateTemplateErrorHandling(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	errors := []struct {
		name    string
		pattern string
	}{
		{"filesystem error", `fmt.Errorf("create migrations filesystem: %w", err)`},
		{"provider error", `fmt.Errorf("create migration provider: %w", err)`},
		{"current version error", `fmt.Errorf("get current version: %w", err)`},
		{"apply migrations error", `fmt.Errorf("apply migrations: %w", err)`},
		{"rollback error", `fmt.Errorf("rollback migration: %w", err)`},
		{"status error", `fmt.Errorf("get migration status: %w", err)`},
	}

	for _, e := range errors {
		t.Run(e.name, func(t *testing.T) {
			assert.Contains(t, result, e.pattern, "should have %s with proper wrapping", e.name)
		})
	}
}

func TestMigrateTemplateAppliedAtHandling(t *testing.T) {
	result := renderMigrateTemplate(t, "postgres")

	assert.Contains(t, result, "!s.AppliedAt.IsZero()", "should check for zero time instead of Valid field")
	assert.NotContains(t, result, "s.AppliedAt.Valid", "should not use Valid field (that's sql.NullTime, not time.Time)")
}
