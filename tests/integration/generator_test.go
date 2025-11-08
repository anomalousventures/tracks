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

func TestGenerateFullProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		databaseDriver string
		initGit        bool
		modulePath     string
	}{
		{
			name:           "go-libsql without git",
			databaseDriver: "go-libsql",
			initGit:        false,
			modulePath:     "github.com/test/libsql-app",
		},
		{
			name:           "sqlite3 with git",
			databaseDriver: "sqlite3",
			initGit:        true,
			modulePath:     "github.com/test/sqlite-app",
		},
		{
			name:           "postgres without git",
			databaseDriver: "postgres",
			initGit:        false,
			modulePath:     "example.com/postgres-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := "testapp"

			cfg := generator.ProjectConfig{
				ProjectName:    projectName,
				ModulePath:     tt.modulePath,
				DatabaseDriver: tt.databaseDriver,
				EnvPrefix:      "APP",
				InitGit:        tt.initGit,
				OutputPath:     tmpDir,
			}

			gen := generator.NewProjectGenerator()
			ctx := context.Background()

			err := gen.Validate(cfg)
			require.NoError(t, err, "validation should succeed")

			err = gen.Generate(ctx, cfg)
			require.NoError(t, err, "generation should succeed")

			projectRoot := filepath.Join(tmpDir, projectName)

			expectedFiles := []string{
				"go.mod",
				"README.md",
				".gitignore",
				".golangci.yml",
				".mockery.yaml",
				".tracks.yaml",
				".env.example",
				"Makefile",
				"sqlc.yaml",
				"cmd/server/main.go",
				"internal/config/config.go",
				"internal/interfaces/health.go",
				"internal/interfaces/logger.go",
				"internal/logging/logger.go",
				"internal/domain/health/service.go",
				"internal/http/server.go",
				"internal/http/routes.go",
				"internal/http/routes/routes.go",
				"internal/http/handlers/health.go",
				"internal/http/middleware/logging.go",
				"internal/db/db.go",
			}

			for _, file := range expectedFiles {
				path := filepath.Join(projectRoot, file)
				_, err := os.Stat(path)
				assert.NoError(t, err, "file should exist: %s", file)
			}

			expectedDirs := []string{
				"cmd",
				"cmd/server",
				"internal",
				"internal/config",
				"internal/interfaces",
				"internal/logging",
				"internal/domain",
				"internal/domain/health",
				"internal/http",
				"internal/http/routes",
				"internal/http/handlers",
				"internal/http/middleware",
				"internal/db",
			}

			for _, dir := range expectedDirs {
				path := filepath.Join(projectRoot, dir)
				stat, err := os.Stat(path)
				assert.NoError(t, err, "directory should exist: %s", dir)
				if err == nil {
					assert.True(t, stat.IsDir(), "%s should be a directory", dir)
				}
			}

			goModPath := filepath.Join(projectRoot, "go.mod")
			content, err := os.ReadFile(goModPath)
			require.NoError(t, err, "should be able to read go.mod")

			assert.Contains(t, string(content), tt.modulePath, "go.mod should contain module path")
			assert.Contains(t, string(content), "go 1.25", "go.mod should contain Go version")

			dbPath := filepath.Join(projectRoot, "internal/db/db.go")
			dbContent, err := os.ReadFile(dbPath)
			require.NoError(t, err, "should be able to read internal/db/db.go")

			driverMapping := map[string]string{
				"go-libsql": "libsql",
				"sqlite3":   "sqlite3",
				"postgres":  "postgres",
			}
			expectedDriver := driverMapping[tt.databaseDriver]
			assert.Contains(t, string(dbContent), expectedDriver, "db.go should contain correct driver")

			if tt.initGit {
				gitDir := filepath.Join(projectRoot, ".git")
				stat, err := os.Stat(gitDir)
				assert.NoError(t, err, ".git directory should exist when git is initialized")
				if err == nil {
					assert.True(t, stat.IsDir(), ".git should be a directory")
				}
			} else {
				gitDir := filepath.Join(projectRoot, ".git")
				_, err := os.Stat(gitDir)
				assert.True(t, os.IsNotExist(err), ".git directory should not exist when git is not initialized")
			}
		})
	}
}

func TestGenerateFullProject_CustomModuleName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "customapp"
	modulePath := "mycorp.com/apps/customapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     modulePath,
		DatabaseDriver: "postgres",
		EnvPrefix:      "CUSTOM",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, projectName)
	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), modulePath)
}

func TestGenerateFullProject_DirectoryAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	projectName := "existing"

	existingDir := filepath.Join(tmpDir, projectName)
	err := os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     "github.com/test/existing",
		DatabaseDriver: "postgres",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()

	err = gen.Validate(cfg)
	assert.Error(t, err, "validation should fail when directory exists")
	assert.Contains(t, err.Error(), "already exists")
}

func TestE2E_SQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "sqlite3")
}
