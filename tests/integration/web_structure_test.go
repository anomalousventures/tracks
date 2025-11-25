package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebDirectoryStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "web-structure-test"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/web-structure-test",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()

	err := gen.Validate(cfg)
	require.NoError(t, err, "validation should succeed")

	err = gen.Generate(ctx, cfg)
	require.NoError(t, err, "generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Run("internal/assets directory exists", func(t *testing.T) {
		assetsDir := filepath.Join(projectRoot, "internal", "assets")
		info, err := os.Stat(assetsDir)
		require.NoError(t, err, "internal/assets/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/ should be a directory")
	})

	t.Run("internal/assets/web/css directory exists with app.css", func(t *testing.T) {
		cssDir := filepath.Join(projectRoot, "internal", "assets", "web", "css")
		info, err := os.Stat(cssDir)
		require.NoError(t, err, "internal/assets/web/css/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/web/css/ should be a directory")

		appCSS := filepath.Join(cssDir, "app.css")
		content, err := os.ReadFile(appCSS)
		require.NoError(t, err, "app.css should exist and be readable")

		contentStr := string(content)
		assert.Contains(t, contentStr, "Tailwind CSS v4 entry point", "app.css should contain TailwindCSS v4 comment")
		assert.Contains(t, contentStr, "@import \"tailwindcss\"", "app.css should use Tailwind v4 @import")
		assert.Contains(t, contentStr, "@theme {", "app.css should use Tailwind v4 @theme directive")
		assert.Contains(t, contentStr, "--color-primary-500", "app.css should define theme variables")
	})

	t.Run("internal/assets/web/js directory exists with app.js", func(t *testing.T) {
		jsDir := filepath.Join(projectRoot, "internal", "assets", "web", "js")
		info, err := os.Stat(jsDir)
		require.NoError(t, err, "internal/assets/web/js/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/web/js/ should be a directory")

		appJS := filepath.Join(jsDir, "app.js")
		content, err := os.ReadFile(appJS)
		require.NoError(t, err, "app.js should exist and be readable")

		contentStr := string(content)
		assert.Contains(t, contentStr, "Main JavaScript Application", "app.js should contain main application comment")
		assert.Contains(t, contentStr, "name: 'web-structure-test'", "app.js should contain project name in App object")
		assert.Contains(t, contentStr, "(function() {", "app.js should use IIFE pattern")
		assert.Contains(t, contentStr, "const App = {", "app.js should define App object")
	})

	t.Run("internal/assets/web/images directory exists with .gitkeep", func(t *testing.T) {
		imagesDir := filepath.Join(projectRoot, "internal", "assets", "web", "images")
		info, err := os.Stat(imagesDir)
		require.NoError(t, err, "internal/assets/web/images/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/web/images/ should be a directory")

		gitkeep := filepath.Join(imagesDir, ".gitkeep")
		info, err = os.Stat(gitkeep)
		require.NoError(t, err, ".gitkeep should exist in internal/assets/web/images/")
		assert.False(t, info.IsDir(), ".gitkeep should be a file, not a directory")
	})

	t.Run("internal/assets/dist directory structure exists", func(t *testing.T) {
		distDir := filepath.Join(projectRoot, "internal", "assets", "dist")
		info, err := os.Stat(distDir)
		require.NoError(t, err, "internal/assets/dist/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/dist/ should be a directory")

		// Check subdirectories
		cssDir := filepath.Join(distDir, "css")
		info, err = os.Stat(cssDir)
		require.NoError(t, err, "internal/assets/dist/css/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/dist/css/ should be a directory")

		jsDir := filepath.Join(distDir, "js")
		info, err = os.Stat(jsDir)
		require.NoError(t, err, "internal/assets/dist/js/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/dist/js/ should be a directory")

		imagesDir := filepath.Join(distDir, "images")
		info, err = os.Stat(imagesDir)
		require.NoError(t, err, "internal/assets/dist/images/ directory should exist")
		assert.True(t, info.IsDir(), "internal/assets/dist/images/ should be a directory")

		gitkeep := filepath.Join(distDir, ".gitkeep")
		info, err = os.Stat(gitkeep)
		require.NoError(t, err, ".gitkeep should exist in internal/assets/dist/")
		assert.False(t, info.IsDir(), ".gitkeep should be a file, not a directory")
	})

	t.Run("web assets are valid", func(t *testing.T) {
		appCSS := filepath.Join(projectRoot, "internal", "assets", "web", "css", "app.css")
		cssContent, err := os.ReadFile(appCSS)
		require.NoError(t, err)

		cssStr := string(cssContent)
		assert.Contains(t, cssStr, "/*", "CSS should contain comment syntax")
		assert.Contains(t, cssStr, "@layer components", "CSS should define component layer")
		assert.Contains(t, cssStr, "@layer utilities", "CSS should define utilities layer")

		appJS := filepath.Join(projectRoot, "internal", "assets", "web", "js", "app.js")
		jsContent, err := os.ReadFile(appJS)
		require.NoError(t, err)

		jsStr := string(jsContent)
		assert.Contains(t, jsStr, "/**", "JavaScript should contain JSDoc comment syntax")
		assert.Contains(t, jsStr, "'use strict'", "JavaScript should use strict mode")
		assert.Contains(t, jsStr, "setupEventListeners", "JavaScript should have event listener setup")
	})
}

func TestWebDirectoryWithDifferentProjectNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		name        string
		projectName string
	}{
		{
			name:        "project with hyphens",
			projectName: "my-awesome-app",
		},
		{
			name:        "project with underscores",
			projectName: "my_awesome_app",
		},
		{
			name:        "simple project name",
			projectName: "simpleapp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := generator.ProjectConfig{
				ProjectName:    tc.projectName,
				ModulePath:     "github.com/test/" + tc.projectName,
				DatabaseDriver: "sqlite3",
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed for project name: %s", tc.projectName)

			projectRoot := filepath.Join(tmpDir, tc.projectName)

			assetsDir := filepath.Join(projectRoot, "internal", "assets")
			_, err = os.Stat(assetsDir)
			require.NoError(t, err, "internal/assets/ directory should exist for project: %s", tc.projectName)

			appJS := filepath.Join(projectRoot, "internal", "assets", "web", "js", "app.js")
			content, err := os.ReadFile(appJS)
			require.NoError(t, err, "app.js should exist for project: %s", tc.projectName)

			contentStr := string(content)
			assert.Contains(t, contentStr, "name: '"+tc.projectName+"'",
				"app.js should contain correct project name in App object")
			assert.Contains(t, contentStr, tc.projectName+" - Main JavaScript Application",
				"app.js should contain correct project name in header comment")
		})
	}
}
