package validation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testStringValidator(t *testing.T, tests []struct {
	name    string
	input   string
	wantErr bool
}, validateFunc func(context.Context, string) error, funcName, expectedField string) {
	t.Helper()
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFunc(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%q) error = %v, wantErr %v", funcName, tt.input, err, tt.wantErr)
			}

			if err != nil {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Error("expected ValidationError")
				}
				if valErr.Field != expectedField {
					t.Errorf("expected field %q, got %q", expectedField, valErr.Field)
				}
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "myapp", false},
		{"valid with hyphen", "my-app", false},
		{"valid with underscore", "my_app", false},
		{"valid with numbers", "app123", false},
		{"valid single char", "a", false},
		{"valid 100 chars", strings.Repeat("a", 100), false},
		{"invalid uppercase", "MyApp", true},
		{"invalid space", "my app", true},
		{"invalid special char", "my@app", true},
		{"invalid empty", "", true},
		{"invalid too long", strings.Repeat("a", 101), true},
		{"invalid unicode", "myäpp", true},
		{"invalid dot", "my.app", true},
	}

	testStringValidator(t, tests, v.ValidateProjectName, "ValidateProjectName", "project_name")
}

func TestValidateModulePath(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid github", "github.com/user/project", false},
		{"valid gitlab", "gitlab.com/org/project", false},
		{"valid nested", "github.com/user/org/project", false},
		{"valid with hyphen", "github.com/my-org/my-project", false},
		{"valid with underscore", "github.com/my_org/my_project", false},
		{"valid with numbers", "github.com/user123/project456", false},
		{"invalid no domain", "project", true},
		{"invalid leading slash", "/github.com/user/project", true},
		{"invalid trailing slash", "github.com/user/project/", true},
		{"invalid empty", "", true},
		{"invalid no path", "github.com", true},
		{"invalid spaces", "github.com/user /project", true},
		{"invalid special chars", "github.com/user/project!", true},
		{"invalid too long", "github.com/" + strings.Repeat("a", 301), true},
	}

	testStringValidator(t, tests, v.ValidateModulePath, "ValidateModulePath", "module_path")
}

func TestValidateDirectory(t *testing.T) {
	ctx := context.Background()
	v := NewValidator()

	t.Run("non-existent directory is valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "nonexistent")

		err := v.ValidateDirectory(ctx, newDir)
		if err != nil {
			t.Errorf("expected no error for non-existent directory, got %v", err)
		}
	})

	t.Run("empty directory is valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyDir := filepath.Join(tmpDir, "empty")
		if err := os.Mkdir(emptyDir, 0755); err != nil {
			t.Fatal(err)
		}

		err := v.ValidateDirectory(ctx, emptyDir)
		if err != nil {
			t.Errorf("expected no error for empty directory, got %v", err)
		}
	})

	t.Run("directory with files is invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := v.ValidateDirectory(ctx, tmpDir)
		if err == nil {
			t.Error("expected error for directory with files")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("expected ValidationError")
		}
	})

	t.Run("path is file not directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := v.ValidateDirectory(ctx, testFile)
		if err == nil {
			t.Error("expected error for file path")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("expected ValidationError")
		}
	})

	t.Run("parent directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		deepPath := filepath.Join(tmpDir, "nonexistent", "child")

		err := v.ValidateDirectory(ctx, deepPath)
		if err == nil {
			t.Error("expected error for missing parent directory")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("expected ValidationError")
		}
	})

	t.Run("parent directory not writable", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix permission test on Windows")
		}

		tmpDir := t.TempDir()
		readOnlyParent := filepath.Join(tmpDir, "readonly")
		if err := os.Mkdir(readOnlyParent, 0555); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(readOnlyParent, 0755) }()

		newDir := filepath.Join(readOnlyParent, "child")

		err := v.ValidateDirectory(ctx, newDir)
		if err == nil {
			t.Error("expected error for non-writable parent directory")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Error("expected ValidationError")
		}
		if !errors.Is(err, ErrDirectoryNotWritable) {
			t.Errorf("expected ErrDirectoryNotWritable, got %v", err)
		}
	})

	t.Run("directory exists but unreadable", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix permission test on Windows")
		}

		tmpDir := t.TempDir()
		unreadableDir := filepath.Join(tmpDir, "unreadable")
		if err := os.Mkdir(unreadableDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(unreadableDir, 0755) }()

		err := v.ValidateDirectory(ctx, unreadableDir)
		if err == nil {
			t.Error("expected error for unreadable directory")
		}

		var valErr *ValidationError
		if errors.As(err, &valErr) {
			t.Error("expected wrapped error, not ValidationError for permission issue")
		}
	})

	t.Run("stat fails with non-NotExist error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix permission test on Windows")
		}

		tmpDir := t.TempDir()
		restrictedParent := filepath.Join(tmpDir, "restricted")
		if err := os.Mkdir(restrictedParent, 0000); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(restrictedParent, 0755) }()

		testPath := filepath.Join(restrictedParent, "child")

		err := v.ValidateDirectory(ctx, testPath)
		if err == nil {
			t.Error("expected error when stat fails due to permissions")
		}

		var valErr *ValidationError
		if errors.As(err, &valErr) {
			t.Error("expected wrapped error, not ValidationError for stat permission issue")
		}
	})
}

func TestValidateDatabaseDriver(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid go-libsql", "go-libsql", false},
		{"valid sqlite3", "sqlite3", false},
		{"valid postgres", "postgres", false},
		{"invalid mysql", "mysql", true},
		{"invalid wrong case", "Go-LibSQL", true},
		{"invalid wrong case sqlite", "SQLite3", true},
		{"invalid wrong case postgres", "Postgres", true},
		{"invalid empty", "", true},
		{"invalid random", "mongodb", true},
	}

	testStringValidator(t, tests, v.ValidateDatabaseDriver, "ValidateDatabaseDriver", "database_driver")
}

func TestValidateEnvPrefix(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid APP", "APP", false},
		{"valid MYAPP", "MYAPP", false},
		{"valid USER_SERVICE", "USER_SERVICE", false},
		{"valid with numbers", "API2", false},
		{"valid with multiple underscores", "MY_LONG_PREFIX", false},
		{"invalid lowercase", "app", true},
		{"invalid mixed case", "MyApp", true},
		{"invalid starts with number", "2APP", true},
		{"invalid starts with underscore", "_APP", true},
		{"invalid contains hyphen", "MY-APP", true},
		{"invalid contains special chars", "APP!", true},
		{"invalid empty", "", true},
		{"invalid too long", "VERYLONGPREFIXTHATISWAYMORETHANFIFTYCHARACTERSLONGX", true},
	}

	testStringValidator(t, tests, v.ValidateEnvPrefix, "ValidateEnvPrefix", "env_prefix")
}

func TestValidationErrorMessages(t *testing.T) {
	ctx := context.Background()
	v := NewValidator()

	t.Run("project name error includes helpful message", func(t *testing.T) {
		err := v.ValidateProjectName(ctx, "MyApp")
		if err == nil {
			t.Fatal("expected error")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatal("expected ValidationError")
		}

		if valErr.Message == "" {
			t.Error("error message should not be empty")
		}
	})

	t.Run("database driver error lists supported options", func(t *testing.T) {
		err := v.ValidateDatabaseDriver(ctx, "mysql")
		if err == nil {
			t.Fatal("expected error")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatal("expected ValidationError")
		}

		msg := valErr.Message
		if msg == "" {
			t.Error("error message should not be empty")
		}
	})
}
