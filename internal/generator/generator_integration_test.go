package generator

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTimeout returns a timeout duration from environment or default.
// Env vars: INTEGRATION_TEST_SHORT_TIMEOUT, INTEGRATION_TEST_MEDIUM_TIMEOUT, INTEGRATION_TEST_LONG_TIMEOUT
func getTimeout(envVar string, defaultTimeout time.Duration) time.Duration {
	if val := os.Getenv(envVar); val != "" {
		if timeout, err := time.ParseDuration(val); err == nil && timeout > 0 {
			return timeout
		}
	}
	return defaultTimeout
}

// cmdWithTimeout creates an exec.Command with a timeout context
// to prevent integration tests from hanging indefinitely.
func cmdWithTimeout(timeout time.Duration, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, cancel
}

var (
	shortTimeout  = getTimeout("INTEGRATION_TEST_SHORT_TIMEOUT", 2*time.Second)   // git, local ops
	mediumTimeout = getTimeout("INTEGRATION_TEST_MEDIUM_TIMEOUT", 10*time.Second) // compile, test, lint
	longTimeout   = getTimeout("INTEGRATION_TEST_LONG_TIMEOUT", 15*time.Second)   // network, downloads
)

// TestGenerateFullProject (#143) - Foundational integration test that generates
// a complete project structure and verifies all files and directories are created correctly.
func TestGenerateFullProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		databaseDriver string
		initGit        bool
		modulePath     string
	}{
		{
			name:           "go-libsql without git",
			databaseDriver: "go-libsql",
			initGit:        false,
			modulePath:     "github.com/test/libsql-app",
		},
		{
			name:           "sqlite3 with git",
			databaseDriver: "sqlite3",
			initGit:        true,
			modulePath:     "github.com/test/sqlite-app",
		},
		{
			name:           "postgres without git",
			databaseDriver: "postgres",
			initGit:        false,
			modulePath:     "example.com/postgres-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     tt.modulePath,
				DatabaseDriver: tt.databaseDriver,
				EnvPrefix:      "APP",
				InitGit:        tt.initGit,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Validate(cfg)
			require.NoError(t, err, "validation should succeed")

			err = gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)

			expectedFiles := []string{
				"go.mod",
				"README.md",
				".gitignore",
				".golangci.yml",
				".mockery.yaml",
				".tracks.yaml",
				".env.example",
				"Makefile",
				"sqlc.yaml",
				"cmd/server/main.go",
				"internal/config/config.go",
				"internal/interfaces/health.go",
				"internal/interfaces/logger.go",
				"internal/logging/logger.go",
				"internal/domain/health/service.go",
				"internal/http/server.go",
				"internal/http/routes.go",
				"internal/http/routes/routes.go",
				"internal/http/handlers/health.go",
				"internal/http/middleware/logging.go",
				"internal/db/db.go",
			}

			for _, file := range expectedFiles {
				path := filepath.Join(projectRoot, file)
				_, err := os.Stat(path)
				assert.NoError(t, err, "file should exist: %s", file)
			}

			expectedDirs := []string{
				"cmd",
				"cmd/server",
				"internal",
				"internal/config",
				"internal/interfaces",
				"internal/logging",
				"internal/domain",
				"internal/domain/health",
				"internal/http",
				"internal/http/routes",
				"internal/http/handlers",
				"internal/http/middleware",
				"internal/db",
			}

			for _, dir := range expectedDirs {
				path := filepath.Join(projectRoot, dir)
				stat, err := os.Stat(path)
				assert.NoError(t, err, "directory should exist: %s", dir)
				if err == nil {
					assert.True(t, stat.IsDir(), "%s should be a directory", dir)
				}
			}

			goModPath := filepath.Join(projectRoot, "go.mod")
			content, err := os.ReadFile(goModPath)
			require.NoError(t, err, "should be able to read go.mod")

			assert.Contains(t, string(content), tt.modulePath, "go.mod should contain module path")
			assert.Contains(t, string(content), "go 1.25", "go.mod should contain Go version")

			dbPath := filepath.Join(projectRoot, "internal/db/db.go")
			dbContent, err := os.ReadFile(dbPath)
			require.NoError(t, err, "should be able to read internal/db/db.go")

			driverMapping := map[string]string{
				"go-libsql": "libsql",
				"sqlite3":   "sqlite3",
				"postgres":  "postgres",
			}
			expectedDriver := driverMapping[tt.databaseDriver]
			assert.Contains(t, string(dbContent), expectedDriver, "db.go should contain correct driver")

			if tt.initGit {
				gitDir := filepath.Join(projectRoot, ".git")
				stat, err := os.Stat(gitDir)
				assert.NoError(t, err, ".git directory should exist when git is initialized")
				if err == nil {
					assert.True(t, stat.IsDir(), ".git should be a directory")
				}
			} else {
				gitDir := filepath.Join(projectRoot, ".git")
				_, err := os.Stat(gitDir)
				assert.True(t, os.IsNotExist(err), ".git directory should not exist when git is not initialized")
			}
		})
	}
}

// TestGenerateFullProject_CustomModuleName tests generation with custom module name
func TestGenerateFullProject_CustomModuleName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "customapp"
	modulePath := "mycorp.com/apps/customapp"

	cfg := ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     modulePath,
		DatabaseDriver: "postgres",
		EnvPrefix:      "CUSTOM",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, projectName)
	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), modulePath)
}

// TestGenerateFullProject_DirectoryAlreadyExists tests that generation fails
// when the target directory already exists
func TestGenerateFullProject_DirectoryAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "existing"

	existingDir := filepath.Join(tmpDir, projectName)
	err := os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	cfg := ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/existing",
		DatabaseDriver: "postgres",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()

	err = gen.Validate(cfg)
	assert.Error(t, err, "validation should fail when directory exists")
	assert.Contains(t, err.Error(), "already exists")
}

// runE2ETest validates the user-facing contract: generated projects must work correctly
// across all supported platforms and database drivers without manual intervention.
func runE2ETest(t *testing.T, driver string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectName := "testapp"

	cfg := ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
		DatabaseDriver: driver,
		EnvPrefix:      "APP",
		InitGit:        true,
		OutputPath:     tmpDir,
	}

	t.Log("1. Generating project...")
	gen := NewProjectGenerator()
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
	generateCmd, cancel3 := cmdWithTimeout(mediumTimeout, "make", "generate")
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
	testCmd, cancel5 := cmdWithTimeout(mediumTimeout, "make", "test")
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
	lintCmd, cancel6 := cmdWithTimeout(mediumTimeout, "make", "lint")
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
	buildCmd, cancel7 := cmdWithTimeout(mediumTimeout, "go", "build", "-o", binaryPath, "./cmd/server")
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
		assert.True(t, stat.Mode().Perm()&0100 != 0, "binary should be executable")
	}

	// 6.5. Configure database for E2E test
	var dbURL string
	switch driver {
	case "sqlite3":
		dataDir := filepath.Join(projectRoot, "data")
		err = os.MkdirAll(dataDir, 0755)
		require.NoError(t, err, "should create data directory")
		dbURL = fmt.Sprintf("file:%s/test.db", dataDir)
		t.Logf("6.5. Using sqlite3 database: %s", dbURL)

	case "go-libsql":
		t.Log("6.5. Starting docker-compose services for go-libsql...")
		composeUpCmd, cancel8 := cmdWithTimeout(longTimeout, "docker-compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel8()
		output, err = composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker-compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker-compose up should succeed")

		defer func() {
			t.Log("Stopping docker-compose services...")
			composeDownCmd, cancel := cmdWithTimeout(mediumTimeout, "docker-compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}()

		// Wait for libsql to be healthy
		t.Log("Waiting for libsql to be ready...")
		time.Sleep(5 * time.Second)
		dbURL = "http://localhost:8080"

	case "postgres":
		t.Log("6.5. Starting docker-compose services for postgres...")
		composeUpCmd, cancel9 := cmdWithTimeout(longTimeout, "docker-compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel9()
		output, err = composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker-compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker-compose up should succeed")

		defer func() {
			t.Log("Stopping docker-compose services...")
			composeDownCmd, cancel := cmdWithTimeout(mediumTimeout, "docker-compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}()

		// Wait for postgres to be healthy
		t.Log("Waiting for postgres to be ready...")
		time.Sleep(5 * time.Second)
		dbURL = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", projectName, projectName, projectName)
	}

	t.Log("7. Starting server...")
	// Server is a long-running process - don't use a timeout context
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

	// Give server time to initialize database connection and start listening
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

// TestE2E_GoLibsql runs full E2E test suite for go-libsql driver
func TestE2E_GoLibsql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "go-libsql")
}

// TestE2E_SQLite3 runs full E2E test suite for sqlite3 driver
func TestE2E_SQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "sqlite3")
}

// TestE2E_Postgres runs full E2E test suite for postgres driver
func TestE2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "postgres")
}
