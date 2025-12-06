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
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("CI") == "" {
		t.Skip("skipping docker-dependent test outside CI (run with CI=true to enable)")
	}

	tmpDir := t.TempDir()
	projectName := "pgmigratetest"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/pgmigratetest",
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

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Log("Starting postgres via docker-compose...")
	composeUp, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
	composeUp.Dir = projectRoot
	defer cancel()
	output, err := composeUp.CombinedOutput()
	if err != nil {
		t.Logf("docker compose up output:\n%s", string(output))
	}
	require.NoError(t, err, "docker compose up should succeed")

	defer func() {
		t.Log("Stopping docker compose services...")
		composeDown, cancelDown := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
		defer cancelDown()
		composeDown.Dir = projectRoot
		_ = composeDown.Run()
	}()

	t.Log("Waiting for postgres to be ready...")
	time.Sleep(5 * time.Second)

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		projectName, projectName, projectName)
	envContent := fmt.Sprintf("DATABASE_URL=%s\n", dbURL)
	envPath := filepath.Join(projectRoot, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err, "should write .env file")

	t.Log("Running tracks db migrate...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate")

	output2 := stdout + stderr
	assert.True(t, strings.Contains(output2, "Successfully applied") ||
		strings.Contains(output2, "No pending migrations"),
		"should report migration status")
}

func TestDBRollback_PostgresProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("CI") == "" {
		t.Skip("skipping docker-dependent test outside CI (run with CI=true to enable)")
	}

	tmpDir := t.TempDir()
	projectName := "pgrollbacktest"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/pgrollbacktest",
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

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Log("Starting postgres via docker-compose...")
	composeUp, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
	composeUp.Dir = projectRoot
	defer cancel()
	output, err := composeUp.CombinedOutput()
	if err != nil {
		t.Logf("docker compose up output:\n%s", string(output))
	}
	require.NoError(t, err, "docker compose up should succeed")

	defer func() {
		t.Log("Stopping docker compose services...")
		composeDown, cancelDown := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
		defer cancelDown()
		composeDown.Dir = projectRoot
		_ = composeDown.Run()
	}()

	t.Log("Waiting for postgres to be ready...")
	time.Sleep(5 * time.Second)

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		projectName, projectName, projectName)
	envContent := fmt.Sprintf("DATABASE_URL=%s\n", dbURL)
	envPath := filepath.Join(projectRoot, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err, "should write .env file")

	t.Log("Running tracks db migrate first...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate")
	output2 := stdout + stderr
	t.Logf("Migrate output: %s", output2)

	t.Log("Running tracks db rollback...")
	stdout, stderr = RunCLIInDirExpectSuccess(t, projectRoot, "db", "rollback")

	output3 := stdout + stderr
	assert.True(t, strings.Contains(output3, "rolled back") ||
		strings.Contains(output3, "No migrations to roll back"),
		"should report rollback status")
}

func TestDBMigrate_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("CI") == "" {
		t.Skip("skipping docker-dependent test outside CI (run with CI=true to enable)")
	}

	tmpDir := t.TempDir()
	projectName := "pgdryruntest"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/pgdryruntest",
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

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Log("Starting postgres via docker-compose...")
	composeUp, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
	composeUp.Dir = projectRoot
	defer cancel()
	output, err := composeUp.CombinedOutput()
	if err != nil {
		t.Logf("docker compose up output:\n%s", string(output))
	}
	require.NoError(t, err, "docker compose up should succeed")

	defer func() {
		t.Log("Stopping docker compose services...")
		composeDown, cancelDown := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
		defer cancelDown()
		composeDown.Dir = projectRoot
		_ = composeDown.Run()
	}()

	t.Log("Waiting for postgres to be ready...")
	time.Sleep(5 * time.Second)

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		projectName, projectName, projectName)
	envContent := fmt.Sprintf("DATABASE_URL=%s\n", dbURL)
	envPath := filepath.Join(projectRoot, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err, "should write .env file")

	t.Log("Running tracks db migrate --dry-run...")
	stdout, stderr := RunCLIInDirExpectSuccess(t, projectRoot, "db", "migrate", "--dry-run")

	output2 := stdout + stderr
	assert.True(t, strings.Contains(output2, "Dry run") ||
		strings.Contains(output2, "No pending migrations"),
		"should show dry run output or no pending migrations")
}
