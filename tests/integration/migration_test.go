package integration

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationFilesGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		databaseDriver string
		migrationDir   string
	}{
		{"go-libsql", "go-libsql", "sqlite"},
		{"sqlite3", "sqlite3", "sqlite"},
		{"postgres", "postgres", "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     "github.com/test/app",
				DatabaseDriver: tt.databaseDriver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)
			migrationDir := filepath.Join(projectRoot, "internal", "db", "migrations", tt.migrationDir)

			entries, err := os.ReadDir(migrationDir)
			require.NoError(t, err, "should be able to read migration directory")

			var migrationFile string
			timestampPattern := regexp.MustCompile(`^\d{14}_initial_schema\.sql$`)

			for _, entry := range entries {
				if timestampPattern.MatchString(entry.Name()) {
					migrationFile = entry.Name()
					break
				}
			}

			require.NotEmpty(t, migrationFile, "should find timestamped initial schema migration")

			content, err := os.ReadFile(filepath.Join(migrationDir, migrationFile))
			require.NoError(t, err)

			assert.Contains(t, string(content), "-- +goose Up", "migration should have Up section")
			assert.Contains(t, string(content), "-- +goose Down", "migration should have Down section")
			assert.Contains(t, string(content), "-- +goose StatementBegin", "migration should have StatementBegin")
			assert.Contains(t, string(content), "-- +goose StatementEnd", "migration should have StatementEnd")
		})
	}
}

func TestMigrationDriverSpecificContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		driver         string
		migrationDir   string
		mustContain    []string
		mustNotContain []string
	}{
		{
			driver:       "sqlite3",
			migrationDir: "sqlite",
			mustContain:  []string{"AFTER UPDATE", "CURRENT_TIMESTAMP"},
			mustNotContain: []string{"TIMESTAMPTZ", "plpgsql"},
		},
		{
			driver:       "postgres",
			migrationDir: "postgres",
			mustContain:  []string{"TIMESTAMPTZ", "BEFORE UPDATE", "plpgsql", "update_updated_at_column"},
			mustNotContain: []string{"AFTER UPDATE ON"},
		},
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
			require.NoError(t, err)

			migrationDir := filepath.Join(tmpDir, projectName, "internal", "db", "migrations", tt.migrationDir)
			entries, err := os.ReadDir(migrationDir)
			require.NoError(t, err)

			timestampPattern := regexp.MustCompile(`^\d{14}_initial_schema\.sql$`)
			var content []byte

			for _, entry := range entries {
				if timestampPattern.MatchString(entry.Name()) {
					content, err = os.ReadFile(filepath.Join(migrationDir, entry.Name()))
					require.NoError(t, err)
					break
				}
			}

			require.NotEmpty(t, content, "should find migration file")

			for _, expected := range tt.mustContain {
				assert.Contains(t, string(content), expected, "should contain %s", expected)
			}

			for _, excluded := range tt.mustNotContain {
				assert.NotContains(t, string(content), excluded, "should not contain %s", excluded)
			}
		})
	}
}

func TestMigrationBothDirectoriesCreated(t *testing.T) {
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
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, projectName)
	timestampPattern := regexp.MustCompile(`^\d{14}_initial_schema\.sql$`)

	for _, dir := range []string{"sqlite", "postgres"} {
		migrationDir := filepath.Join(projectRoot, "internal", "db", "migrations", dir)

		stat, err := os.Stat(migrationDir)
		require.NoError(t, err, "migration directory %s should exist", dir)
		assert.True(t, stat.IsDir(), "%s should be a directory", dir)

		entries, err := os.ReadDir(migrationDir)
		require.NoError(t, err)

		found := false
		for _, entry := range entries {
			if timestampPattern.MatchString(entry.Name()) {
				found = true
				break
			}
		}
		assert.True(t, found, "should find timestamped migration in %s directory", dir)
	}
}

func TestMigrationTimestampFormat(t *testing.T) {
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
	require.NoError(t, err)

	migrationDir := filepath.Join(tmpDir, projectName, "internal", "db", "migrations", "sqlite")
	entries, err := os.ReadDir(migrationDir)
	require.NoError(t, err)

	timestampPattern := regexp.MustCompile(`^(\d{14})_initial_schema\.sql$`)

	var timestamp string
	for _, entry := range entries {
		matches := timestampPattern.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			timestamp = matches[1]
			break
		}
	}

	require.NotEmpty(t, timestamp, "should extract timestamp from filename")
	assert.Len(t, timestamp, 14, "timestamp should be 14 digits (YYYYMMDDHHMMSS)")

	year := timestamp[0:4]
	month := timestamp[4:6]
	day := timestamp[6:8]
	hour := timestamp[8:10]
	minute := timestamp[10:12]
	second := timestamp[12:14]

	assert.Regexp(t, `^20\d{2}$`, year, "year should be valid (20XX)")
	assert.Regexp(t, `^(0[1-9]|1[0-2])$`, month, "month should be 01-12")
	assert.Regexp(t, `^(0[1-9]|[12]\d|3[01])$`, day, "day should be 01-31")
	assert.Regexp(t, `^([01]\d|2[0-3])$`, hour, "hour should be 00-23")
	assert.Regexp(t, `^[0-5]\d$`, minute, "minute should be 00-59")
	assert.Regexp(t, `^[0-5]\d$`, second, "second should be 00-59")
}
