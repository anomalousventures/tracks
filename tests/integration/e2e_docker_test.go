//go:build docker

package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type databaseSetup struct {
	dbURL          string
	composeStarted bool
	cleanupFunc    func()
}

func buildDockerImage(t *testing.T, projectRoot, imageName string) func() {
	t.Helper()

	buildCmd, cancel := dockerCmdWithTimeout(e2eTimeout, "build", "-t", imageName, ".")
	buildCmd.Dir = projectRoot
	defer cancel()
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("docker build output:\n%s", string(output))
	}
	require.NoError(t, err, "docker build should succeed")

	return func() {
		rmiCmd, cancel := dockerCmdWithTimeout(shortTimeout, "rmi", "-f", imageName)
		defer cancel()
		_ = rmiCmd.Run()
	}
}

func scanDockerImage(t *testing.T, imageName string) {
	t.Helper()

	trivyCmd, cancel := dockerCmdWithTimeout(e2eTimeout, "run", "--rm",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"aquasec/trivy:latest", "image",
		"--severity", "CRITICAL,HIGH",
		"--exit-code", "0",
		imageName)
	defer cancel()
	output, err := trivyCmd.CombinedOutput()
	if err != nil {
		t.Logf("trivy scan output:\n%s", string(output))
	}
	assert.NoError(t, err, "trivy scan should complete successfully")
}

func setupDatabaseForDocker(t *testing.T, driver, tmpDir, projectRoot, projectName string) databaseSetup {
	t.Helper()

	setup := databaseSetup{}

	switch driver {
	case "sqlite3":
		dataDir := filepath.Join(tmpDir, "data")
		err := os.MkdirAll(dataDir, 0755)
		require.NoError(t, err, "should create data directory")
		setup.dbURL = "file:/app/data/test.db"
		setup.cleanupFunc = func() {}

	case "go-libsql", "postgres":
		composeUpCmd, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel()
		output, err := composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker compose up should succeed")
		setup.composeStarted = true

		setup.cleanupFunc = func() {
			composeDownCmd, cancel := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}

		time.Sleep(5 * time.Second)

		if driver == "go-libsql" {
			setup.dbURL = "http://libsql:8080"
		} else {
			setup.dbURL = fmt.Sprintf("postgres://%s:%s@postgres:5432/%s?sslmode=disable", projectName, projectName, projectName)
		}
	}

	return setup
}

func startDockerContainer(t *testing.T, containerName, imageName, dbURL, driver, tmpDir, projectName string) func() {
	t.Helper()

	var runArgs []string
	if driver == "sqlite3" {
		dataDir := filepath.Join(tmpDir, "data")
		runArgs = []string{
			"run", "-d",
			"--name", containerName,
			"-p", "18082:8080",
			"-e", fmt.Sprintf("APP_DATABASE_URL=%s", dbURL),
			"-v", fmt.Sprintf("%s:/app/data", dataDir),
			imageName,
		}
	} else {
		networkName := fmt.Sprintf("%s_default", projectName)
		runArgs = []string{
			"run", "-d",
			"--name", containerName,
			"--network", networkName,
			"-p", "18082:8080",
			"-e", fmt.Sprintf("APP_DATABASE_URL=%s", dbURL),
			imageName,
		}
	}

	runCmd, cancel := dockerCmdWithTimeout(mediumTimeout, runArgs...)
	defer cancel()
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Logf("docker run output:\n%s", string(output))
	}
	require.NoError(t, err, "docker run should succeed")

	time.Sleep(3 * time.Second)

	return func() {
		stopCmd, cancel := dockerCmdWithTimeout(shortTimeout, "stop", containerName)
		defer cancel()
		_ = stopCmd.Run()

		rmCmd, cancel2 := dockerCmdWithTimeout(shortTimeout, "rm", "-f", containerName)
		defer cancel2()
		_ = rmCmd.Run()
	}
}

func waitForHealthEndpoint(t *testing.T, containerName string, port int, maxRetries int) {
	t.Helper()

	var resp *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/health", port))
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if i < maxRetries-1 {
			t.Logf("Health check attempt %d/%d failed: %v (retrying...)", i+1, maxRetries, err)
			time.Sleep(1 * time.Second)
		} else {
			logsCmd, cancel := dockerCmdWithTimeout(shortTimeout, "logs", containerName)
			defer cancel()
			if logs, logErr := logsCmd.CombinedOutput(); logErr == nil {
				t.Logf("Container logs:\n%s", string(logs))
			}
			t.Logf("Health check attempt %d/%d failed: %v", i+1, maxRetries, err)
		}
	}

	assert.NoError(t, err, "health endpoint should be accessible after retries")
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "health endpoint should return 200 OK")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "should return JSON")
	}
}

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
