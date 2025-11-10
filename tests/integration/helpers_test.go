package integration

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerCmdWithTimeout(t *testing.T) {
	t.Run("constructs command with correct binary", func(t *testing.T) {
		cmd, cancel := dockerCmdWithTimeout(5*time.Second, "compose", "up", "-d")
		defer cancel()

		if runtime.GOOS == "windows" {
			assert.Contains(t, cmd.Path, "wsl", "should use wsl on Windows")
			expectedArgs := []string{"wsl", "--", "docker", "compose", "up", "-d"}
			assert.Equal(t, expectedArgs, cmd.Args, "should prefix docker command with wsl --")
		} else {
			assert.Contains(t, cmd.Path, "docker", "should use docker directly on Unix")
			expectedArgs := []string{"docker", "compose", "up", "-d"}
			assert.Equal(t, expectedArgs, cmd.Args, "should pass docker args directly")
		}
	})

	t.Run("timeout is respected", func(t *testing.T) {
		timeout := 100 * time.Millisecond
		cmd, cancel := dockerCmdWithTimeout(timeout, "sleep", "10")
		defer cancel()

		require.NotNil(t, cmd, "command should be created")

		err := cmd.Run()
		assert.Error(t, err, "command should timeout")
	})

	t.Run("cancel function stops command", func(t *testing.T) {
		cmd, cancel := dockerCmdWithTimeout(5*time.Second, "sleep", "10")

		require.NotNil(t, cmd, "command should be created")

		err := cmd.Start()
		if err != nil {
			t.Skip("cannot start command (docker may not be available)")
		}

		cancel()

		waitErr := cmd.Wait()
		assert.Error(t, waitErr, "command should be killed after cancel()")
	})

	t.Run("handles empty args", func(t *testing.T) {
		cmd, cancel := dockerCmdWithTimeout(5*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			expectedArgs := []string{"wsl", "--", "docker"}
			assert.Equal(t, expectedArgs, cmd.Args, "should handle empty args on Windows")
		} else {
			expectedArgs := []string{"docker"}
			assert.Equal(t, expectedArgs, cmd.Args, "should handle empty args on Unix")
		}
	})

	t.Run("handles multiple args correctly", func(t *testing.T) {
		args := []string{"compose", "-f", "docker-compose.yml", "up", "-d", "--wait"}
		cmd, cancel := dockerCmdWithTimeout(5*time.Second, args...)
		defer cancel()

		if runtime.GOOS == "windows" {
			expectedArgs := append([]string{"wsl", "--", "docker"}, args...)
			assert.Equal(t, expectedArgs, cmd.Args, "should preserve arg order on Windows")
		} else {
			expectedArgs := append([]string{"docker"}, args...)
			assert.Equal(t, expectedArgs, cmd.Args, "should preserve arg order on Unix")
		}
	})
}

func TestCmdWithTimeout(t *testing.T) {
	t.Run("constructs command correctly", func(t *testing.T) {
		cmd, cancel := cmdWithTimeout(5*time.Second, "go", "version")
		defer cancel()

		assert.Contains(t, cmd.Path, "go", "should use specified command")
		assert.Equal(t, []string{"go", "version"}, cmd.Args, "should pass args correctly")
	})

	t.Run("timeout is respected", func(t *testing.T) {
		timeout := 100 * time.Millisecond
		cmd, cancel := cmdWithTimeout(timeout, "sleep", "10")
		defer cancel()

		require.NotNil(t, cmd, "command should be created")

		err := cmd.Run()
		assert.Error(t, err, "command should timeout")
	})
}

func TestGetTimeout(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		defaultTimeout := 10 * time.Second
		timeout := getTimeout("NONEXISTENT_ENV_VAR", defaultTimeout)
		assert.Equal(t, defaultTimeout, timeout, "should return default timeout")
	})

	t.Run("returns default when env var is invalid", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUT", "invalid")
		defaultTimeout := 10 * time.Second
		timeout := getTimeout("TEST_TIMEOUT", defaultTimeout)
		assert.Equal(t, defaultTimeout, timeout, "should return default for invalid duration")
	})

	t.Run("returns default when env var is zero", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUT", "0s")
		defaultTimeout := 10 * time.Second
		timeout := getTimeout("TEST_TIMEOUT", defaultTimeout)
		assert.Equal(t, defaultTimeout, timeout, "should return default for zero duration")
	})

	t.Run("returns default when env var is negative", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUT", "-5s")
		defaultTimeout := 10 * time.Second
		timeout := getTimeout("TEST_TIMEOUT", defaultTimeout)
		assert.Equal(t, defaultTimeout, timeout, "should return default for negative duration")
	})

	t.Run("returns parsed value when env var is valid", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUT", "30s")
		defaultTimeout := 10 * time.Second
		timeout := getTimeout("TEST_TIMEOUT", defaultTimeout)
		assert.Equal(t, 30*time.Second, timeout, "should return parsed timeout from env var")
	})
}

func TestSetupDatabaseForDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database setup test in short mode")
	}

	t.Run("sqlite3 creates data directory and returns file URL", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectRoot := tmpDir
		projectName := "testapp"

		setup := setupDatabaseForDocker(t, "sqlite3", tmpDir, projectRoot, projectName)
		defer setup.cleanupFunc()

		assert.Equal(t, "file:/app/data/test.db", setup.dbURL, "should return correct SQLite URL")
		assert.False(t, setup.composeStarted, "compose should not be started for sqlite3")
		assert.NotNil(t, setup.cleanupFunc, "cleanup function should not be nil")
	})

	t.Run("postgres returns correct connection string", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping postgres docker test on Windows")
		}

		tmpDir := t.TempDir()
		projectRoot := tmpDir
		projectName := "testapp"

		expectedURL := "postgres://testapp:testapp@postgres:5432/testapp?sslmode=disable"

		setup := setupDatabaseForDocker(t, "postgres", tmpDir, projectRoot, projectName)
		if setup.composeStarted {
			defer setup.cleanupFunc()
		}

		assert.Equal(t, expectedURL, setup.dbURL, "should return correct Postgres URL")
		assert.NotNil(t, setup.cleanupFunc, "cleanup function should not be nil")
	})

	t.Run("go-libsql returns correct connection string", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping go-libsql docker test on Windows")
		}

		tmpDir := t.TempDir()
		projectRoot := tmpDir
		projectName := "testapp"

		expectedURL := "http://libsql:8080"

		setup := setupDatabaseForDocker(t, "go-libsql", tmpDir, projectRoot, projectName)
		if setup.composeStarted {
			defer setup.cleanupFunc()
		}

		assert.Equal(t, expectedURL, setup.dbURL, "should return correct LibSQL URL")
		assert.NotNil(t, setup.cleanupFunc, "cleanup function should not be nil")
	})
}

func TestWaitForHealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping health endpoint test in short mode")
	}

	t.Run("fails when endpoint not available", func(t *testing.T) {
		mockT := &testing.T{}

		port := 19999
		maxRetries := 2
		containerName := "nonexistent-container"

		waitForHealthEndpoint(mockT, containerName, port, maxRetries)

		assert.True(t, mockT.Failed(), "test should fail when endpoint is not available")
	})
}

func TestBuildDockerImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker build test in short mode")
	}

	t.Run("returns cleanup function", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping docker build test on Windows")
		}

		mockT := &testing.T{}
		tmpDir := t.TempDir()
		imageName := "test-image-that-will-fail"

		cleanup := buildDockerImage(mockT, tmpDir, imageName)

		assert.NotNil(t, cleanup, "should return cleanup function")
		assert.True(t, mockT.Failed(), "test should fail when Dockerfile doesn't exist")
	})
}

func TestScanDockerImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker scan test in short mode")
	}

	t.Run("requires valid image", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping trivy scan test on Windows")
		}

		mockT := &testing.T{}
		imageName := "nonexistent-image-for-scan"

		scanDockerImage(mockT, imageName)

		assert.True(t, mockT.Failed(), "test should fail when image doesn't exist")
	})
}

func TestStartDockerContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker container test in short mode")
	}

	t.Run("returns cleanup function for sqlite3", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping docker container test on Windows")
		}

		mockT := &testing.T{}
		tmpDir := t.TempDir()
		containerName := "test-container"
		imageName := "nonexistent-image"
		dbURL := "file:/app/data/test.db"
		driver := "sqlite3"
		projectName := "testapp"

		cleanup := startDockerContainer(mockT, containerName, imageName, dbURL, driver, tmpDir, projectName)

		assert.NotNil(t, cleanup, "should return cleanup function")
		assert.True(t, mockT.Failed(), "test should fail when image doesn't exist")

		if cleanup != nil {
			cleanup()
		}
	})

	t.Run("returns cleanup function for postgres", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping docker container test on Windows")
		}

		mockT := &testing.T{}
		tmpDir := t.TempDir()
		containerName := "test-container-pg"
		imageName := "nonexistent-image"
		dbURL := "postgres://user:pass@postgres:5432/db"
		driver := "postgres"
		projectName := "testapp"

		cleanup := startDockerContainer(mockT, containerName, imageName, dbURL, driver, tmpDir, projectName)

		assert.NotNil(t, cleanup, "should return cleanup function")
		assert.True(t, mockT.Failed(), "test should fail when image doesn't exist")

		if cleanup != nil {
			cleanup()
		}
	})
}
