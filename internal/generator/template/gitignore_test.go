package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitignoreTemplateRenders(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/user/myapp",
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		GoVersion:   "1.25",
		Year:        2025,
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err, "template should render without errors")
	assert.NotEmpty(t, result, "template should produce non-empty output")
}

func TestGitignoreExcludesEnvironmentFiles(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)

	t.Run("excludes .env file", func(t *testing.T) {
		assert.Contains(t, result, ".env", "should exclude .env file")
	})

	t.Run("includes .env.example", func(t *testing.T) {
		assert.Contains(t, result, "!.env.example", "should NOT exclude .env.example (using negation pattern)")
	})
}

func TestGitignoreExcludesBuildArtifacts(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)

	t.Run("excludes bin directory", func(t *testing.T) {
		assert.Contains(t, result, "bin/", "should exclude bin/ directory")
	})

	t.Run("excludes compiled binaries", func(t *testing.T) {
		assert.Contains(t, result, "*.exe", "should exclude Windows executables")
		assert.Contains(t, result, "*.dll", "should exclude Windows DLLs")
		assert.Contains(t, result, "*.so", "should exclude Linux shared objects")
		assert.Contains(t, result, "*.dylib", "should exclude macOS dynamic libraries")
	})

	t.Run("excludes coverage files", func(t *testing.T) {
		assert.Contains(t, result, "coverage.out", "should exclude coverage.out")
		assert.Contains(t, result, "coverage.html", "should exclude coverage.html")
	})
}

func TestGitignoreExcludesIDEFiles(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)

	t.Run("excludes VSCode directory", func(t *testing.T) {
		assert.Contains(t, result, ".vscode/", "should exclude .vscode/ directory")
	})

	t.Run("excludes IntelliJ directory", func(t *testing.T) {
		assert.Contains(t, result, ".idea/", "should exclude .idea/ directory")
	})

	t.Run("excludes vim swap files", func(t *testing.T) {
		assert.Contains(t, result, "*.swp", "should exclude vim swap files")
	})
}

func TestGitignoreExcludesOSFiles(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)

	t.Run("excludes macOS DS_Store", func(t *testing.T) {
		assert.Contains(t, result, ".DS_Store", "should exclude .DS_Store files")
	})

	t.Run("excludes Windows Thumbs.db", func(t *testing.T) {
		assert.Contains(t, result, "Thumbs.db", "should exclude Thumbs.db files")
	})
}

func TestGitignoreIsStatic(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data1 := TemplateData{
		ModuleName:  "github.com/user/app1",
		ProjectName: "app1",
		DBDriver:    "sqlite3",
	}

	data2 := TemplateData{
		ModuleName:  "gitlab.com/org/app2",
		ProjectName: "app2",
		DBDriver:    "postgres",
	}

	result1, err1 := renderer.Render(".gitignore.tmpl", data1)
	require.NoError(t, err1)

	result2, err2 := renderer.Render(".gitignore.tmpl", data2)
	require.NoError(t, err2)

	t.Run("content is identical regardless of template data", func(t *testing.T) {
		assert.Equal(t, result1, result2, "gitignore should be static and not use template variables")
	})
}
