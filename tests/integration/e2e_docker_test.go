//go:build docker

package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/require"
)

func runDockerE2ETest(t *testing.T, driver string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectName := "testapp"
	imageName := fmt.Sprintf("%s:test", projectName)

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
		DatabaseDriver: driver,
		EnvPrefix:      "APP",
		InitGit:        true,
		OutputPath:     tmpDir,
	}

	t.Log("1. Generating project...")
	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Log("2. Building Docker image...")
	cleanupImage := buildDockerImage(t, projectRoot, imageName)
	defer cleanupImage()

	t.Log("3. Scanning image with Trivy...")
	scanDockerImage(t, imageName)

	t.Logf("4. Setting up database for %s...", driver)
	dbSetup := setupDatabaseForDocker(t, driver, tmpDir, projectRoot, projectName)
	defer dbSetup.cleanupFunc()

	containerName := fmt.Sprintf("%s-test", projectName)

	t.Log("5. Starting container...")
	cleanupContainer := startDockerContainer(t, containerName, imageName, dbSetup.dbURL, driver, tmpDir, projectName)
	defer cleanupContainer()

	t.Log("6. Checking health endpoint...")
	waitForHealthEndpoint(t, containerName, 18082, 30)

	t.Log("Docker E2E test completed successfully!")
}

func TestE2E_GoLibsql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "go-libsql")
}

func TestE2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "postgres")
}

func TestDockerE2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "postgres")
}

func TestDockerE2E_GoLibsql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "go-libsql")
}

func TestDockerE2E_SQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "sqlite3")
}
