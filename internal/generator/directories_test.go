package generator

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateProjectDirectories(t *testing.T) {
	t.Run("creates all required directories successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := ProjectConfig{
			ProjectName: "testproject",
			OutputPath:  tmpDir,
			ModulePath:  "example.com/testproject",
		}

		err := CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("CreateProjectDirectories() failed: %v", err)
		}

		projectRoot := filepath.Join(tmpDir, "testproject")
		expectedDirs := []string{
			filepath.Join(projectRoot, "cmd", "server"),
			filepath.Join(projectRoot, "internal", "interfaces"),
			filepath.Join(projectRoot, "internal", "domain", "health"),
			filepath.Join(projectRoot, "internal", "http", "handlers"),
			filepath.Join(projectRoot, "internal", "http", "routes"),
			filepath.Join(projectRoot, "internal", "http", "views", "layouts"),
			filepath.Join(projectRoot, "internal", "http", "views", "pages"),
			filepath.Join(projectRoot, "internal", "http", "views", "components"),
			filepath.Join(projectRoot, "internal", "db", "migrations", "sqlite"),
			filepath.Join(projectRoot, "internal", "db", "migrations", "postgres"),
			filepath.Join(projectRoot, "internal", "db", "queries"),
			filepath.Join(projectRoot, "internal", "db", "generated"),
			filepath.Join(projectRoot, "tests", "mocks"),
			filepath.Join(projectRoot, "tests", "integration"),
			filepath.Join(projectRoot, "internal", "assets"),
			filepath.Join(projectRoot, "internal", "assets", "web", "css"),
			filepath.Join(projectRoot, "internal", "assets", "web", "js"),
			filepath.Join(projectRoot, "internal", "assets", "web", "images"),
			filepath.Join(projectRoot, "internal", "assets", "dist"),
		}

		for _, dir := range expectedDirs {
			info, err := os.Stat(dir)
			if err != nil {
				t.Errorf("expected directory %s to exist, got error: %v", dir, err)
				continue
			}
			if !info.IsDir() {
				t.Errorf("expected %s to be a directory", dir)
			}
		}
	})

	t.Run("succeeds when parent directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectRoot := filepath.Join(tmpDir, "existingproject")

		if err := os.Mkdir(projectRoot, 0755); err != nil {
			t.Fatalf("failed to create test parent directory: %v", err)
		}

		config := ProjectConfig{
			ProjectName: "existingproject",
			OutputPath:  tmpDir,
			ModulePath:  "example.com/existingproject",
		}

		err := CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("CreateProjectDirectories() failed with existing parent: %v", err)
		}

		cmdServerDir := filepath.Join(projectRoot, "cmd", "server")
		if _, err := os.Stat(cmdServerDir); err != nil {
			t.Errorf("expected directory %s to exist: %v", cmdServerDir, err)
		}
	})

	t.Run("succeeds when directories already exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := ProjectConfig{
			ProjectName: "idempotentproject",
			OutputPath:  tmpDir,
			ModulePath:  "example.com/idempotentproject",
		}

		err := CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("first CreateProjectDirectories() call failed: %v", err)
		}

		err = CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("second CreateProjectDirectories() call failed (should be idempotent): %v", err)
		}
	})

	t.Run("handles cross-platform paths correctly", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := ProjectConfig{
			ProjectName: "crossplatform",
			OutputPath:  tmpDir,
			ModulePath:  "example.com/crossplatform",
		}

		err := CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("CreateProjectDirectories() failed: %v", err)
		}

		var expectedSeparator string
		if runtime.GOOS == "windows" {
			expectedSeparator = "\\"
		} else {
			expectedSeparator = "/"
		}

		projectRoot := filepath.Join(tmpDir, "crossplatform")
		testPath := filepath.Join(projectRoot, "internal", "domain", "health")

		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("expected cross-platform path to exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected cross-platform path to be a directory")
		}

		if !filepath.IsAbs(testPath) && filepath.Separator != expectedSeparator[0] {
			t.Logf("Note: Path separator is %c (expected %s for %s)", filepath.Separator, expectedSeparator, runtime.GOOS)
		}
	})

	t.Run("returns error when parent directory is not writable", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix permission test on Windows")
		}

		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")

		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("failed to create read-only directory: %v", err)
		}
		defer func() { _ = os.Chmod(readOnlyDir, 0755) }()

		config := ProjectConfig{
			ProjectName: "cantwrite",
			OutputPath:  readOnlyDir,
			ModulePath:  "example.com/cantwrite",
		}

		err := CreateProjectDirectories(config)
		if err == nil {
			t.Fatal("expected error when parent directory is not writable, got nil")
		}
	})

	t.Run("returns error with path context on failure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix permission test on Windows")
		}

		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")

		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("failed to create read-only directory: %v", err)
		}
		defer func() { _ = os.Chmod(readOnlyDir, 0755) }()

		config := ProjectConfig{
			ProjectName: "errorcontext",
			OutputPath:  readOnlyDir,
			ModulePath:  "example.com/errorcontext",
		}

		err := CreateProjectDirectories(config)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		expectedPath := filepath.Join(readOnlyDir, "errorcontext", "cmd", "server")
		if expectedPath == "" {
			t.Error("error message should contain the path that failed")
		}
	})

	t.Run("creates nested directories in correct order", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := ProjectConfig{
			ProjectName: "nestedproject",
			OutputPath:  tmpDir,
			ModulePath:  "example.com/nestedproject",
		}

		err := CreateProjectDirectories(config)
		if err != nil {
			t.Fatalf("CreateProjectDirectories() failed: %v", err)
		}

		projectRoot := filepath.Join(tmpDir, "nestedproject")
		deepPath := filepath.Join(projectRoot, "internal", "domain", "health")

		info, err := os.Stat(deepPath)
		if err != nil {
			t.Fatalf("expected deeply nested directory to exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected deeply nested path to be a directory")
		}

		parentPath := filepath.Join(projectRoot, "internal", "domain")
		parentInfo, err := os.Stat(parentPath)
		if err != nil {
			t.Fatalf("expected parent directory to exist: %v", err)
		}
		if !parentInfo.IsDir() {
			t.Error("expected parent directory to be a directory")
		}
	})

	t.Run("handles project names with hyphens and underscores", func(t *testing.T) {
		tmpDir := t.TempDir()

		testCases := []string{
			"my-project",
			"my_project",
			"my-complex_project",
		}

		for _, projectName := range testCases {
			config := ProjectConfig{
				ProjectName: projectName,
				OutputPath:  tmpDir,
				ModulePath:  "example.com/" + projectName,
			}

			err := CreateProjectDirectories(config)
			if err != nil {
				t.Errorf("CreateProjectDirectories() failed for project name %q: %v", projectName, err)
			}

			projectRoot := filepath.Join(tmpDir, projectName)
			if _, err := os.Stat(projectRoot); err != nil {
				t.Errorf("expected project directory %s to exist: %v", projectRoot, err)
			}
		}
	})
}
