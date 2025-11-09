package generator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeGit_SkipGit(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	err := InitializeGit(ctx, tmpDir, true)
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
	ctx := context.Background()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(ctx, tmpDir, false)
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
	ctx := context.Background()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(ctx, tmpDir, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize git repository")
}

func TestInitializeGit_InitFails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	invalidPath := "/this/path/definitely/does/not/exist/and/cannot/be/created"
	ctx := context.Background()

	err := InitializeGit(ctx, invalidPath, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize git repository")
}

func TestInitializeGit_EmptyDirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()
	ctx := context.Background()

	err := InitializeGit(ctx, tmpDir, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create initial commit")
}

func TestInitializeGit_ConfiguresUser(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()
	ctx := context.Background()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	err = InitializeGit(ctx, tmpDir, false)
	require.NoError(t, err)

	cmd := exec.Command("git", "config", "--local", "user.name")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "Tracks\n", string(output))

	cmd = exec.Command("git", "config", "--local", "user.email")
	cmd.Dir = tmpDir
	output, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "info@anomalous.ventures\n", string(output))
}
