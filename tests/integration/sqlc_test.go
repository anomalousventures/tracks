package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthQueryGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     "github.com/test/app",
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			err := gen.Generate(context.Background(), cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)
			healthSQL := filepath.Join(projectRoot, "internal", "db", "queries", "health.sql")

			content, err := os.ReadFile(healthSQL)
			require.NoError(t, err, "health.sql should exist")

			assert.Contains(t, string(content), "-- name: HealthCheck :one", "should have SQLC annotation")
			assert.Contains(t, string(content), "SELECT 1 as healthy", "should have health check query")
		})
	}
}

func TestGeneratedDirectoryCreated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "testapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/app",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	err := gen.Generate(context.Background(), cfg)
	require.NoError(t, err, "generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)
	generatedDir := filepath.Join(projectRoot, "internal", "db", "generated")

	stat, err := os.Stat(generatedDir)
	require.NoError(t, err, "generated directory should exist")
	assert.True(t, stat.IsDir(), "should be a directory")

	entries, err := os.ReadDir(generatedDir)
	require.NoError(t, err, "should be able to read generated directory")
	require.NotEmpty(t, entries, "generated directory should contain SQLC-generated files")

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}

	assert.Contains(t, fileNames, "db.go", "should contain db.go from SQLC")
	assert.Contains(t, fileNames, "health.sql.go", "should contain health.sql.go from SQLC")
}

func TestSqlcConfigGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		driver         string
		expectedEngine string
		expectedSchema string
	}{
		{"go-libsql", "sqlite", "internal/db/migrations/sqlite"},
		{"sqlite3", "sqlite", "internal/db/migrations/sqlite"},
		{"postgres", "postgresql", "internal/db/migrations/postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     "github.com/test/app",
				DatabaseDriver: tt.driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			err := gen.Generate(context.Background(), cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)
			sqlcYaml := filepath.Join(projectRoot, "sqlc.yaml")

			content, err := os.ReadFile(sqlcYaml)
			require.NoError(t, err, "sqlc.yaml should exist")

			assert.Contains(t, string(content), "engine: \""+tt.expectedEngine+"\"", "should have correct engine for %s", tt.driver)
			assert.Contains(t, string(content), "schema: \""+tt.expectedSchema+"\"", "should have correct schema path for %s", tt.driver)
			assert.Contains(t, string(content), "queries: \"internal/db/queries\"", "should have queries path")
			assert.Contains(t, string(content), "out: \"internal/db/generated\"", "should have output path")
		})
	}
}

func TestQueriesDirectoryStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "testapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/app",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	err := gen.Generate(context.Background(), cfg)
	require.NoError(t, err, "generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	queriesDir := filepath.Join(projectRoot, "internal", "db", "queries")
	stat, err := os.Stat(queriesDir)
	require.NoError(t, err, "queries directory should exist")
	assert.True(t, stat.IsDir(), "should be a directory")

	entries, err := os.ReadDir(queriesDir)
	require.NoError(t, err)

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}

	assert.Contains(t, fileNames, "health.sql", "should contain health.sql")
	assert.Contains(t, fileNames, ".gitkeep", "should contain .gitkeep")
}
