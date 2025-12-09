// Integration tests for tracks db commands.
//
//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/require"
)

func TestDBCommand_Help(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "db", "--help")

	AssertContains(t, stdout, "Database management commands")
	AssertContains(t, stdout, "migrate")
	AssertContains(t, stdout, "rollback")
	AssertContains(t, stdout, "status")
	AssertContains(t, stdout, "reset")
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

func TestDBStatus_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, tmpDir, "db", "status")

	output := stdout + stderr
	AssertContains(t, output, "not in a Tracks project directory")
}

func TestDBStatus_UnsupportedDriver_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "sqlitestatusapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/sqlitestatusapp",
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

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "status")

	output := stdout + stderr
	AssertContains(t, output, "only supports Postgres")
	AssertContains(t, output, "make migrate-status")
}

func TestDBReset_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, tmpDir, "db", "reset", "--force")

	output := stdout + stderr
	AssertContains(t, output, "not in a Tracks project directory")
}

func TestDBReset_UnsupportedDriver_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "sqliteresetapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/sqliteresetapp",
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

	stdout, stderr, _ := RunCLIInDirExpectFailure(t, projectRoot, "db", "reset", "--force")

	output := stdout + stderr
	AssertContains(t, output, "only supports Postgres")
	AssertContains(t, output, "make migrate-reset")
}
