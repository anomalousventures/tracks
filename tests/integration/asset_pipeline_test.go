package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "asset-pipeline-test"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/asset-pipeline-test",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "generation should succeed")

	projectRoot := filepath.Join(tmpDir, projectName)

	t.Run("TailwindCSS compilation (#446)", func(t *testing.T) {
		appCSS := filepath.Join(projectRoot, "internal", "assets", "web", "css", "app.css")
		content, err := os.ReadFile(appCSS)
		require.NoError(t, err, "app.css should exist")

		cssStr := string(content)

		assert.Contains(t, cssStr, "@import \"tailwindcss\"",
			"source CSS should use Tailwind v4 @import syntax")
		assert.Contains(t, cssStr, "@theme inline {",
			"source CSS should contain @theme inline directive for templUI")
		assert.Contains(t, cssStr, ":root {",
			"source CSS should define CSS custom properties in :root")
		assert.Contains(t, cssStr, "--primary",
			"source CSS should define templUI-compatible theme variables")
		assert.Contains(t, cssStr, ".dark {",
			"source CSS should define dark mode theme")
	})

	t.Run("JavaScript bundling (#447)", func(t *testing.T) {
		appJS := filepath.Join(projectRoot, "internal", "assets", "web", "js", "app.js")
		content, err := os.ReadFile(appJS)
		require.NoError(t, err, "app.js should exist")

		jsStr := string(content)

		assert.Contains(t, jsStr, "'use strict'",
			"JavaScript should use strict mode")
		assert.Contains(t, jsStr, "const App = {",
			"JavaScript should define App object")
		assert.Contains(t, jsStr, "init()",
			"JavaScript App should have init method")
		assert.Contains(t, jsStr, "(function() {",
			"JavaScript should use IIFE pattern to avoid global pollution")
		assert.Contains(t, jsStr, "import './lib/htmx.js'",
			"JavaScript should import HTMX library")

		htmxJS := filepath.Join(projectRoot, "internal", "assets", "web", "js", "lib", "htmx.js")
		_, err = os.Stat(htmxJS)
		require.NoError(t, err, "htmx.js library file should exist")
	})

	t.Run("hashfs content-addressed URLs (#448)", func(t *testing.T) {
		embedGo := filepath.Join(projectRoot, "internal", "assets", "embed.go")
		content, err := os.ReadFile(embedGo)
		require.NoError(t, err, "embed.go should exist")

		embedStr := string(content)

		assert.Contains(t, embedStr, "github.com/benbjohnson/hashfs",
			"embed.go should import hashfs package")
		assert.Contains(t, embedStr, "hashfs.NewFS",
			"embed.go should use hashfs.NewFS for content-addressed serving")
		assert.Contains(t, embedStr, "func AssetURL(path string) string",
			"embed.go should export AssetURL helper function")
		assert.Contains(t, embedStr, "func CSSURL() string",
			"embed.go should export CSSURL helper function")
		assert.Contains(t, embedStr, "func JSURL() string",
			"embed.go should export JSURL helper function")
		assert.Contains(t, embedStr, "fsys.HashName(path)",
			"AssetURL should use hashfs HashName for content-addressed URLs")
	})

	t.Run("middleware chain (#449)", func(t *testing.T) {
		middlewareDir := filepath.Join(projectRoot, "internal", "http", "middleware")

		compressGo := filepath.Join(middlewareDir, "compress.go")
		compressContent, err := os.ReadFile(compressGo)
		require.NoError(t, err, "compress.go should exist")
		compressStr := string(compressContent)
		assert.Contains(t, compressStr, "middleware.Compress",
			"compress.go should use chi compress middleware")

		securityGo := filepath.Join(middlewareDir, "security.go")
		securityContent, err := os.ReadFile(securityGo)
		require.NoError(t, err, "security.go should exist")
		securityStr := string(securityContent)
		assert.Contains(t, securityStr, "templ.WithNonce",
			"security.go should use templ CSP nonce middleware")
		assert.Contains(t, securityStr, "NewCSPMiddleware",
			"security.go should export CSP middleware constructor")

		loggingGo := filepath.Join(middlewareDir, "logging.go")
		loggingContent, err := os.ReadFile(loggingGo)
		require.NoError(t, err, "logging.go should exist")
		loggingStr := string(loggingContent)
		assert.Contains(t, loggingStr, "Logger",
			"logging.go should have logger middleware")

		cacheGo := filepath.Join(middlewareDir, "cache.go")
		cacheContent, err := os.ReadFile(cacheGo)
		require.NoError(t, err, "cache.go should exist")
		cacheStr := string(cacheContent)
		assert.Contains(t, cacheStr, "Cache-Control",
			"cache.go should set cache headers")
	})

	t.Run("HTMX functionality (#450)", func(t *testing.T) {
		counterTempl := filepath.Join(projectRoot, "internal", "http", "views", "components", "counter.templ")
		counterContent, err := os.ReadFile(counterTempl)
		require.NoError(t, err, "counter.templ should exist")

		counterStr := string(counterContent)

		assert.Contains(t, counterStr, "hx-post",
			"counter component should use hx-post for HTMX requests")
		assert.Contains(t, counterStr, "hx-swap",
			"counter component should use hx-swap for partial updates")

		htmxJS := filepath.Join(projectRoot, "internal", "assets", "web", "js", "lib", "htmx.js")
		_, err = os.Stat(htmxJS)
		require.NoError(t, err, "htmx.js library file should exist")
	})

	t.Run("Air live reload config (#451)", func(t *testing.T) {
		airToml := filepath.Join(projectRoot, ".air.toml")
		content, err := os.ReadFile(airToml)
		require.NoError(t, err, ".air.toml should exist")

		var cfg map[string]any
		err = toml.Unmarshal(content, &cfg)
		require.NoError(t, err, ".air.toml should be valid TOML")

		build, ok := cfg["build"].(map[string]any)
		require.True(t, ok, ".air.toml should have [build] section")

		includeExt, ok := build["include_ext"].([]any)
		require.True(t, ok, "build.include_ext should be an array")
		includeExtStr := toStringSlice(includeExt)
		assert.Contains(t, includeExtStr, "css", "Air should watch CSS files")
		assert.Contains(t, includeExtStr, "js", "Air should watch JS files")
		assert.Contains(t, includeExtStr, "templ", "Air should watch templ files")
		assert.Contains(t, includeExtStr, "go", "Air should watch Go files")

		excludeDir, ok := build["exclude_dir"].([]any)
		require.True(t, ok, "build.exclude_dir should be an array")
		excludeDirStr := toStringSlice(excludeDir)
		assert.Contains(t, excludeDirStr, "internal/assets/dist",
			"Air should exclude dist directory to prevent rebuild loops")
		assert.Contains(t, excludeDirStr, "internal/db/generated",
			"Air should exclude generated SQLC code")
		assert.Contains(t, excludeDirStr, "tests/mocks",
			"Air should exclude generated mocks")

		excludeRegex, ok := build["exclude_regex"].([]any)
		require.True(t, ok, "build.exclude_regex should be an array")
		excludeRegexStr := toStringSlice(excludeRegex)
		assert.Contains(t, excludeRegexStr, "_templ.go",
			"Air should exclude generated templ files")
		assert.Contains(t, excludeRegexStr, "_test.go",
			"Air should exclude test files")

		preCmd, ok := build["pre_cmd"].([]any)
		require.True(t, ok, "build.pre_cmd should be an array")
		preCmdStr := toStringSlice(preCmd)
		found := false
		for _, cmd := range preCmdStr {
			if strings.Contains(cmd, "make generate assets") {
				found = true
				break
			}
		}
		assert.True(t, found, "Air pre_cmd should include 'make generate assets'")
	})

	t.Run("templUI configuration (#465)", func(t *testing.T) {
		templuiConfig := filepath.Join(projectRoot, ".templui.json")
		content, err := os.ReadFile(templuiConfig)
		require.NoError(t, err, ".templui.json should exist")

		configStr := string(content)
		assert.Contains(t, configStr, "componentsDir",
			".templui.json should define components directory")
		assert.Contains(t, configStr, "internal/http/views/components/ui",
			".templui.json should point to correct ui directory")
		assert.Contains(t, configStr, "internal/http/views/components/utils",
			".templui.json should point to correct utils directory")
		assert.Contains(t, configStr, "github.com/test/asset-pipeline-test",
			".templui.json should contain module name")
		assert.Contains(t, configStr, "internal/assets/web/js",
			".templui.json should define jsDir for JavaScript assets")
	})

	t.Run("complete asset pipeline (#452)", func(t *testing.T) {
		webCSSDir := filepath.Join(projectRoot, "internal", "assets", "web", "css")
		webJSDir := filepath.Join(projectRoot, "internal", "assets", "web", "js")
		distDir := filepath.Join(projectRoot, "internal", "assets", "dist")

		_, err := os.Stat(webCSSDir)
		require.NoError(t, err, "web/css source directory should exist")

		_, err = os.Stat(webJSDir)
		require.NoError(t, err, "web/js source directory should exist")

		_, err = os.Stat(distDir)
		require.NoError(t, err, "dist output directory should exist")

		makefile := filepath.Join(projectRoot, "Makefile")
		makeContent, err := os.ReadFile(makefile)
		require.NoError(t, err, "Makefile should exist")

		makeStr := string(makeContent)
		assert.Contains(t, makeStr, "assets:",
			"Makefile should have 'assets' target")
		assert.Contains(t, makeStr, "css:",
			"Makefile should have 'css' target")
		assert.Contains(t, makeStr, "js:",
			"Makefile should have 'js' target")
		assert.Contains(t, makeStr, "tailwindcss",
			"Makefile should use tailwindcss for CSS compilation")
		assert.Contains(t, makeStr, "esbuild",
			"Makefile should use esbuild for JS bundling")

		goMod := filepath.Join(projectRoot, "go.mod")
		modContent, err := os.ReadFile(goMod)
		require.NoError(t, err, "go.mod should exist")
		modStr := string(modContent)
		assert.Contains(t, modStr, "github.com/benbjohnson/hashfs",
			"go.mod should include hashfs dependency")
	})
}

func toStringSlice(arr []any) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
