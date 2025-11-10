package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTimeout(envVar string, defaultTimeout time.Duration) time.Duration {
	if val := os.Getenv(envVar); val != "" {
		if timeout, err := time.ParseDuration(val); err == nil && timeout > 0 {
			return timeout
		}
	}
	return defaultTimeout
}

func cmdWithTimeout(timeout time.Duration, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, cancel
}

func dockerCmdWithTimeout(timeout time.Duration, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		wslArgs := append([]string{"--", "docker"}, args...)
		cmd = exec.CommandContext(ctx, "wsl", wslArgs...)
	} else {
		cmd = exec.CommandContext(ctx, "docker", args...)
	}
	return cmd, cancel
}

var (
	shortTimeout  = getTimeout("INTEGRATION_TEST_SHORT_TIMEOUT", 2*time.Second)
	mediumTimeout = getTimeout("INTEGRATION_TEST_MEDIUM_TIMEOUT", 10*time.Second)
	longTimeout   = getTimeout("INTEGRATION_TEST_LONG_TIMEOUT", 30*time.Second)
	e2eTimeout    = getTimeout("INTEGRATION_TEST_E2E_TIMEOUT", 180*time.Second)
)

func runE2ETest(t *testing.T, driver string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectName := "testapp"

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

	t.Log("2. Verifying go mod tidy is idempotent...")
	tidyCmd, cancel := cmdWithTimeout(longTimeout, "go", "mod", "tidy")
	tidyCmd.Dir = projectRoot
	defer cancel()
	output, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output:\n%s", string(output))
	}
	require.NoError(t, err, "go mod tidy should succeed")

	gitStatusCmd, cancel2 := cmdWithTimeout(shortTimeout, "git", "status", "--porcelain")
	gitStatusCmd.Dir = projectRoot
	defer cancel2()
	statusOutput, err := gitStatusCmd.CombinedOutput()
	require.NoError(t, err, "git status should succeed")
	statusStr := strings.TrimSpace(string(statusOutput))
	assert.Empty(t, statusStr, "go mod tidy should be idempotent (no changes after generation)")

	t.Log("3. Verifying make generate is idempotent...")
	generateCmd, cancel3 := cmdWithTimeout(e2eTimeout, "make", "generate")
	generateCmd.Dir = projectRoot
	defer cancel3()
	output, err = generateCmd.CombinedOutput()
	if err != nil {
		t.Logf("make generate output:\n%s", string(output))
	}
	require.NoError(t, err, "make generate should succeed")

	gitStatusCmd, cancel4 := cmdWithTimeout(shortTimeout, "git", "status", "--porcelain")
	gitStatusCmd.Dir = projectRoot
	defer cancel4()
	statusOutput, err = gitStatusCmd.CombinedOutput()
	require.NoError(t, err, "git status should succeed")
	statusStr = strings.TrimSpace(string(statusOutput))
	assert.Empty(t, statusStr, "make generate should be idempotent (no changes after generation)")

	t.Log("4. Running tests...")
	testCmd, cancel5 := cmdWithTimeout(e2eTimeout, "make", "test")
	testCmd.Dir = projectRoot
	defer cancel5()
	output, err = testCmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		t.Logf("make test output:\n%s", outputStr)
	}
	require.NoError(t, err, "make test should pass")
	assert.Contains(t, outputStr, "ok", "test output should show passing tests")
	assert.NotContains(t, strings.ToLower(outputStr), "fail", "test output should not contain failures")

	t.Log("5. Running linter...")
	lintCmd, cancel6 := cmdWithTimeout(e2eTimeout, "make", "lint")
	lintCmd.Dir = projectRoot
	defer cancel6()
	output, err = lintCmd.CombinedOutput()
	if err != nil {
		t.Logf("make lint output:\n%s", string(output))
	}
	assert.NoError(t, err, "make lint should succeed with no errors")
	outputStr = strings.ToLower(string(output))
	assert.NotContains(t, outputStr, "error:", "lint output should not contain errors")

	t.Log("6. Building binary...")
	binDir := filepath.Join(projectRoot, "bin")
	err = os.MkdirAll(binDir, 0755)
	require.NoError(t, err, "should create bin directory")

	binaryName := "server"
	if runtime.GOOS == "windows" {
		binaryName = "server.exe"
	}
	binaryPath := filepath.Join(binDir, binaryName)
	buildCmd, cancel7 := cmdWithTimeout(e2eTimeout, "go", "build", "-o", binaryPath, "./cmd/server")
	buildCmd.Dir = projectRoot
	defer cancel7()
	output, err = buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("go build output:\n%s", string(output))
	}
	assert.NoError(t, err, "go build should succeed")

	stat, err := os.Stat(binaryPath)
	assert.NoError(t, err, "binary should exist")
	if err == nil {
		assert.True(t, stat.Size() > 0, "binary should not be empty")
		if runtime.GOOS != "windows" {
			assert.True(t, stat.Mode().Perm()&0100 != 0, "binary should be executable")
		}
	}

	var dbURL string
	switch driver {
	case "sqlite3":
		dataDir := filepath.Join(projectRoot, "data")
		err = os.MkdirAll(dataDir, 0755)
		require.NoError(t, err, "should create data directory")
		dbURL = fmt.Sprintf("file:%s/test.db", dataDir)
		t.Logf("6.5. Using sqlite3 database: %s", dbURL)

	case "go-libsql":
		t.Log("6.5. Starting docker compose services for go-libsql...")
		composeUpCmd, cancel8 := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel8()
		output, err = composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker compose up should succeed")

		defer func() {
			t.Log("Stopping docker compose services...")
			composeDownCmd, cancel := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}()

		t.Log("Waiting for libsql to be ready...")
		time.Sleep(5 * time.Second)
		dbURL = "http://localhost:8080"

	case "postgres":
		t.Log("6.5. Starting docker compose services for postgres...")
		composeUpCmd, cancel9 := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel9()
		output, err = composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker compose up should succeed")

		defer func() {
			t.Log("Stopping docker compose services...")
			composeDownCmd, cancel := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}()

		t.Log("Waiting for postgres to be ready...")
		time.Sleep(5 * time.Second)
		dbURL = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", projectName, projectName, projectName)
	}

	t.Log("7. Starting server...")
	serverCmd := exec.Command(binaryPath)
	serverCmd.Dir = projectRoot
	serverCmd.Env = append(os.Environ(),
		"APP_SERVER_PORT=:18081",
		fmt.Sprintf("APP_DATABASE_URL=%s", dbURL),
	)

	serverLogFile := filepath.Join(tmpDir, "server.log")
	logFile, err := os.Create(serverLogFile)
	require.NoError(t, err, "should create log file")
	defer logFile.Close()
	serverCmd.Stdout = logFile
	serverCmd.Stderr = logFile

	err = serverCmd.Start()
	require.NoError(t, err, "server should start")

	defer func() {
		if serverCmd.Process != nil {
			_ = serverCmd.Process.Kill()
			_ = serverCmd.Wait()
		}
		if t.Failed() {
			if logs, err := os.ReadFile(serverLogFile); err == nil {
				t.Logf("Server logs:\n%s", string(logs))
			}
		}
	}()

	t.Log("Waiting for server to start...")
	time.Sleep(3 * time.Second)

	if serverCmd.ProcessState != nil && serverCmd.ProcessState.Exited() {
		if logs, err := os.ReadFile(serverLogFile); err == nil {
			t.Logf("Server logs:\n%s", string(logs))
		}
		t.Fatal("server exited immediately after starting")
	}
	assert.NotNil(t, serverCmd.Process, "server process should still be running")

	t.Log("8. Checking health endpoint...")
	var resp *http.Response
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if serverCmd.ProcessState != nil && serverCmd.ProcessState.Exited() {
			t.Log("Server process has exited during health checks")
			break
		}

		resp, err = http.Get("http://localhost:18081/api/health")
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			t.Logf("Health check attempt %d/%d failed: %v (retrying...)", i+1, maxRetries, err)
			time.Sleep(1 * time.Second)
		} else {
			t.Logf("Health check attempt %d/%d failed: %v", i+1, maxRetries, err)
		}
	}

	assert.NoError(t, err, "health endpoint should be accessible after retries")
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "health endpoint should return 200 OK")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "should return JSON")
	}

	t.Log("E2E test completed successfully!")
}

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
