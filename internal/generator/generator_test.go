package generator

import (
	"context"
	"encoding/base64"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anomalousventures/tracks/internal/generator/template"
	"github.com/anomalousventures/tracks/internal/templates"
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
		"internal/http/routes/health.go",
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

func TestGenerateSecretKey(t *testing.T) {
	key, err := generateSecretKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
}

func TestGenerateSecretKey_Length(t *testing.T) {
	key, err := generateSecretKey()
	require.NoError(t, err)
	assert.Len(t, key, 44)
}

func TestGenerateSecretKey_ValidBase64(t *testing.T) {
	key, err := generateSecretKey()
	require.NoError(t, err)

	decoded, err := base64.StdEncoding.DecodeString(key)
	assert.NoError(t, err, "key should be valid base64")
	assert.Len(t, decoded, 32, "decoded key should be 32 bytes")
}

func TestGenerateSecretKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		key, err := generateSecretKey()
		require.NoError(t, err)

		assert.False(t, keys[key], "generated key should be unique")
		keys[key] = true
	}

	assert.Len(t, keys, iterations, "all keys should be unique")
}

func TestUsersRouteTemplate_Renders(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "users.go")

	renderer := template.NewRenderer(templates.FS)

	data := template.TemplateData{
		ModuleName:  "github.com/test/testapp",
		ProjectName: "testapp",
		DBDriver:    "sqlite3",
		GoVersion:   "1.25",
		Year:        time.Now().Year(),
		EnvPrefix:   "APP",
		SecretKey:   "test-secret-key",
	}

	err := renderer.RenderToFile("examples/routes/users.go.tmpl", data, outputPath)
	require.NoError(t, err, "template should render without errors")

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err, "rendered file should exist")

	contentStr := string(content)

	assert.Contains(t, contentStr, "package routes")
	assert.Contains(t, contentStr, "usersPath     = \"users\"")
	assert.Contains(t, contentStr, "UserSlugParam = \"username\"")

	assert.Contains(t, contentStr, "UserIndex")
	assert.Contains(t, contentStr, "UserShow")
	assert.Contains(t, contentStr, "UserNew")
	assert.Contains(t, contentStr, "UserCreate")
	assert.Contains(t, contentStr, "UserEdit")
	assert.Contains(t, contentStr, "UserUpdate")
	assert.Contains(t, contentStr, "UserDelete")

	// Helper functions are in the template, but RouteURL is now in routes.go (shared)
	assert.Contains(t, contentStr, "func UserIndexURL() string")
	assert.Contains(t, contentStr, "func UserShowURL(username string) string")
	assert.Contains(t, contentStr, "func UserNewURL() string")
	assert.Contains(t, contentStr, "func UserCreateURL() string")
	assert.Contains(t, contentStr, "func UserEditURL(username string) string")
	assert.Contains(t, contentStr, "func UserUpdateURL(username string) string")
	assert.Contains(t, contentStr, "func UserDeleteURL(username string) string")

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, outputPath, content, parser.AllErrors)
	require.NoError(t, err, "generated code should be valid Go and compile without errors")
}

func TestUsersRouteTemplate_URLEncoding(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "users.go")

	renderer := template.NewRenderer(templates.FS)

	data := template.TemplateData{
		ModuleName:  "github.com/test/testapp",
		ProjectName: "testapp",
		DBDriver:    "sqlite3",
		GoVersion:   "1.25",
		Year:        time.Now().Year(),
		EnvPrefix:   "APP",
		SecretKey:   "test-secret-key",
	}

	err := renderer.RenderToFile("examples/routes/users.go.tmpl", data, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, outputPath, content, 0)
	require.NoError(t, err)

	tmpTestFile := filepath.Join(tmpDir, "users_test.go")
	testCode := `package routes

import (
	"testing"
)

func TestUserShowURL_SpecialCharacters(t *testing.T) {
	tests := []struct {
		username string
		expected string
	}{
		{"alice+bob@example", "/users/alice%2Bbob%40example"},
		{"user with spaces", "/users/user%20with%20spaces"},
		{"user/slash", "/users/user%2Fslash"},
		{"user?query", "/users/user%3Fquery"},
		{"user&ampersand", "/users/user%26ampersand"},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			result := UserShowURL(tt.username)
			if result != tt.expected {
				t.Errorf("UserShowURL(%q) = %q, want %q", tt.username, result, tt.expected)
			}
		})
	}
}

func TestUserEditURL_SpecialCharacters(t *testing.T) {
	result := UserEditURL("alice+bob@example")
	expected := "/users/alice%2Bbob%40example/edit"
	if result != expected {
		t.Errorf("UserEditURL(%q) = %q, want %q", "alice+bob@example", result, expected)
	}
}
`
	err = os.WriteFile(tmpTestFile, []byte(testCode), 0644)
	require.NoError(t, err)

	_, err = parser.ParseFile(fset, tmpTestFile, nil, 0)
	require.NoError(t, err, "test code should be valid Go")

	assert.NotNil(t, f, "parsed file should not be nil")
}

func TestProjectGenerator_ExampleTemplatesNotGenerated(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		DatabaseDriver: "sqlite3",
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := NewProjectGenerator()
	ctx := context.Background()

	err := gen.Generate(ctx, cfg)
	require.NoError(t, err)

	projectRoot := filepath.Join(tmpDir, "testapp")

	// Verify health.go IS generated
	healthPath := filepath.Join(projectRoot, "internal/http/routes/health.go")
	_, err = os.Stat(healthPath)
	assert.NoError(t, err, "health.go should be generated")

	// Verify users.go is NOT generated (it's an example template)
	usersPath := filepath.Join(projectRoot, "internal/http/routes/users.go")
	_, err = os.Stat(usersPath)
	assert.Error(t, err, "users.go should NOT be generated (example template only)")
	assert.True(t, os.IsNotExist(err), "users.go should not exist")

	// Verify users_test.go is NOT generated (it's an example template)
	usersTestPath := filepath.Join(projectRoot, "internal/http/routes/users_test.go")
	_, err = os.Stat(usersTestPath)
	assert.Error(t, err, "users_test.go should NOT be generated (example template only)")
	assert.True(t, os.IsNotExist(err), "users_test.go should not exist")
}
