package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSqlcYamlTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestSqlcYamlValidYAML(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err, "generated YAML should be valid for %s", driver)
		})
	}
}

func TestSqlcYamlVersion(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		DBDriver: "postgres",
	}

	result, err := renderer.Render("sqlc.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	assert.Equal(t, "2", config["version"], "version should be 2")
}

func TestSqlcYamlSchemaPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		driver         string
		expectedSchema string
	}{
		{"go-libsql", "internal/db/migrations/sqlite"},
		{"sqlite3", "internal/db/migrations/sqlite"},
		{"postgres", "internal/db/migrations/postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: tt.driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")
			require.Len(t, sql, 1, "sql array should have one element")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			assert.Equal(t, tt.expectedSchema, sqlConfig["schema"], "schema path should be correct for %s", tt.driver)
		})
	}
}

func TestSqlcYamlEngine(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		driver         string
		expectedEngine string
	}{
		{"go-libsql", "sqlite"},
		{"sqlite3", "sqlite"},
		{"postgres", "postgresql"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: tt.driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			assert.Equal(t, tt.expectedEngine, sqlConfig["engine"], "engine should be %s for driver %s", tt.expectedEngine, tt.driver)
		})
	}
}

func TestSqlcYamlQueriesPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			assert.Equal(t, "internal/db/queries", sqlConfig["queries"], "queries path should be internal/db/queries for %s", driver)
		})
	}
}

func TestSqlcYamlOutputPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			gen, ok := sqlConfig["gen"].(map[string]interface{})
			require.True(t, ok, "gen should be a map")

			goGen, ok := gen["go"].(map[string]interface{})
			require.True(t, ok, "gen.go should be a map")

			assert.Equal(t, "internal/db/generated", goGen["out"], "output path should be internal/db/generated for %s", driver)
		})
	}
}

func TestSqlcYamlPackageName(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			gen, ok := sqlConfig["gen"].(map[string]interface{})
			require.True(t, ok, "gen should be a map")

			goGen, ok := gen["go"].(map[string]interface{})
			require.True(t, ok, "gen.go should be a map")

			assert.Equal(t, "generated", goGen["package"], "package name should be generated for %s", driver)
		})
	}
}

func TestSqlcYamlSqlPackage(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			gen, ok := sqlConfig["gen"].(map[string]interface{})
			require.True(t, ok, "gen should be a map")

			goGen, ok := gen["go"].(map[string]interface{})
			require.True(t, ok, "gen.go should be a map")

			assert.Equal(t, "database/sql", goGen["sql_package"], "sql_package should be database/sql for %s", driver)
		})
	}
}

func TestSqlcYamlEmitFlags(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			gen, ok := sqlConfig["gen"].(map[string]interface{})
			require.True(t, ok, "gen should be a map")

			goGen, ok := gen["go"].(map[string]interface{})
			require.True(t, ok, "gen.go should be a map")

			assert.Equal(t, true, goGen["emit_json_tags"], "emit_json_tags should be true for %s", driver)
			assert.Equal(t, true, goGen["emit_interface"], "emit_interface should be true for %s", driver)
			assert.Equal(t, true, goGen["emit_empty_slices"], "emit_empty_slices should be true for %s", driver)
		})
	}
}

func TestSqlcYamlStructure(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		DBDriver: "postgres",
	}

	result, err := renderer.Render("sqlc.yaml.tmpl", data)
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &config)
	require.NoError(t, err)

	// Top level should have version and sql
	assert.Contains(t, config, "version", "config should have version")
	assert.Contains(t, config, "sql", "config should have sql")

	// SQL should be an array with one element
	sql, ok := config["sql"].([]interface{})
	require.True(t, ok, "sql should be an array")
	require.Len(t, sql, 1, "sql array should have one element")

	// SQL element should have required fields
	sqlConfig, ok := sql[0].(map[string]interface{})
	require.True(t, ok, "sql[0] should be a map")

	requiredFields := []string{"schema", "queries", "engine", "gen"}
	for _, field := range requiredFields {
		assert.Contains(t, sqlConfig, field, "sql config should have %s field", field)
	}
}

func TestSqlcYamlOverrides(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			data := TemplateData{
				DBDriver: driver,
			}

			result, err := renderer.Render("sqlc.yaml.tmpl", data)
			require.NoError(t, err)

			var config map[string]interface{}
			err = yaml.Unmarshal([]byte(result), &config)
			require.NoError(t, err)

			sql, ok := config["sql"].([]interface{})
			require.True(t, ok, "sql should be an array")

			sqlConfig, ok := sql[0].(map[string]interface{})
			require.True(t, ok, "sql[0] should be a map")

			gen, ok := sqlConfig["gen"].(map[string]interface{})
			require.True(t, ok, "gen should be a map")

			goGen, ok := gen["go"].(map[string]interface{})
			require.True(t, ok, "gen.go should be a map")

			overrides, ok := goGen["overrides"].([]interface{})
			require.True(t, ok, "overrides should exist for %s", driver)
			require.Len(t, overrides, 1, "should have one override for %s", driver)

			override, ok := overrides[0].(map[string]interface{})
			require.True(t, ok, "override should be a map")

			assert.Equal(t, "TEXT", override["db_type"], "db_type should be TEXT for %s", driver)
			assert.Equal(t, "string", override["go_type"], "go_type should be string for %s", driver)
			assert.Equal(t, false, override["nullable"], "nullable should be false for %s", driver)
		})
	}
}
