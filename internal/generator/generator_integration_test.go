package generator

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
func cmdWithTimeout(timeout time.Duration, name string, args ...string) *exec.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cmd := exec.CommandContext(ctx, name, args...)

	// Store cancel in a way that won't cause issues, but note that
	// the context will be canceled after timeout regardless
	_ = cancel

	return cmd
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

// TestGoModDownload (#144) - Verifies go mod download succeeds on generated projects
func TestGoModDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			cmd := cmdWithTimeout(longTimeout, "go", "mod", "download")
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Logf("go mod download output:\n%s", string(output))
			}

			assert.NoError(t, err, "go mod download should succeed")

			goSumPath := filepath.Join(projectRoot, "go.sum")
			stat, err := os.Stat(goSumPath)
			assert.NoError(t, err, "go.sum should be created")
			if err == nil {
				assert.True(t, stat.Size() > 0, "go.sum should not be empty")
			}
		})
	}
}

// TestGoTestPasses (#145) - Verifies all generated tests pass
func TestGoTestPasses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			// Run go mod tidy to download dependencies and populate go.sum
			tidyCmd := cmdWithTimeout(mediumTimeout, "go", "mod", "tidy")
			tidyCmd.Dir = projectRoot
			output, err := tidyCmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy output:\n%s", string(output))
				t.Fatalf("go mod tidy failed: %v", err)
			}

			testCmd := cmdWithTimeout(mediumTimeout, "go", "test", "./...")
			testCmd.Dir = projectRoot
			output, err = testCmd.CombinedOutput()

			if err != nil {
				t.Logf("go test output:\n%s", string(output))
			}

			assert.NoError(t, err, "go test should pass")
			assert.Contains(t, string(output), "ok", "should have passing test packages")
			assert.NotContains(t, string(output), "FAIL", "test output should not contain failures")
		})
	}
}

// TestGoBuildSucceeds (#146) - Verifies the server binary builds successfully
func TestGoBuildSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			// Run go mod tidy to download dependencies and populate go.sum
			tidyCmd := cmdWithTimeout(mediumTimeout, "go", "mod", "tidy")
			tidyCmd.Dir = projectRoot
			output, err := tidyCmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy output:\n%s", string(output))
				t.Fatalf("go mod tidy failed: %v", err)
			}

			binDir := filepath.Join(projectRoot, "bin")
			err = os.MkdirAll(binDir, 0755)
			require.NoError(t, err)

			binaryPath := filepath.Join(binDir, "server")
			buildCmd := cmdWithTimeout(mediumTimeout, "go", "build", "-o", binaryPath, "./cmd/server")
			buildCmd.Dir = projectRoot
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
		})
	}
}

// TestServerRuns (#147) - Verifies the server binary starts and runs
func TestServerRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			// Run go mod tidy to download dependencies and populate go.sum
			tidyCmd := cmdWithTimeout(mediumTimeout, "go", "mod", "tidy")
			tidyCmd.Dir = projectRoot
			output, err := tidyCmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy output:\n%s", string(output))
				t.Fatalf("go mod tidy failed: %v", err)
			}

			binDir := filepath.Join(projectRoot, "bin")
			err = os.MkdirAll(binDir, 0755)
			require.NoError(t, err)

			binaryPath := filepath.Join(binDir, "server")
			buildCmd := cmdWithTimeout(mediumTimeout, "go", "build", "-o", binaryPath, "./cmd/server")
			buildCmd.Dir = projectRoot
			err = buildCmd.Run()
			require.NoError(t, err)

			cmd := cmdWithTimeout(shortTimeout, binaryPath)
			cmd.Dir = projectRoot
			cmd.Env = append(os.Environ(), "APP_SERVER_PORT=:18081")

			err = cmd.Start()
			require.NoError(t, err, "server should start")

			defer func() {
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
				}
			}()

			// Wait for server to initialize and check it's still running
			time.Sleep(2 * time.Second)

			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				t.Fatal("server exited immediately after starting")
			}

			assert.NotNil(t, cmd.Process, "server process should still be running")
		})
	}
}

// TestHealthCheckEndpoint (#148) - Verifies health check endpoint returns 200 OK
func TestHealthCheckEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			// Run go mod tidy to download dependencies and populate go.sum
			tidyCmd := cmdWithTimeout(mediumTimeout, "go", "mod", "tidy")
			tidyCmd.Dir = projectRoot
			output, err := tidyCmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy output:\n%s", string(output))
				t.Fatalf("go mod tidy failed: %v", err)
			}

			binDir := filepath.Join(projectRoot, "bin")
			err = os.MkdirAll(binDir, 0755)
			require.NoError(t, err)

			binaryPath := filepath.Join(binDir, "server")
			buildCmd := cmdWithTimeout(mediumTimeout, "go", "build", "-o", binaryPath, "./cmd/server")
			buildCmd.Dir = projectRoot
			err = buildCmd.Run()
			require.NoError(t, err)

			port := "18080"

			// Set database URL based on driver
			var dbURL string
			switch driver {
			case "go-libsql":
				dbURL = "libsql://:memory:"
			case "sqlite3":
				dbURL = ":memory:"
			case "postgres":
				t.Skip("postgres requires running database server")
				return
			}

			cmd := cmdWithTimeout(shortTimeout, binaryPath)
			cmd.Dir = projectRoot
			envVars := []string{
				fmt.Sprintf("APP_SERVER_PORT=:%s", port),
				fmt.Sprintf("APP_DATABASE_URL=%s", dbURL),
			}
			t.Logf("Setting env vars: %v", envVars)
			cmd.Env = append(os.Environ(), envVars...)

			// Capture server output for debugging
			var stdout, stderr strings.Builder
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Start()
			require.NoError(t, err)

			defer func() {
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
				}
			}()

			// Poll for server readiness with timeout
			healthURL := fmt.Sprintf("http://localhost:%s/api/health", port)
			var resp *http.Response
			maxRetries := 30
			for i := 0; i < maxRetries; i++ {
				// Check if process is still running
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					t.Fatalf("server process exited unexpectedly")
				}

				resp, err = http.Get(healthURL)
				if err == nil {
					break
				}

				if i == maxRetries-1 {
					t.Logf("Failed to GET %s after %d retries: %v", healthURL, maxRetries, err)
					t.Logf("Server stdout:\n%s", stdout.String())
					t.Logf("Server stderr:\n%s", stderr.String())
					require.NoError(t, err, "should be able to GET health endpoint after retries")
				}

				time.Sleep(200 * time.Millisecond)
			}

			require.NoError(t, err, "should be able to GET health endpoint")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "health check should return 200 OK")
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "should return JSON")
		})
	}
}

// TestMakeGenerateIdempotent (#225) - Verifies make generate is idempotent after project generation
func TestMakeGenerateIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
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

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			idempotencyCheckCmd := cmdWithTimeout(mediumTimeout, "make", "generate")
			idempotencyCheckCmd.Dir = projectRoot
			output, err := idempotencyCheckCmd.CombinedOutput()
			if err != nil {
				t.Logf("idempotency check output:\n%s", string(output))
			}
			require.NoError(t, err, "idempotency check should succeed")

			gitStatusCmd := cmdWithTimeout(shortTimeout, "git", "status", "--porcelain")
			gitStatusCmd.Dir = projectRoot
			statusOutput, err := gitStatusCmd.CombinedOutput()
			require.NoError(t, err)

			statusStr := strings.TrimSpace(string(statusOutput))
			assert.Empty(t, statusStr, "make generate should be idempotent (no changes on second run)")
		})
	}
}

// TestMakeLintSucceeds (#226) - Verifies make lint succeeds after project generation
func TestMakeLintSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			lintCmd := cmdWithTimeout(mediumTimeout, "make", "lint")
			lintCmd.Dir = projectRoot
			output, err := lintCmd.CombinedOutput()

			if err != nil {
				t.Logf("make lint output:\n%s", string(output))
			}

			assert.NoError(t, err, "make lint should succeed with no errors")

			outputStr := strings.ToLower(string(output))
			assert.NotContains(t, outputStr, "error:", "lint output should not contain errors")
		})
	}
}

// TestGeneratedProjectTestsShouldPass verifies that generated projects include
// working tests that pass. Project generation now includes running make generate
// before writing test files, so tests can import generated mocks.
func TestGeneratedProjectTestsShouldPass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	drivers := []string{"go-libsql", "sqlite3"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
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

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, projectName)

			testCmd := cmdWithTimeout(mediumTimeout, "make", "test")
			testCmd.Dir = projectRoot
			testOutput, err := testCmd.CombinedOutput()

			if err != nil {
				t.Logf("make test output:\n%s", string(testOutput))
			}

			require.NoError(t, err, "make test should succeed")

			outputStr := string(testOutput)
			assert.Contains(t, outputStr, "ok", "test output should show passing tests")
			assert.NotContains(t, strings.ToLower(outputStr), "fail", "test output should not contain failures")
		})
	}
}
