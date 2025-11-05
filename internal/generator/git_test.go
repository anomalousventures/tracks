package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeGit_SkipGit(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitializeGit(tmpDir, true)
	require.NoError(t, err)

	gitDir := filepath.Join(tmpDir, ".git")
	_, err = os.Stat(gitDir)
	assert.True(t, os.IsNotExist(err))
}

func TestInitializeGit_Success(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(tmpDir, false)
	require.NoError(t, err)

	gitDir := filepath.Join(tmpDir, ".git")
	_, err = os.Stat(gitDir)
	require.NoError(t, err)

	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Initial commit from Tracks")
}

func TestInitializeGit_GitNotFound(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	os.Setenv("PATH", "")

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(tmpDir, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize git repository")
}

func TestInitializeGit_InitFails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	invalidPath := "/this/path/definitely/does/not/exist/and/cannot/be/created"

	err := InitializeGit(invalidPath, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize git repository")
}

func TestInitializeGit_AddFails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	err := runGitCommand(tmpDir, "init")
	require.NoError(t, err)

	unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
	err = os.WriteFile(unreadableFile, []byte("content"), 0644)
	require.NoError(t, err)

	err = os.Chmod(unreadableFile, 0000)
	require.NoError(t, err)
	defer func() {
		_ = os.Chmod(unreadableFile, 0644)
	}()

	err = runGitCommand(tmpDir, "add", ".")
	if err != nil {
		assert.Error(t, err)
	}
}

func TestInitializeGit_CommitFails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	err := runGitCommand(tmpDir, "init")
	require.NoError(t, err)

	err = runGitCommand(tmpDir, "commit", "-m", "test")
	assert.Error(t, err)
}

func TestInitializeGit_EmptyDirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	err := InitializeGit(tmpDir, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create initial commit")
}

func TestInitializeGit_ConfiguresUser(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(tmpDir, false)
	require.NoError(t, err)

	cmd := exec.Command("git", "config", "user.name")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "Tracks\n", string(output))

	cmd = exec.Command("git", "config", "user.email")
	cmd.Dir = tmpDir
	output, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "tracks@tracks.local\n", string(output))
}
