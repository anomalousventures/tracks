package template

import (
	"embed"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRendererInterface verifies that TemplateRenderer implements the Renderer interface
func TestRendererInterface(t *testing.T) {
	var _ Renderer = (*TemplateRenderer)(nil)
}

// TestNewRenderer tests the NewRenderer constructor function
func TestNewRenderer(t *testing.T) {
	var testFS embed.FS
	renderer := NewRenderer(testFS)

	if renderer == nil {
		t.Fatal("NewRenderer returned nil")
	}

	tr, ok := renderer.(*TemplateRenderer)
	if !ok {
		t.Fatal("NewRenderer did not return a *TemplateRenderer")
	}

	if tr.fs != testFS {
		t.Error("TemplateRenderer.fs not set correctly")
	}
}

// TestRendererInterfaceMethods verifies the Renderer interface has the expected methods
func TestRendererInterfaceMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"Render method exists", "Render"},
		{"RenderToFile method exists", "RenderToFile"},
		{"Validate method exists", "Validate"},
	}

	var testFS embed.FS
	renderer := NewRenderer(testFS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.method {
			case "Render":
				_, err := renderer.Render("test.tmpl", TemplateData{})
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			case "RenderToFile":
				err := renderer.RenderToFile("test.tmpl", TemplateData{}, "/tmp/test")
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			case "Validate":
				err := renderer.Validate("test.tmpl")
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			}
		})
	}
}

// TestRenderWithRealTemplates tests rendering with actual embedded templates
func TestRenderWithRealTemplates(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		templateName string
		data         TemplateData
		wantContains []string
	}{
		{
			name:         "test.tmpl with module name",
			templateName: "test.tmpl",
			data: TemplateData{
				ModuleName:  "github.com/user/myapp",
				ProjectName: "myapp",
			},
			wantContains: []string{"Module: github.com/user/myapp", "Project: myapp"},
		},
		{
			name:         "nested template",
			templateName: "cmd/server/test.tmpl",
			data: TemplateData{
				ModuleName: "github.com/example/project",
			},
			wantContains: []string{"Module: github.com/example/project"},
		},
		{
			name:         "all template data fields",
			templateName: "test.tmpl",
			data: TemplateData{
				ModuleName:  "github.com/org/repo",
				ProjectName: "repo",
				DBDriver:    "postgres",
				GoVersion:   "1.25",
				Year:        2025,
			},
			wantContains: []string{"Module: github.com/org/repo", "Project: repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render(tt.templateName, tt.data)
			require.NoError(t, err, "rendering should not fail")
			assert.NotEmpty(t, result, "result should not be empty")

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want, "rendered content should contain expected text")
			}
		})
	}
}

// TestRenderErrors tests error handling in Render method
func TestRenderErrors(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		templateName string
		data         TemplateData
		wantErrType  string
	}{
		{
			name:         "non-existent template",
			templateName: "nonexistent.tmpl",
			data:         TemplateData{},
			wantErrType:  "*template.TemplateError",
		},
		{
			name:         "invalid path",
			templateName: "../../../etc/passwd",
			data:         TemplateData{},
			wantErrType:  "*template.TemplateError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := renderer.Render(tt.templateName, tt.data)
			require.Error(t, err, "should return error")
			assert.IsType(t, &TemplateError{}, err, "error should be TemplateError")
		})
	}
}

// TestRenderWithEmptyData tests rendering with empty TemplateData
func TestRenderWithEmptyData(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	result, err := renderer.Render("test.tmpl", TemplateData{})
	require.NoError(t, err, "rendering with empty data should not fail")
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Module: ")
	assert.Contains(t, result, "Project: ")
}

// TestRenderWithSpecialCharacters tests rendering with special characters in data
func TestRenderWithSpecialCharacters(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/user/my-app_v2",
		ProjectName: "my-app_v2",
		DBDriver:    "go-libsql",
		GoVersion:   "1.25",
	}

	result, err := renderer.Render("test.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "Module: github.com/user/my-app_v2")
	assert.Contains(t, result, "Project: my-app_v2")
}

// TestRenderToFileBasic tests basic file writing functionality
func TestRenderToFileBasic(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	outputPath := filepath.Join(tmpDir, "output.txt")
	data := TemplateData{
		ModuleName:  "github.com/test/app",
		ProjectName: "app",
	}

	err := renderer.RenderToFile("test.tmpl", data, outputPath)
	require.NoError(t, err, "RenderToFile should succeed")

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err, "should be able to read written file")
	assert.Contains(t, string(content), "Module: github.com/test/app")
}

// TestRenderToFileCreatesDirectories tests that parent directories are created
func TestRenderToFileCreatesDirectories(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	outputPath := filepath.Join(tmpDir, "cmd", "server", "main.go")
	data := TemplateData{
		ModuleName: "github.com/test/nested",
	}

	err := renderer.RenderToFile("cmd/server/test.tmpl", data, outputPath)
	require.NoError(t, err, "RenderToFile should create parent directories")

	_, err = os.Stat(outputPath)
	require.NoError(t, err, "file should exist")

	_, err = os.Stat(filepath.Dir(outputPath))
	require.NoError(t, err, "parent directory should exist")
}

// TestRenderToFileCrossPlatform tests cross-platform path handling
func TestRenderToFileCrossPlatform(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		templateName string
		relativePath string
	}{
		{
			name:         "nested unix-style path",
			templateName: "cmd/server/test.tmpl",
			relativePath: filepath.Join("cmd", "server", "main.go"),
		},
		{
			name:         "root level file",
			templateName: "test.tmpl",
			relativePath: "output.txt",
		},
		{
			name:         "deeply nested",
			templateName: "test.tmpl",
			relativePath: filepath.Join("a", "b", "c", "file.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, tt.relativePath)
			data := TemplateData{ModuleName: "test"}

			err := renderer.RenderToFile(tt.templateName, data, outputPath)
			require.NoError(t, err)

			_, err = os.Stat(outputPath)
			require.NoError(t, err, "file should exist at expected path")
		})
	}
}

// TestRenderToFileOverwrites tests that existing files are overwritten
func TestRenderToFileOverwrites(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	outputPath := filepath.Join(tmpDir, "file.txt")

	err := os.WriteFile(outputPath, []byte("old content"), 0644)
	require.NoError(t, err)

	data := TemplateData{ModuleName: "new-module"}
	err = renderer.RenderToFile("test.tmpl", data, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "new-module")
	assert.NotContains(t, string(content), "old content")
}

// TestRenderToFilePermissions tests file permissions
func TestRenderToFilePermissions(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	outputPath := filepath.Join(tmpDir, "file.txt")
	data := TemplateData{ModuleName: "test"}

	err := renderer.RenderToFile("test.tmpl", data, outputPath)
	require.NoError(t, err)

	info, err := os.Stat(outputPath)
	require.NoError(t, err)

	mode := info.Mode()
	assert.True(t, mode.IsRegular(), "should be a regular file")

	// Verify cross-platform permission properties
	perm := mode.Perm()
	assert.NotEqual(t, os.FileMode(0), perm&0400, "file should be readable by owner")
	assert.NotEqual(t, os.FileMode(0), perm&0200, "file should be writable by owner")
	assert.Equal(t, os.FileMode(0), perm&0111, "file should not be executable")

	// On Unix, verify exact permissions
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0644), perm, "file permissions should be 0644")
	}
}

// TestRenderToFileError tests error handling in RenderToFile
func TestRenderToFileError(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		templateName string
		outputPath   string
	}{
		{
			name:         "non-existent template",
			templateName: "nonexistent.tmpl",
			outputPath:   filepath.Join(t.TempDir(), "output.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderer.RenderToFile(tt.templateName, TemplateData{}, tt.outputPath)
			require.Error(t, err)
			assert.IsType(t, &TemplateError{}, err)
		})
	}
}

// TestRenderToFileInvalidPath tests handling of invalid output paths
func TestRenderToFileInvalidPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	invalidPath := filepath.Join(string([]byte{0}), "invalid")
	err := renderer.RenderToFile("test.tmpl", TemplateData{}, invalidPath)
	require.Error(t, err)
}

// TestRenderConsistency tests that Render and RenderToFile produce consistent results
func TestRenderConsistency(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	tmpDir := t.TempDir()

	data := TemplateData{
		ModuleName:  "github.com/test/consistency",
		ProjectName: "consistency",
	}

	renderResult, err := renderer.Render("test.tmpl", data)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.txt")
	err = renderer.RenderToFile("test.tmpl", data, outputPath)
	require.NoError(t, err)

	fileContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	assert.Equal(t, renderResult, string(fileContent), "Render and RenderToFile should produce identical output")
}

// TestRenderMultipleVariables tests templates with multiple variable substitutions
func TestRenderMultipleVariables(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/multi/var",
		ProjectName: "var",
		DBDriver:    "postgres",
		GoVersion:   "1.25",
		Year:        2025,
	}

	result, err := renderer.Render("test.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "github.com/multi/var")
	assert.Contains(t, result, "var")
}

// TestRenderLineEndings tests that line endings are preserved
func TestRenderLineEndings(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	result, err := renderer.Render("test.tmpl", TemplateData{})
	require.NoError(t, err)

	assert.True(t, strings.Contains(result, "\n"), "should contain newlines")
}
