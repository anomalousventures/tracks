package templates

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFSNotNil verifies the embedded FS is accessible
func TestFSNotNil(t *testing.T) {
	assert.NotNil(t, FS, "FS should not be nil")
}

func TestReadProjectDir(t *testing.T) {
	entries, err := fs.ReadDir(FS, "project")
	require.NoError(t, err, "should be able to read project directory")
	assert.NotEmpty(t, entries, "project directory should not be empty")
}

func TestReadNestedDir(t *testing.T) {
	entries, err := fs.ReadDir(FS, "project/cmd")
	require.NoError(t, err, "should be able to read project/cmd directory")
	assert.NotEmpty(t, entries, "cmd directory should not be empty")

	entries, err = fs.ReadDir(FS, "project/cmd/server")
	require.NoError(t, err, "should be able to read project/cmd/server directory")
	assert.NotEmpty(t, entries, "server directory should not be empty")
}

func TestReadTemplateFile(t *testing.T) {
	data, err := fs.ReadFile(FS, "project/test.tmpl")
	require.NoError(t, err, "should be able to read test.tmpl")
	assert.NotEmpty(t, data, "file content should not be empty")
	assert.Contains(t, string(data), "Test template", "file should contain expected content")
}

func TestReadNestedTemplateFile(t *testing.T) {
	data, err := fs.ReadFile(FS, "project/cmd/server/test.tmpl")
	require.NoError(t, err, "should be able to read nested test.tmpl")
	assert.NotEmpty(t, data, "file content should not be empty")
	assert.Contains(t, string(data), "Nested test template", "file should contain expected content")
}

// TestCrossPlatformPaths verifies embed FS uses forward slashes across all platforms.
// This test ensures that paths work consistently on Windows, macOS, and Linux.
func TestCrossPlatformPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"root template", "project/test.tmpl"},
		{"nested cmd", "project/cmd/server/test.tmpl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := fs.ReadFile(FS, tt.path)
			require.NoError(t, err, "embed FS should use forward slashes on all platforms")
			assert.NotEmpty(t, data, "should be able to read file with forward slash path")
		})
	}
}

// TestPathSeparators verifies that embed.FS always uses forward slashes
func TestPathSeparators(t *testing.T) {
	forwardSlashPath := "project/cmd/server/test.tmpl"
	backslashPath := filepath.FromSlash("project/cmd/server/test.tmpl")

	data, err := fs.ReadFile(FS, forwardSlashPath)
	require.NoError(t, err, "forward slash path should work")
	assert.NotEmpty(t, data)

	if forwardSlashPath != backslashPath {
		_, err = fs.ReadFile(FS, backslashPath)
		assert.Error(t, err, "backslash path should fail (embed.FS uses forward slashes only)")
	}
}

func TestWalkEmbeddedFiles(t *testing.T) {
	var templateCount int
	err := fs.WalkDir(FS, "project", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".tmpl" {
			templateCount++
		}
		return nil
	})

	require.NoError(t, err, "should be able to walk embedded FS")
	assert.GreaterOrEqual(t, templateCount, 2, "should find at least 2 test template files")
}

func TestGlobPattern(t *testing.T) {
	matches, err := fs.Glob(FS, "project/*.tmpl")
	require.NoError(t, err, "should be able to use glob patterns")
	assert.NotEmpty(t, matches, "should find template files with glob pattern")

	nestedMatches, err := fs.Glob(FS, "project/cmd/server/*.tmpl")
	require.NoError(t, err, "should be able to use glob patterns on nested paths")
	assert.NotEmpty(t, nestedMatches, "should find nested template files")
}

// TestDirectoryStructure verifies the expected directory structure exists
func TestDirectoryStructure(t *testing.T) {
	expectedDirs := []string{
		"project",
		"project/cmd",
		"project/cmd/server",
	}

	for _, dir := range expectedDirs {
		t.Run(dir, func(t *testing.T) {
			entries, err := fs.ReadDir(FS, dir)
			require.NoError(t, err, "directory %s should exist", dir)
			assert.NotNil(t, entries, "should be able to read directory %s", dir)
		})
	}
}

// TestNonExistentPath verifies proper error handling for missing files
func TestNonExistentPath(t *testing.T) {
	_, err := fs.ReadFile(FS, "project/nonexistent.tmpl")
	assert.Error(t, err, "should return error for non-existent file")
	assert.ErrorIs(t, err, fs.ErrNotExist, "error should be fs.ErrNotExist")
}
