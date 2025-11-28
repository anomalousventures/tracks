package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	if detector == nil {
		t.Fatal("NewDetector returned nil")
	}
}

func TestDetector_Detect_NotTracksProject(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()

	proj, dir, err := detector.Detect(ctx, tmpDir)
	if !errors.Is(err, ErrNotTracksProject) {
		t.Errorf("expected ErrNotTracksProject, got %v", err)
	}
	if proj != nil {
		t.Error("expected nil project")
	}
	if dir != "" {
		t.Errorf("expected empty dir, got %q", dir)
	}
}

func TestDetector_Detect_CurrentDirectory(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".tracks.yaml")
	configContent := `schema_version: "1.0"
project:
  name: "testapp"
  module_path: "example.com/testapp"
  tracks_version: "dev"
  last_upgraded_version: "dev"
  database_driver: "go-libsql"
  env_prefix: "APP"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	proj, dir, err := detector.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if proj.Name != "testapp" {
		t.Errorf("expected Name 'testapp', got %q", proj.Name)
	}
	if proj.ModulePath != "example.com/testapp" {
		t.Errorf("expected ModulePath 'example.com/testapp', got %q", proj.ModulePath)
	}
	if proj.DBDriver != "go-libsql" {
		t.Errorf("expected DBDriver 'go-libsql', got %q", proj.DBDriver)
	}

	absDir, _ := filepath.Abs(tmpDir)
	if dir != absDir {
		t.Errorf("expected dir %q, got %q", absDir, dir)
	}
}

func TestDetector_Detect_ParentDirectory(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".tracks.yaml")
	configContent := `schema_version: "1.0"
project:
  name: "parentapp"
  module_path: "example.com/parentapp"
  tracks_version: "dev"
  last_upgraded_version: "dev"
  database_driver: "postgres"
  env_prefix: "APP"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	childDir := filepath.Join(tmpDir, "internal", "http")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	proj, dir, err := detector.Detect(ctx, childDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if proj.Name != "parentapp" {
		t.Errorf("expected Name 'parentapp', got %q", proj.Name)
	}
	if proj.DBDriver != "postgres" {
		t.Errorf("expected DBDriver 'postgres', got %q", proj.DBDriver)
	}

	absDir, _ := filepath.Abs(tmpDir)
	if dir != absDir {
		t.Errorf("expected dir %q, got %q", absDir, dir)
	}
}

func TestDetector_Detect_InvalidYAML(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".tracks.yaml")
	invalidContent := `this is not valid yaml: [[[`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	proj, dir, err := detector.Detect(ctx, tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if proj != nil {
		t.Error("expected nil project")
	}
	if dir != "" {
		t.Errorf("expected empty dir, got %q", dir)
	}
}

func TestDetector_HasTemplUIConfig(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	t.Run("no config", func(t *testing.T) {
		tmpDir := t.TempDir()
		if detector.HasTemplUIConfig(ctx, tmpDir) {
			t.Error("expected false when .templui.json does not exist")
		}
	})

	t.Run("config exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".templui.json")
		if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}
		if !detector.HasTemplUIConfig(ctx, tmpDir) {
			t.Error("expected true when .templui.json exists")
		}
	})
}

func TestLoadConfig_AllFields(t *testing.T) {
	d := &detector{}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".tracks.yaml")
	configContent := `schema_version: "1.0"
project:
  name: "myapp"
  module_path: "github.com/user/myapp"
  tracks_version: "v1.0.0"
  last_upgraded_version: "v1.0.0"
  database_driver: "sqlite3"
  env_prefix: "MYAPP"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	proj, err := d.loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if proj.Name != "myapp" {
		t.Errorf("expected Name 'myapp', got %q", proj.Name)
	}
	if proj.ModulePath != "github.com/user/myapp" {
		t.Errorf("expected ModulePath 'github.com/user/myapp', got %q", proj.ModulePath)
	}
	if proj.DBDriver != "sqlite3" {
		t.Errorf("expected DBDriver 'sqlite3', got %q", proj.DBDriver)
	}
}
