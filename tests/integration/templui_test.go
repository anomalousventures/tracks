package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplUIInitDuringGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		databaseDriver string
	}{
		{name: "sqlite3", databaseDriver: "sqlite3"},
		{name: "postgres", databaseDriver: "postgres"},
		{name: "go-libsql", databaseDriver: "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && tt.databaseDriver != "sqlite3" {
				t.Skip("skipping non-sqlite3 driver on Windows (slow go mod download)")
			}

			tmpDir := t.TempDir()
			projectName := "templui-init-test"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     "github.com/test/templui-init-test",
				DatabaseDriver: tt.databaseDriver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)

			t.Run(".templui.json exists with correct configuration (#482)", func(t *testing.T) {
				configPath := filepath.Join(projectRoot, ".templui.json")
				content, err := os.ReadFile(configPath)
				require.NoError(t, err, ".templui.json should exist")

				configStr := string(content)
				assert.Contains(t, configStr, "componentsDir",
					".templui.json should define components directory")
				assert.Contains(t, configStr, "internal/http/views/components/ui",
					".templui.json should point to correct ui directory")
				assert.Contains(t, configStr, "github.com/test/templui-init-test",
					".templui.json should contain module name")
			})

			t.Run("components directory structure exists (#482)", func(t *testing.T) {
				uiDir := filepath.Join(projectRoot, "internal", "http", "views", "components", "ui")
				stat, err := os.Stat(uiDir)
				require.NoError(t, err, "ui components directory should exist")
				assert.True(t, stat.IsDir(), "ui should be a directory")

				alertDir := filepath.Join(uiDir, "alert")
				_, err = os.Stat(alertDir)
				require.NoError(t, err, "alert component directory should exist")

				buttonDir := filepath.Join(uiDir, "button")
				_, err = os.Stat(buttonDir)
				require.NoError(t, err, "button component directory should exist")

				cardDir := filepath.Join(uiDir, "card")
				_, err = os.Stat(cardDir)
				require.NoError(t, err, "card component directory should exist")
			})

			t.Run("script injection markers present in base.templ (#482)", func(t *testing.T) {
				baseTempl := filepath.Join(projectRoot, "internal", "http", "views", "layouts", "base.templ")
				content, err := os.ReadFile(baseTempl)
				require.NoError(t, err, "base.templ should exist")

				baseStr := string(content)
				assert.Contains(t, baseStr, "<!-- TRACKS:UI_SCRIPTS:BEGIN -->",
					"base.templ should contain script begin marker")
				assert.Contains(t, baseStr, "<!-- TRACKS:UI_SCRIPTS:END -->",
					"base.templ should contain script end marker")
			})
		})
	}
}

func TestTemplUIPagesRender(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		databaseDriver string
	}{
		{name: "sqlite3", databaseDriver: "sqlite3"},
		{name: "postgres", databaseDriver: "postgres"},
		{name: "go-libsql", databaseDriver: "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && tt.databaseDriver != "sqlite3" {
				t.Skip("skipping non-sqlite3 driver on Windows (slow go mod download)")
			}

			tmpDir := t.TempDir()
			projectName := "templui-render-test"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     "github.com/test/templui-render-test",
				DatabaseDriver: tt.databaseDriver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)

			t.Run("error page uses templUI Alert component (#483)", func(t *testing.T) {
				errorTempl := filepath.Join(projectRoot, "internal", "http", "views", "pages", "error.templ")
				content, err := os.ReadFile(errorTempl)
				require.NoError(t, err, "error.templ should exist")

				errorStr := string(content)
				assert.Contains(t, errorStr, "ui/alert",
					"error page should import ui/alert package")
				assert.Contains(t, errorStr, "alert.Alert",
					"error page should use alert.Alert component")
				assert.Contains(t, errorStr, "alert.VariantDestructive",
					"error page should use destructive variant for errors")
				assert.Contains(t, errorStr, "alert.Title",
					"error page should use alert.Title subcomponent")
				assert.Contains(t, errorStr, "alert.Description",
					"error page should use alert.Description subcomponent")
			})

			t.Run("counter component uses templUI Button and Card (#483)", func(t *testing.T) {
				counterTempl := filepath.Join(projectRoot, "internal", "http", "views", "components", "counter.templ")
				content, err := os.ReadFile(counterTempl)
				require.NoError(t, err, "counter.templ should exist")

				counterStr := string(content)
				assert.Contains(t, counterStr, "ui/button",
					"counter component should import ui/button package")
				assert.Contains(t, counterStr, "ui/card",
					"counter component should import ui/card package")
				assert.Contains(t, counterStr, "button.Button",
					"counter component should use button.Button")
				assert.Contains(t, counterStr, "card.Card",
					"counter component should use card.Card")
			})

			t.Run("HTMX attributes preserved on counter buttons (#483)", func(t *testing.T) {
				counterTempl := filepath.Join(projectRoot, "internal", "http", "views", "components", "counter.templ")
				content, err := os.ReadFile(counterTempl)
				require.NoError(t, err, "counter.templ should exist")

				counterStr := string(content)
				assert.Contains(t, counterStr, "hx-post",
					"counter buttons should have hx-post attribute")
				assert.Contains(t, counterStr, "hx-target",
					"counter buttons should have hx-target attribute")
				assert.Contains(t, counterStr, "hx-swap",
					"counter buttons should have hx-swap attribute")
				assert.Contains(t, counterStr, "CounterIncrement",
					"counter should reference routes.CounterIncrement")
				assert.Contains(t, counterStr, "CounterDecrement",
					"counter should reference routes.CounterDecrement")
				assert.Contains(t, counterStr, "CounterReset",
					"counter should reference routes.CounterReset")
			})

			t.Run("home page imports components package (#483)", func(t *testing.T) {
				homeTempl := filepath.Join(projectRoot, "internal", "http", "views", "pages", "home.templ")
				content, err := os.ReadFile(homeTempl)
				require.NoError(t, err, "home.templ should exist")

				homeStr := string(content)
				assert.Contains(t, homeStr, "views/components",
					"home page should import components package")
				assert.Contains(t, homeStr, "components.CounterCard",
					"home page should use CounterCard component")
			})

			t.Run("templUI component files are valid templ syntax (#483)", func(t *testing.T) {
				uiDir := filepath.Join(projectRoot, "internal", "http", "views", "components", "ui")

				templFiles := []string{
					filepath.Join(uiDir, "alert", "alert.templ"),
					filepath.Join(uiDir, "button", "button.templ"),
					filepath.Join(uiDir, "card", "card.templ"),
				}

				for _, templFile := range templFiles {
					content, err := os.ReadFile(templFile)
					if os.IsNotExist(err) {
						t.Logf("Skipping %s (not installed)", filepath.Base(filepath.Dir(templFile)))
						continue
					}
					require.NoError(t, err, "%s should be readable", templFile)

					fileStr := string(content)
					assert.True(t, strings.Contains(fileStr, "package") ||
						strings.Contains(fileStr, "templ "),
						"%s should contain valid templ syntax", filepath.Base(templFile))
				}
			})
		})
	}
}
