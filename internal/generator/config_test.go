package generator

import "testing"

func TestProjectConfig_Fields(t *testing.T) {
	cfg := ProjectConfig{
		ProjectName:    "myapp",
		ModulePath:     "github.com/user/myapp",
		DatabaseDriver: "go-libsql",
		InitGit:        true,
		OutputPath:     "/tmp/myapp",
	}

	tests := []struct {
		name  string
		got   string
		want  string
		field string
	}{
		{"project name", cfg.ProjectName, "myapp", "ProjectName"},
		{"module path", cfg.ModulePath, "github.com/user/myapp", "ModulePath"},
		{"database driver", cfg.DatabaseDriver, "go-libsql", "DatabaseDriver"},
		{"output path", cfg.OutputPath, "/tmp/myapp", "OutputPath"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.field, tt.got, tt.want)
			}
		})
	}

	if !cfg.InitGit {
		t.Errorf("InitGit = %v, want true", cfg.InitGit)
	}
}

func TestProjectConfig_ZeroValue(t *testing.T) {
	var cfg ProjectConfig

	if cfg.ProjectName != "" {
		t.Errorf("zero value ProjectName = %q, want empty string", cfg.ProjectName)
	}

	if cfg.InitGit {
		t.Error("zero value InitGit = true, want false")
	}
}
