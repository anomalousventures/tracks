package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitignoreTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render(".gitignore.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	expectedPatterns := []string{
		"bin/",
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",
		"coverage.out",
		"coverage.html",
		".env",
		"!.env.example",
		".DS_Store",
		"Thumbs.db",
		".vscode/",
		".idea/",
		"*.swp",
	}

	for _, pattern := range expectedPatterns {
		assert.Contains(t, result, pattern, "should contain pattern: %s", pattern)
	}
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
