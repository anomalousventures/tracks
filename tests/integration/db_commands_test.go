// Integration tests for tracks db commands.
//
//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPostgresProject generates a Postgres project, starts Docker, and returns
// the project root path along with a cleanup function. Skips the test if not in CI
// or if running in short mode.
func setupPostgresProject(t *testing.T, projectName string) (projectRoot string, cleanup func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("CI") == "" {
		t.Skip("skipping docker-dependent test outside CI (run with CI=true to enable)")
	}

	tmpDir := t.TempDir()

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     fmt.Sprintf("github.com/test/%s", projectName),
		DatabaseDriver: "postgres",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	t.Log("Generating postgres project...")
	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot = filepath.Join(tmpDir, projectName)

	t.Log("Starting postgres via docker-compose...")
	composeUp, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
	composeUp.Dir = projectRoot
	defer cancel()
	output, err := composeUp.CombinedOutput()
	if err != nil {
		t.Logf("docker compose up output:\n%s", string(output))
	}
	require.NoError(t, err, "docker compose up should succeed")

	cleanup = func() {
		t.Log("Stopping docker compose services...")
		composeDown, cancelDown := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
		defer cancelDown()
		composeDown.Dir = projectRoot
		_ = composeDown.Run()
	}

	t.Log("Waiting for postgres to be ready...")
	time.Sleep(5 * time.Second)

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		projectName, projectName, projectName)
	envContent := fmt.Sprintf("DATABASE_URL=%s\n", dbURL)
	envPath := filepath.Join(projectRoot, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err, "should write .env file")

	return projectRoot, cleanup
}

func TestDBCommand_Help(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "db", "--help")

	AssertContains(t, stdout, "Database management commands")
	AssertContains(t, stdout, "migrate")
	AssertContains(t, stdout, "rollback")
}

func TestDBMigrate_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, tmpDir, "db", "migrate")

	output := stdout + stderr
	AssertContains(t, output, "not in a Tracks project directory")
}

func TestDBRollback_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, tmpDir, "db", "rollback")

	output := stdout + stderr
	AssertContains(t, output, "not in a Tracks project directory")
}

func TestDBMigrate_UnsupportedDriver_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "sqliteapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/sqliteapp",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "migrate")

	output := stdout + stderr
	AssertContains(t, output, "only supports Postgres")
	AssertContains(t, output, "make migrate-up")
}

func TestDBRollback_UnsupportedDriver_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "sqliteapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/sqliteapp",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "rollback")

	output := stdout + stderr
	AssertContains(t, output, "only supports Postgres")
	AssertContains(t, output, "make migrate-down")
}

func TestDBMigrate_UnsupportedDriver_LibSQL(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "libsqlapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/libsqlapp",
		DatabaseDriver: "go-libsql",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "migrate")

	output := stdout + stderr
	AssertContains(t, output, "only supports Postgres")
}

func TestDBMigrate_MissingDatabaseURL(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "pgapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/pgapp",
		DatabaseDriver: "postgres",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	envPath := filepath.Join(projectRoot, ".env")
	_ = os.Remove(envPath)

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "migrate")

	output := stdout + stderr
	AssertContains(t, output, "DATABASE_URL")
}

func TestDBMigrate_PostgresProject(t *testing.T) {
	projectRoot, cleanup := setupPostgresProject(t, "pgmigratetest")
	defer cleanup()

	t.Log("Running tracks db migrate...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate")

	output := stdout + stderr
	assert.True(t, strings.Contains(output, "Successfully applied") ||
		strings.Contains(output, "No pending migrations"),
		"should report migration status")
}

func TestDBRollback_PostgresProject(t *testing.T) {
	projectRoot, cleanup := setupPostgresProject(t, "pgrollbacktest")
	defer cleanup()

	t.Log("Running tracks db migrate first...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate")
	t.Logf("Migrate output: %s", stdout+stderr)

	t.Log("Running tracks db rollback...")
	stdout, stderr = RunCLIInDirExpectSuccess(t, projectRoot, "db", "rollback")

	output := stdout + stderr
	assert.True(t, strings.Contains(output, "rolled back") ||
		strings.Contains(output, "No migrations to roll back"),
		"should report rollback status")
}

func TestDBMigrate_DryRun(t *testing.T) {
	projectRoot, cleanup := setupPostgresProject(t, "pgdryruntest")
	defer cleanup()

	t.Log("Running tracks db migrate --dry-run...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate", "--dry-run")

	output := stdout + stderr
	assert.True(t, strings.Contains(output, "Dry run") ||
		strings.Contains(output, "No pending migrations"),
		"should show dry run output or no pending migrations")
}
