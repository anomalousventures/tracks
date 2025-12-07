package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestGetMigrationsDir(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		driver     string
		want       string
	}{
		{
			name:       "postgres driver",
			projectDir: "/home/user/myproject",
			driver:     "postgres",
			want:       filepath.Join("/home/user/myproject", "internal", "db", "migrations", "postgres"),
		},
		{
			name:       "sqlite3 driver",
			projectDir: "/home/user/myproject",
			driver:     "sqlite3",
			want:       filepath.Join("/home/user/myproject", "internal", "db", "migrations", "sqlite3"),
		},
		{
			name:       "go-libsql driver",
			projectDir: "/tmp/testapp",
			driver:     "go-libsql",
			want:       filepath.Join("/tmp/testapp", "internal", "db", "migrations", "go-libsql"),
		},
		{
			name:       "relative project dir",
			projectDir: "./myproject",
			driver:     "postgres",
			want:       filepath.Join("./myproject", "internal", "db", "migrations", "postgres"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMigrationsDir(tt.projectDir, tt.driver)
			if got != tt.want {
				t.Errorf("GetMigrationsDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGooseDialect(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "postgres driver",
			driver:  "postgres",
			wantErr: false,
		},
		{
			name:    "sqlite3 driver",
			driver:  "sqlite3",
			wantErr: false,
		},
		{
			name:    "go-libsql driver",
			driver:  "go-libsql",
			wantErr: false,
		},
		{
			name:      "unknown driver",
			driver:    "mysql",
			wantErr:   true,
			errSubstr: "unsupported database driver",
		},
		{
			name:      "empty driver",
			driver:    "",
			wantErr:   true,
			errSubstr: "unsupported database driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialect, err := gooseDialect(tt.driver)
			if (err != nil) != tt.wantErr {
				t.Errorf("gooseDialect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("gooseDialect() error = %v, want error containing %q", err, tt.errSubstr)
				}
			}
			if !tt.wantErr && dialect == "" {
				t.Error("gooseDialect() returned empty dialect for valid driver")
			}
		})
	}
}

func TestNewMigrationRunner_MissingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "nonexistent")

	_, err := NewMigrationRunner(nil, "postgres", migrationsDir)
	if err == nil {
		t.Fatal("expected error for missing migrations directory")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestNewMigrationRunner_NotADirectory(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "notadir")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := NewMigrationRunner(nil, "postgres", filePath)
	if err == nil {
		t.Fatal("expected error for path that is not a directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", err)
	}
}

func TestNewMigrationRunner_UnsupportedDriver(t *testing.T) {
	tempDir := t.TempDir()

	_, err := NewMigrationRunner(nil, "mysql", tempDir)
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("expected 'unsupported database driver' error, got: %v", err)
	}
}

func TestGooseResultsToStatus_NilResults(t *testing.T) {
	result := gooseResultsToStatus(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d items", len(result))
	}
}

func TestGooseResultsToStatus_EmptyResults(t *testing.T) {
	result := gooseResultsToStatus(nil)
	if result == nil {
		t.Error("expected non-nil slice for nil input")
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d items", len(result))
	}
}

func TestGooseResultsToStatus_WithNilElementsInSlice(t *testing.T) {
	results := make([]*goose.MigrationResult, 3)
	results[0] = nil
	results[1] = nil
	results[2] = nil

	statuses := gooseResultsToStatus(results)
	if len(statuses) != 0 {
		t.Errorf("expected empty result for slice of nils, got %d items", len(statuses))
	}
}

func TestGooseResultsToStatus_WithNilSource(t *testing.T) {
	results := []*goose.MigrationResult{
		{Source: nil},
	}

	statuses := gooseResultsToStatus(results)
	if len(statuses) != 0 {
		t.Errorf("expected empty result when Source is nil, got %d items", len(statuses))
	}
}

func TestGooseResultsToStatus_WithValidResults(t *testing.T) {
	results := []*goose.MigrationResult{
		{
			Source: &goose.Source{
				Version: 1,
				Path:    "/migrations/001_init.sql",
			},
		},
		{
			Source: &goose.Source{
				Version: 2,
				Path:    "/migrations/002_users.sql",
			},
		},
	}

	statuses := gooseResultsToStatus(results)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 results, got %d", len(statuses))
	}

	if statuses[0].Version != 1 {
		t.Errorf("expected version 1, got %d", statuses[0].Version)
	}
	if statuses[0].Name != "001_init.sql" {
		t.Errorf("expected name '001_init.sql', got %q", statuses[0].Name)
	}
	if !statuses[0].Applied {
		t.Error("expected Applied to be true")
	}

	if statuses[1].Version != 2 {
		t.Errorf("expected version 2, got %d", statuses[1].Version)
	}
	if statuses[1].Name != "002_users.sql" {
		t.Errorf("expected name '002_users.sql', got %q", statuses[1].Name)
	}
}

func TestGooseResultsToStatus_MixedNilAndValid(t *testing.T) {
	results := []*goose.MigrationResult{
		nil,
		{
			Source: &goose.Source{
				Version: 1,
				Path:    "/migrations/001_init.sql",
			},
		},
		{Source: nil},
		{
			Source: &goose.Source{
				Version: 2,
				Path:    "/migrations/002_users.sql",
			},
		},
		nil,
	}

	statuses := gooseResultsToStatus(results)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 valid results, got %d", len(statuses))
	}

	if statuses[0].Version != 1 {
		t.Errorf("expected first valid version 1, got %d", statuses[0].Version)
	}
	if statuses[1].Version != 2 {
		t.Errorf("expected second valid version 2, got %d", statuses[1].Version)
	}
}
