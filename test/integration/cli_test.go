// Integration tests for the tracks CLI binary.
//
//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestCLIVersionFlag(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "--version")

	// Cobra's built-in --version flag shows shorter format
	AssertContains(t, stdout, "tracks version")
}

func TestCLIVersionCommand(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "version")

	AssertContains(t, stdout, "Tracks")
	AssertContains(t, stdout, "Commit:")
	AssertContains(t, stdout, "Built:")
}

func TestCLIHelp(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "--help")

	AssertContains(t, stdout, "Usage:")
	AssertContains(t, stdout, "tracks")
	AssertContains(t, stdout, "Available Commands:")
	AssertContains(t, stdout, "Flags:")
}

func TestCLIHelpShortFlag(t *testing.T) {
	stdout, _ := RunCLIExpectSuccess(t, "-h")

	AssertContains(t, stdout, "Usage:")
	AssertContains(t, stdout, "tracks")
}

func TestCLINoArgs(t *testing.T) {
	stdout, stderr, exitCode, _ := RunCLI(t)

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d\nStdout: %s\nStderr: %s",
			exitCode, stdout, stderr)
	}

	output := stdout + stderr
	AssertContains(t, output, "Interactive TUI mode coming in Phase 4")
}

func TestCLIInvalidCommand(t *testing.T) {
	stdout, stderr, _ := RunCLIExpectFailure(t, "nonexistent")

	output := stdout + stderr
	AssertContains(t, output, "unknown command")
}

func TestCLINewCommandMissingArg(t *testing.T) {
	stdout, stderr, _ := RunCLIExpectFailure(t, "new")

	output := stdout + stderr
	AssertContains(t, output, "project-name")
}

func TestCLINewCommandWithArg(t *testing.T) {
	stdout, stderr := RunCLIExpectSuccess(t, "new", "test-project")

	output := stdout + stderr
	AssertContains(t, output, "test-project")
}

func TestCLIGlobalFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"verbose flag", []string{"--verbose", "version"}},
		{"verbose short flag", []string{"-v", "version"}},
		{"quiet flag", []string{"--quiet", "version"}},
		{"quiet short flag", []string{"-q", "version"}},
		{"no-color flag", []string{"--no-color", "version"}},
		{"interactive flag", []string{"--interactive", "version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode, _ := RunCLI(t, tt.args...)

			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d\nStdout: %s\nStderr: %s",
					exitCode, stdout, stderr)
			}
		})
	}
}

func TestCLIVerboseAndQuietMutuallyExclusive(t *testing.T) {
	stdout, stderr, exitCode := RunCLIExpectFailure(t, "--verbose", "--quiet", "version")

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code when both --verbose and --quiet are set")
	}

	output := stdout + stderr
	if !strings.Contains(output, "verbose") || !strings.Contains(output, "quiet") {
		t.Errorf("Expected error message about verbose and quiet being mutually exclusive\nGot: %s", output)
	}
}

func TestCLIColorOutput(t *testing.T) {
	t.Run("default has color codes", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "version")

		if !strings.Contains(stdout, "\033[") {
			t.Skip("Terminal doesn't support color or colors disabled in CI")
		}
	})

	t.Run("no-color removes color codes", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "--no-color", "version")

		if strings.Contains(stdout, "\033[") {
			t.Errorf("Expected no ANSI color codes with --no-color flag\nGot: %s", stdout)
		}
	})
}

func TestCLIJSONOutput(t *testing.T) {
	t.Run("json flag before version command", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "--json", "version")

		AssertJSONOutput(t, stdout)
		AssertContains(t, stdout, "Tracks")
		AssertContains(t, stdout, "Commit:")
		AssertContains(t, stdout, "Built:")
	})

	t.Run("json flag after version command", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "version", "--json")

		AssertJSONOutput(t, stdout)
		AssertContains(t, stdout, "Tracks")
		AssertContains(t, stdout, "Commit:")
		AssertContains(t, stdout, "Built:")
	})

	t.Run("json flag with root command", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "--json")

		AssertJSONOutput(t, stdout)
		AssertContains(t, stdout, "Interactive TUI mode")
	})

	t.Run("json flag with new command", func(t *testing.T) {
		stdout, _ := RunCLIExpectSuccess(t, "--json", "new", "testapp")

		AssertJSONOutput(t, stdout)
		AssertContains(t, stdout, "Creating new Tracks application")
		AssertContains(t, stdout, "testapp")
	})
}

func TestCLIBinaryExists(t *testing.T) {
	binaryPath := GetBinaryPath()
	if binaryPath == "" {
		t.Fatal("Binary path is empty")
	}

	t.Logf("Binary path: %s", binaryPath)
}
