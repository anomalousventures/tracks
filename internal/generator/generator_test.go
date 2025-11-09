package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProjectGenerator(t *testing.T) {
	gen := NewProjectGenerator()
	assert.NotNil(t, gen)
}

func TestProjectGenerator_Generate_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		DatabaseDriver: "go-libsql",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, "testapp")

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
}

func TestProjectGenerator_Generate_WithGit(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		DatabaseDriver: "postgres",
		EnvPrefix:      "MYAPP",
		InitGit:        true,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, "testapp")
	gitDir := filepath.Join(projectRoot, ".git")
	_, err = os.Stat(gitDir)
	assert.NoError(t, err, ".git directory should exist")
}

func TestProjectGenerator_Generate_InvalidConfig(t *testing.T) {
	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, "not a ProjectConfig")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

func TestProjectGenerator_Generate_TemplateData(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := ProjectConfig{
		ProjectName:    "myapp",
		ModulePath:     "github.com/user/myapp",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "MYAPP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, "myapp")
	goModPath := filepath.Join(projectRoot, "go.mod")

	content, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "github.com/user/myapp")
	assert.Contains(t, string(content), "go 1.25")
}

func TestProjectGenerator_Validate_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		DatabaseDriver: "go-libsql",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()

	err := gen.Validate(cfg)
	assert.NoError(t, err)
}

func TestProjectGenerator_Validate_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "testapp")
	err := os.Mkdir(projectDir, 0755)
	require.NoError(t, err)

	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		DatabaseDriver: "go-libsql",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()

	err = gen.Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestProjectGenerator_Validate_InvalidConfig(t *testing.T) {
	gen := NewProjectGenerator()

	err := gen.Validate("not a ProjectConfig")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

func TestProjectGenerator_Generate_AllDatabaseDrivers(t *testing.T) {
	drivers := map[string]string{
		"go-libsql": "libsql",
		"sqlite3":   "sqlite3",
		"postgres":  "postgres",
	}

	for driver, expectedInFile := range drivers {
		t.Run(driver, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := ProjectConfig{
				ProjectName:    "testapp",
				ModulePath:     "github.com/test/testapp",
				DatabaseDriver: driver,
				EnvPrefix:      "APP",
				InitGit:        false,
				OutputPath:     tmpDir,
			}

			gen := NewProjectGenerator()
			ctx := context.Background()

			err := gen.Generate(ctx, cfg)
			require.NoError(t, err)

			projectRoot := filepath.Join(tmpDir, "testapp")
			dbPath := filepath.Join(projectRoot, "internal/db/db.go")

			content, err := os.ReadFile(dbPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), expectedInFile)
		})
	}
}
