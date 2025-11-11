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

