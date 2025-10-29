package template

import (
	"embed"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator/interfaces"
	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRendererInterface verifies that templateRenderer implements the interfaces.TemplateRenderer interface
func TestRendererInterface(t *testing.T) {
	var _ interfaces.TemplateRenderer = (*templateRenderer)(nil)
}

func TestNewRenderer(t *testing.T) {
	var testFS embed.FS
	renderer := NewRenderer(testFS)

	if renderer == nil {
		t.Fatal("NewRenderer returned nil")
	}

	tr, ok := renderer.(*templateRenderer)
	if !ok {
		t.Fatal("NewRenderer did not return a *templateRenderer")
	}

	if tr.fs != testFS {
		t.Error("templateRenderer.fs not set correctly")
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

func TestRenderWithEmptyData(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	result, err := renderer.Render("test.tmpl", TemplateData{})
	require.NoError(t, err, "rendering with empty data should not fail")
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Module: ")
	assert.Contains(t, result, "Project: ")
}

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

func TestRenderToFileInvalidPath(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	invalidPath := filepath.Join(string([]byte{0}), "invalid")
	err := renderer.RenderToFile("test.tmpl", TemplateData{}, invalidPath)
	require.Error(t, err)
}

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

func TestRenderLineEndings(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	result, err := renderer.Render("test.tmpl", TemplateData{})
	require.NoError(t, err)

	assert.True(t, strings.Contains(result, "\n"), "should contain newlines")
}

func TestValidate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
		errType      interface{}
	}{
		{
			name:         "valid template - go.mod",
			templateName: "go.mod.tmpl",
			wantErr:      false,
		},
		{
			name:         "valid template - .gitignore",
			templateName: ".gitignore.tmpl",
			wantErr:      false,
		},
		{
			name:         "valid template - main.go",
			templateName: "cmd/server/main.go.tmpl",
			wantErr:      false,
		},
		{
			name:         "valid template - tracks.yaml",
			templateName: "tracks.yaml.tmpl",
			wantErr:      false,
		},
		{
			name:         "valid template - .env.example",
			templateName: ".env.example.tmpl",
			wantErr:      false,
		},
		{
			name:         "valid template - README.md",
			templateName: "README.md.tmpl",
			wantErr:      false,
		},
		{
			name:         "non-existent template",
			templateName: "nonexistent.tmpl",
			wantErr:      true,
			errType:      &TemplateError{},
		},
		{
			name:         "invalid path",
			templateName: "../../../etc/passwd",
			wantErr:      true,
			errType:      &TemplateError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderer.Validate(tt.templateName)
			if tt.wantErr {
				require.Error(t, err, "Validate should return error")
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err, "error should be correct type")
				}
				var te *TemplateError
				var ve *ValidationError
				if errors.As(err, &te) {
					assert.Contains(t, te.Error(), tt.templateName, "error should include template name")
				} else if errors.As(err, &ve) {
					assert.Contains(t, ve.Error(), tt.templateName, "error should include template name")
				}
			} else {
				require.NoError(t, err, "Validate should not return error for valid template")
			}
		})
	}
}

func TestValidateErrorMessages(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name            string
		templateName    string
		wantErrContains string
	}{
		{
			name:            "non-existent template includes name",
			templateName:    "missing.tmpl",
			wantErrContains: "missing.tmpl",
		},
		{
			name:            "invalid path includes name",
			templateName:    "invalid/path.tmpl",
			wantErrContains: "invalid/path.tmpl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderer.Validate(tt.templateName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}
