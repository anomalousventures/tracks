package template

import (
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderMigrationTemplate(t *testing.T, driver string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)

	templatePath := "internal/db/migrations/sqlite/initial_schema.sql.tmpl"
	if driver == "postgres" {
		templatePath = "internal/db/migrations/postgres/initial_schema.sql.tmpl"
	}

	data := TemplateData{
		ModuleName:         "github.com/test/app",
		MigrationTimestamp: "20251130143022",
	}

	result, err := renderer.Render(templatePath, data)
	require.NoError(t, err)
	return result
}

func TestMigrationSQLTemplateRenders(t *testing.T) {
	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite", "sqlite"},
		{"postgres", "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := renderMigrationTemplate(t, tt.driver)
			assert.NotEmpty(t, result, "template should render")
		})
	}
}

func TestMigrationSQLGooseAnnotations(t *testing.T) {
	tests := []string{"sqlite", "postgres"}

	for _, driver := range tests {
		t.Run(driver, func(t *testing.T) {
			result := renderMigrationTemplate(t, driver)

			assert.Contains(t, result, "-- +goose Up", "should have +goose Up annotation")
			assert.Contains(t, result, "-- +goose Down", "should have +goose Down annotation")
			assert.Contains(t, result, "-- +goose StatementBegin", "should have +goose StatementBegin")
			assert.Contains(t, result, "-- +goose StatementEnd", "should have +goose StatementEnd")
		})
	}
}

func TestMigrationSQLReversible(t *testing.T) {
	tests := []string{"sqlite", "postgres"}

	for _, driver := range tests {
		t.Run(driver, func(t *testing.T) {
			result := renderMigrationTemplate(t, driver)

			upIdx := strings.Index(result, "-- +goose Up")
			downIdx := strings.Index(result, "-- +goose Down")

			require.NotEqual(t, -1, upIdx, "should have Up section")
			require.NotEqual(t, -1, downIdx, "should have Down section")
			assert.Less(t, upIdx, downIdx, "Up section should come before Down section")
		})
	}
}

func TestMigrationSQLSQLitePatterns(t *testing.T) {
	result := renderMigrationTemplate(t, "sqlite")

	assert.Contains(t, result, "AFTER UPDATE", "should document AFTER UPDATE trigger pattern")
	assert.Contains(t, result, "CURRENT_TIMESTAMP", "should use CURRENT_TIMESTAMP")
	assert.Contains(t, result, "TEXT PRIMARY KEY", "should document TEXT for UUIDv7")
}

func TestMigrationSQLPostgresPatterns(t *testing.T) {
	result := renderMigrationTemplate(t, "postgres")

	assert.Contains(t, result, "update_updated_at_column()", "should create trigger function")
	assert.Contains(t, result, "RETURNS TRIGGER", "should return TRIGGER type")
	assert.Contains(t, result, "BEFORE UPDATE", "should document BEFORE UPDATE trigger pattern")
	assert.Contains(t, result, "TIMESTAMPTZ", "should use TIMESTAMPTZ")
	assert.Contains(t, result, "NOW()", "should use NOW()")
	assert.Contains(t, result, "plpgsql", "should use plpgsql language")
}

func TestMigrationSQLPostgresDownDropsFunction(t *testing.T) {
	result := renderMigrationTemplate(t, "postgres")

	downIdx := strings.Index(result, "-- +goose Down")
	require.NotEqual(t, -1, downIdx)

	downSection := result[downIdx:]
	assert.Contains(t, downSection, "DROP FUNCTION", "Down should drop the function")
	assert.Contains(t, downSection, "update_updated_at_column", "Down should reference the function name")
}

func TestMigrationSQLSQLiteNoPostgresPatterns(t *testing.T) {
	result := renderMigrationTemplate(t, "sqlite")

	assert.NotContains(t, result, "TIMESTAMPTZ", "should not use PostgreSQL TIMESTAMPTZ")
	assert.NotContains(t, result, "plpgsql", "should not use plpgsql")
	assert.NotContains(t, result, "EXECUTE FUNCTION", "should not use PostgreSQL trigger syntax")
}

func TestMigrationSQLPostgresNoSQLitePatterns(t *testing.T) {
	result := renderMigrationTemplate(t, "postgres")

	assert.NotContains(t, result, "AFTER UPDATE ON", "should not use SQLite AFTER UPDATE trigger pattern")
}
