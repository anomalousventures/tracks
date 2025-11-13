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
				".dockerignore",
				".golangci.yml",
				".mockery.yaml",
				".tracks.yaml",
				".env.example",
				".env",
				"Makefile",
				"Dockerfile",
				"docker-compose.yml",
				"sqlc.yaml",
				".github/workflows/ci.yml",
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
				".github",
				".github/workflows",
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

			envPath := filepath.Join(projectRoot, ".env")
			envContent, err := os.ReadFile(envPath)
			require.NoError(t, err, "should be able to read .env")

			envDatabaseURLs := map[string]string{
				"go-libsql": "APP_DATABASE_URL=http://localhost:8081",
				"sqlite3":   "APP_DATABASE_URL=file:./data/testapp.db",
				"postgres":  "APP_DATABASE_URL=postgres://postgres:postgres@localhost:5432/testapp?sslmode=disable",
			}
			expectedDBURL := envDatabaseURLs[tt.databaseDriver]
			assert.Contains(t, string(envContent), expectedDBURL, ".env should contain correct database URL for %s", tt.databaseDriver)

			assert.Contains(t, string(envContent), "SECRET_KEY=", ".env should contain SECRET_KEY")
			assert.NotContains(t, string(envContent), "SECRET_KEY=your-secret-key-here", ".env should not contain placeholder secret key")
			assert.Regexp(t, `SECRET_KEY=[A-Za-z0-9+/]{43}=`, string(envContent), ".env should contain valid base64-encoded secret key (44 chars)")

			makefilePath := filepath.Join(projectRoot, "Makefile")
			makefileContent, err := os.ReadFile(makefilePath)
			require.NoError(t, err, "should be able to read Makefile")

			assert.Contains(t, string(makefileContent), "grep -q '^  [a-z]' docker-compose.yml", "Makefile should contain auto-start logic")
			assert.Contains(t, string(makefileContent), "docker-compose up -d", "Makefile should auto-start services")
			assert.NotContains(t, string(makefileContent), "dev-full:", "Makefile should not contain dev-full target")
			assert.NotContains(t, string(makefileContent), "dev-full     - Start docker services and dev server", "Makefile help should not mention dev-full")

			dockerignorePath := filepath.Join(projectRoot, ".dockerignore")
			dockerignoreContent, err := os.ReadFile(dockerignorePath)
			require.NoError(t, err, "should be able to read .dockerignore")
			assert.Contains(t, string(dockerignoreContent), ".git/", ".dockerignore should exclude .git directory")
			assert.Contains(t, string(dockerignoreContent), "*_test.go", ".dockerignore should exclude test files")
			assert.Contains(t, string(dockerignoreContent), ".env", ".dockerignore should exclude .env files")
			assert.Contains(t, string(dockerignoreContent), "!.env.example", ".dockerignore should include .env.example")

			dockerfilePath := filepath.Join(projectRoot, "Dockerfile")
			dockerfileContent, err := os.ReadFile(dockerfilePath)
			require.NoError(t, err, "should be able to read Dockerfile")

			if tt.databaseDriver == "postgres" {
				assert.Contains(t, string(dockerfileContent), "CGO_ENABLED=0", "Dockerfile should disable CGO for postgres")
				assert.NotContains(t, string(dockerfileContent), "gcc musl-dev", "Dockerfile should not install CGO dependencies for postgres")
				assert.Contains(t, string(dockerfileContent), "distroless", "Dockerfile should use distroless for postgres (minimal image)")
				assert.Contains(t, string(dockerfileContent), "nonroot", "Dockerfile should run as non-root for security")
			} else {
				assert.Contains(t, string(dockerfileContent), "CGO_ENABLED=1", "Dockerfile should enable CGO for %s", tt.databaseDriver)
				assert.Contains(t, string(dockerfileContent), "gcc musl-dev", "Dockerfile should install CGO dependencies for %s", tt.databaseDriver)
				assert.Contains(t, string(dockerfileContent), "libc6-compat", "Dockerfile should include runtime CGO dependencies for %s", tt.databaseDriver)
				assert.Contains(t, string(dockerfileContent), "adduser", "Dockerfile should create non-root user for %s", tt.databaseDriver)
				assert.Contains(t, string(dockerfileContent), "USER appuser", "Dockerfile should run as non-root for security")
			}

			assert.Contains(t, string(dockerfileContent), "-ldflags=\"-w -s\"", "Dockerfile should use build optimizations to reduce binary size")
			assert.Contains(t, string(dockerfileContent), "-trimpath", "Dockerfile should use -trimpath to remove filesystem paths")
			assert.NotContains(t, string(dockerfileContent), "migrations", "Dockerfile should not copy migrations (not implemented yet)")

			readmePath := filepath.Join(projectRoot, "README.md")
			readmeContent, err := os.ReadFile(readmePath)
			require.NoError(t, err, "should be able to read README.md")
			assert.Contains(t, string(readmeContent), "## Docker", "README should contain Docker section")
			assert.Contains(t, string(readmeContent), "docker build", "README should contain docker build instructions")

			ciWorkflowPath := filepath.Join(projectRoot, ".github/workflows/ci.yml")
			ciWorkflowContent, err := os.ReadFile(ciWorkflowPath)
			require.NoError(t, err, "should be able to read .github/workflows/ci.yml")
			assert.Contains(t, string(ciWorkflowContent), "name: CI", "CI workflow should have name")
			assert.Contains(t, string(ciWorkflowContent), "golangci-lint", "CI workflow should run linter")
			assert.Contains(t, string(ciWorkflowContent), "go test", "CI workflow should run tests")
			assert.Contains(t, string(ciWorkflowContent), "docker build", "CI workflow should build Docker image")
			assert.Contains(t, string(ciWorkflowContent), "trivy", "CI workflow should scan with Trivy")
			assert.Contains(t, string(ciWorkflowContent), "/api/health", "CI workflow should test health endpoint")

			if tt.databaseDriver == "postgres" {
				assert.Contains(t, string(ciWorkflowContent), "postgres:16-alpine", "CI workflow should include postgres service")
				assert.Contains(t, string(ciWorkflowContent), "services:", "CI workflow should have services section for postgres")
			}

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
