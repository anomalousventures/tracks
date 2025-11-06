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

func renderDBTemplate(t *testing.T, driver string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/test/app",
		DBDriver:   driver,
	}
	result, err := renderer.Render("internal/db/db.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestDBTemplate(t *testing.T) {
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
			result := renderDBTemplate(t, tt.driver)

			assert.Contains(t, result, "package db", "should have package db")
			assert.Contains(t, result, "func New(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error)", "should have New function with correct signature")
			assert.Contains(t, result, "db.PingContext(", "should call PingContext to verify connection")
			assert.Contains(t, result, "db.SetMaxOpenConns(cfg.MaxOpenConns)", "should configure connection pool")
			assert.NotEmpty(t, result, "template should render")
		})
	}
}

func TestDBValidGoCode(t *testing.T) {
	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			result := renderDBTemplate(t, driver)

			fset := token.NewFileSet()
			_, err := parser.ParseFile(fset, "db.go", result, parser.AllErrors)
			require.NoError(t, err, "generated db.go for %s should be valid Go code", driver)
		})
	}
}

func TestDBConditionalImports(t *testing.T) {
	tests := []struct {
		driver          string
		wantImport      string
		excludeImports  []string
	}{
		{
			driver:     "go-libsql",
			wantImport: `_ "github.com/tursodatabase/libsql-client-go/libsql"`,
			excludeImports: []string{
				`_ "github.com/mattn/go-sqlite3"`,
				`_ "github.com/lib/pq"`,
			},
		},
		{
			driver:     "sqlite3",
			wantImport: `_ "github.com/mattn/go-sqlite3"`,
			excludeImports: []string{
				`_ "github.com/tursodatabase/libsql-client-go/libsql"`,
				`_ "github.com/lib/pq"`,
			},
		},
		{
			driver:     "postgres",
			wantImport: `_ "github.com/lib/pq"`,
			excludeImports: []string{
				`_ "github.com/tursodatabase/libsql-client-go/libsql"`,
				`_ "github.com/mattn/go-sqlite3"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderDBTemplate(t, tt.driver)

			assert.Contains(t, result, tt.wantImport, "should import correct driver for %s", tt.driver)

			for _, exclude := range tt.excludeImports {
				assert.NotContains(t, result, exclude, "should not import %s for driver %s", exclude, tt.driver)
			}
		})
	}
}

func TestDBDriverNames(t *testing.T) {
	tests := []struct {
		driver       string
		wantOpenCall string
	}{
		{"go-libsql", `sql.Open("libsql", cfg.URL)`},
		{"sqlite3", `sql.Open("sqlite3", cfg.URL)`},
		{"postgres", `sql.Open("postgres", cfg.URL)`},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderDBTemplate(t, tt.driver)

			assert.Contains(t, result, tt.wantOpenCall, "should use correct driver name in sql.Open for %s", tt.driver)
		})
	}
}

func TestDBErrorHandling(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	errors := []struct {
		name    string
		pattern string
	}{
		{"open error", `fmt.Errorf("open database: %w", err)`},
		{"ping error", `fmt.Errorf("ping database: %w", err)`},
	}

	for _, e := range errors {
		t.Run(e.name, func(t *testing.T) {
			assert.Contains(t, result, e.pattern, "should have %s with proper wrapping", e.name)
		})
	}
}

func TestDBConnectionVerification(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	assert.Contains(t, result, "db.PingContext(", "should call PingContext to verify connection")
	assert.Contains(t, result, "context.WithTimeout(ctx, cfg.ConnectTimeout)", "should use timeout for ping")

	openIdx := strings.Index(result, "sql.Open(")
	pingIdx := strings.Index(result, "db.PingContext(")

	require.NotEqual(t, -1, openIdx, "should have sql.Open call")
	require.NotEqual(t, -1, pingIdx, "should have PingContext call")
	assert.Greater(t, pingIdx, openIdx, "PingContext should be called after Open")
}

func TestDBFunctionSignature(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	tests := []struct {
		name     string
		contains string
	}{
		{"function name", "func New("},
		{"context parameter", "ctx context.Context"},
		{"config parameter", "cfg config.DatabaseConfig"},
		{"return pointer", "(*sql.DB"},
		{"return error", "error)"},
		{"full signature", "func New(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, result, tt.contains, "signature should contain %s", tt.name)
		})
	}
}

func TestDBImports(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	imports := []string{
		`"context"`,
		`"database/sql"`,
		`"fmt"`,
		`"github.com/test/app/internal/config"`,
	}

	for _, imp := range imports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestDBConnectionPoolConfiguration(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	poolSettings := []string{
		"db.SetMaxOpenConns(cfg.MaxOpenConns)",
		"db.SetMaxIdleConns(cfg.MaxIdleConns)",
		"db.SetConnMaxLifetime(cfg.ConnMaxLifetime)",
	}

	for _, setting := range poolSettings {
		assert.Contains(t, result, setting, "should configure %s", setting)
	}
}

func TestDBPackageDeclaration(t *testing.T) {
	result := renderDBTemplate(t, "postgres")

	assert.Contains(t, result, "package db", "should have package db declaration")
	assert.NotContains(t, result, "package database", "should not have package database")
	assert.NotContains(t, result, "package main", "should not have package main")
}
