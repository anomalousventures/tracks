package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var productionTemplates = []string{
	"go.mod.tmpl",
	".gitignore.tmpl",
	"cmd/server/main.go.tmpl",
	"tracks.yaml.tmpl",
	".env.example.tmpl",
	"README.md.tmpl",
}

func TestRenderAllTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/user/myapp",
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		GoVersion:   "1.25",
		Year:        2025,
	}

	templateFiles := []struct {
		template   string
		outputPath string
		contains   []string
	}{
		{
			template:   "go.mod.tmpl",
			outputPath: "go.mod",
			contains:   []string{"module github.com/user/myapp", "go 1.25"},
		},
		{
			template:   ".gitignore.tmpl",
			outputPath: ".gitignore",
			contains:   []string{"bin/", "*.exe", "coverage.out", ".env"},
		},
		{
			template:   "cmd/server/main.go.tmpl",
			outputPath: filepath.Join("cmd", "server", "main.go"),
			contains:   []string{"package main", "func run() error", "config.Load()", "logging.NewLogger"},
		},
		{
			template:   "tracks.yaml.tmpl",
			outputPath: "tracks.yaml",
			contains:   []string{"database:", "server:", `port: ":8080"`, "logging:"},
		},
		{
			template:   ".env.example.tmpl",
			outputPath: ".env.example",
			contains:   []string{"WARNING", "DATABASE_URL", "APP_SERVER_PORT"},
		},
		{
			template:   "README.md.tmpl",
			outputPath: "README.md",
			contains:   []string{"# myapp", "Generated with [Tracks]", "## Setup", "## Development"},
		},
	}

	for _, tf := range templateFiles {
		t.Run(tf.template, func(t *testing.T) {
			fullOutputPath := filepath.Join(tmpDir, tf.outputPath)
			err := renderer.RenderToFile(tf.template, data, fullOutputPath)
			require.NoError(t, err, "rendering %s should not fail", tf.template)

			_, err = os.Stat(fullOutputPath)
			require.NoError(t, err, "file should exist at %s", fullOutputPath)

			content, err := os.ReadFile(fullOutputPath)
			require.NoError(t, err, "should be able to read file")

			for _, expectedContent := range tf.contains {
				assert.Contains(t, string(content), expectedContent, "file should contain expected content")
			}
		})
	}

	t.Run("verify directory structure", func(t *testing.T) {
		cmdDir := filepath.Join(tmpDir, "cmd")
		serverDir := filepath.Join(tmpDir, "cmd", "server")

		_, err := os.Stat(cmdDir)
		require.NoError(t, err, "cmd directory should exist")

		_, err = os.Stat(serverDir)
		require.NoError(t, err, "cmd/server directory should exist")
	})

	t.Run("verify all files exist", func(t *testing.T) {
		expectedFiles := []string{
			"go.mod",
			".gitignore",
			filepath.Join("cmd", "server", "main.go"),
			"tracks.yaml",
			".env.example",
			"README.md",
		}

		for _, file := range expectedFiles {
			fullPath := filepath.Join(tmpDir, file)
			_, err := os.Stat(fullPath)
			require.NoError(t, err, "file %s should exist", file)
		}
	})
}

func TestRenderAllTemplatesWithDifferentDrivers(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run("driver="+driver, func(t *testing.T) {
			tmpDir := t.TempDir()

			data := TemplateData{
				ModuleName:  "github.com/test/app",
				ProjectName: "app",
				DBDriver:    driver,
				GoVersion:   "1.25",
				Year:        2025,
			}

			tracksYamlPath := filepath.Join(tmpDir, "tracks.yaml")
			err := renderer.RenderToFile("tracks.yaml.tmpl", data, tracksYamlPath)
			require.NoError(t, err)

			content, err := os.ReadFile(tracksYamlPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), "database:")
			assert.Contains(t, string(content), "url:")
		})
	}
}

func TestRenderAllTemplatesConsistency(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/org/project",
		ProjectName: "project",
		DBDriver:    "postgres",
		GoVersion:   "1.25",
		Year:        2025,
	}

	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	for _, tmpl := range productionTemplates {
		outputName := strings.TrimSuffix(tmpl, ".tmpl")
		path1 := filepath.Join(tmpDir1, outputName)
		path2 := filepath.Join(tmpDir2, outputName)

		err := renderer.RenderToFile(tmpl, data, path1)
		require.NoError(t, err)

		err = renderer.RenderToFile(tmpl, data, path2)
		require.NoError(t, err)

		content1, err := os.ReadFile(path1)
		require.NoError(t, err)

		content2, err := os.ReadFile(path2)
		require.NoError(t, err)

		assert.Equal(t, content1, content2, "rendering %s should be consistent", tmpl)
	}
}

func TestCrossPlatformIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName:  "github.com/user/myapp",
		ProjectName: "myapp",
		GoVersion:   "1.25",
	}

	t.Run("nested directory creation", func(t *testing.T) {
		mainPath := filepath.Join(tmpDir, "cmd", "server", "main.go")
		err := renderer.RenderToFile("cmd/server/main.go.tmpl", data, mainPath)
		require.NoError(t, err, "should create nested directories")

		_, err = os.Stat(mainPath)
		require.NoError(t, err, "file should exist at nested path")

		info, err := os.Stat(filepath.Join(tmpDir, "cmd"))
		require.NoError(t, err, "cmd directory should exist")
		assert.True(t, info.IsDir(), "cmd should be a directory")

		info, err = os.Stat(filepath.Join(tmpDir, "cmd", "server"))
		require.NoError(t, err, "cmd/server directory should exist")
		assert.True(t, info.IsDir(), "cmd/server should be a directory")
	})

	t.Run("path separators are correct", func(t *testing.T) {
		mainPath := filepath.Join(tmpDir, "test", "nested", "file.txt")
		err := renderer.RenderToFile("test.tmpl", data, mainPath)
		require.NoError(t, err)

		_, err = os.Stat(mainPath)
		require.NoError(t, err, "file should exist with correct path separators")
	})

	t.Run("multiple levels of nesting", func(t *testing.T) {
		deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "file.txt")
		err := renderer.RenderToFile("test.tmpl", data, deepPath)
		require.NoError(t, err, "should create deeply nested directories")

		_, err = os.Stat(deepPath)
		require.NoError(t, err, "file should exist at deeply nested path")
	})
}
