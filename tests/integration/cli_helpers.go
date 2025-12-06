// Package integration provides helpers for CLI integration tests.
//
//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binaryName := "tracks"
	if runtime.GOOS == "windows" {
		binaryName = "tracks.exe"
	}
	binaryPath = filepath.Join(projectRoot, "bin", binaryName)
}

// RunCLI executes the tracks binary with the given arguments.
// Returns stdout, stderr, exit code, and any error.
func RunCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	return RunCLIInDir(t, "", args...)
}

// RunCLIInDir executes the tracks binary from a specific directory.
// If dir is empty, runs in the current working directory.
// Returns stdout, stderr, exit code, and any error.
func RunCLIInDir(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("tracks binary not found at %s. Run 'make build' first.", binaryPath)
	}

	cmd := exec.Command(binaryPath, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to execute CLI: %v", err)
		}
	}

	return stdout, stderr, exitCode, err
}

// RunCLIExpectSuccess runs the CLI and expects a zero exit code.
// Returns stdout and stderr. Fails the test if exit code is non-zero.
func RunCLIExpectSuccess(t *testing.T, args ...string) (stdout, stderr string) {
	t.Helper()
	return RunCLIInDirExpectSuccess(t, "", args...)
}

// RunCLIInDirExpectSuccess runs the CLI from a directory and expects a zero exit code.
func RunCLIInDirExpectSuccess(t *testing.T, dir string, args ...string) (stdout, stderr string) {
	t.Helper()

	stdout, stderr, exitCode, err := RunCLIInDir(t, dir, args...)
	if exitCode != 0 {
		t.Fatalf("CLI failed with exit code %d\nDir: %s\nArgs: %v\nStdout: %s\nStderr: %s\nError: %v",
			exitCode, dir, args, stdout, stderr, err)
	}

	return stdout, stderr
}

// RunCLIExpectFailure runs the CLI and expects a non-zero exit code.
// Returns stdout, stderr, and exit code. Fails the test if exit code is zero.
func RunCLIExpectFailure(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	return RunCLIInDirExpectFailure(t, "", args...)
}

// RunCLIInDirExpectFailure runs the CLI from a directory and expects a non-zero exit code.
func RunCLIInDirExpectFailure(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	stdout, stderr, exitCode, _ = RunCLIInDir(t, dir, args...)
	if exitCode == 0 {
		t.Fatalf("CLI succeeded but expected failure\nDir: %s\nArgs: %v\nStdout: %s\nStderr: %s",
			dir, args, stdout, stderr)
	}

	return stdout, stderr, exitCode
}

// AssertContains fails the test if the string doesn't contain the substring.
func AssertContains(t *testing.T, got, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q\nGot: %s", want, got)
	}
}

// AssertNotContains fails the test if the string contains the substring.
func AssertNotContains(t *testing.T, got, unwanted string) {
	t.Helper()

	if strings.Contains(got, unwanted) {
		t.Errorf("expected output to NOT contain %q\nGot: %s", unwanted, got)
	}
}

// AssertJSONOutput verifies that output is valid JSON.
func AssertJSONOutput(t *testing.T, output string) {
	t.Helper()

	if !json.Valid([]byte(output)) {
		t.Errorf("expected valid JSON output\nGot: %s", output)
	}
}

// GetBinaryPath returns the absolute path to the tracks binary.
func GetBinaryPath() string {
	return binaryPath
}
