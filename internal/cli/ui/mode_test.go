package ui

import "testing"

func TestUIModeConstants(t *testing.T) {
	tests := []struct {
		name  string
		mode  UIMode
		value int
	}{
		{"ModeAuto is 0", ModeAuto, 0},
		{"ModeConsole is 1", ModeConsole, 1},
		{"ModeJSON is 2", ModeJSON, 2},
		{"ModeTUI is 3", ModeTUI, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.mode) != tt.value {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.mode), tt.value)
			}
		})
	}
}

func TestUIModeString(t *testing.T) {
	tests := []struct {
		name string
		mode UIMode
		want string
	}{
		{"ModeAuto string", ModeAuto, "auto"},
		{"ModeConsole string", ModeConsole, "console"},
		{"ModeJSON string", ModeJSON, "json"},
		{"ModeTUI string", ModeTUI, "tui"},
		{"Unknown mode", UIMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("UIMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUIConfigZeroValue(t *testing.T) {
	var cfg UIConfig

	if cfg.Mode != ModeAuto {
		t.Errorf("zero value Mode = %v, want ModeAuto", cfg.Mode)
	}
	if cfg.NoColor {
		t.Error("zero value NoColor should be false")
	}
	if cfg.Interactive {
		t.Error("zero value Interactive should be false")
	}
}

func TestUIConfigFields(t *testing.T) {
	tests := []struct {
		name        string
		config      UIConfig
		wantMode    UIMode
		wantNoColor bool
		wantInteractive bool
	}{
		{
			name:        "default config",
			config:      UIConfig{},
			wantMode:    ModeAuto,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name: "console mode with color",
			config: UIConfig{
				Mode:    ModeConsole,
				NoColor: false,
				Interactive: false,
			},
			wantMode:    ModeConsole,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name: "JSON mode",
			config: UIConfig{
				Mode:    ModeJSON,
				NoColor: true,
				Interactive: false,
			},
			wantMode:    ModeJSON,
			wantNoColor: true,
			wantInteractive: false,
		},
		{
			name: "TUI mode interactive",
			config: UIConfig{
				Mode:    ModeTUI,
				NoColor: false,
				Interactive: true,
			},
			wantMode:    ModeTUI,
			wantNoColor: false,
			wantInteractive: true,
		},
		{
			name: "all fields set",
			config: UIConfig{
				Mode:    ModeConsole,
				NoColor: true,
				Interactive: true,
			},
			wantMode:    ModeConsole,
			wantNoColor: true,
			wantInteractive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Mode != tt.wantMode {
				t.Errorf("Mode = %v, want %v", tt.config.Mode, tt.wantMode)
			}
			if tt.config.NoColor != tt.wantNoColor {
				t.Errorf("NoColor = %v, want %v", tt.config.NoColor, tt.wantNoColor)
			}
			if tt.config.Interactive != tt.wantInteractive {
				t.Errorf("Interactive = %v, want %v", tt.config.Interactive, tt.wantInteractive)
			}
		})
	}
}

func TestUIModeStringRoundTrip(t *testing.T) {
	modes := []UIMode{ModeAuto, ModeConsole, ModeJSON, ModeTUI}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			str := mode.String()
			if str == "" {
				t.Error("String() should not return empty string")
			}
			if str == "unknown" && mode <= ModeTUI {
				t.Errorf("valid mode %v should not return 'unknown'", mode)
			}
		})
	}
}

func TestUIConfigCanBeCompared(t *testing.T) {
	cfg1 := UIConfig{Mode: ModeConsole, NoColor: false, Interactive: false}
	cfg2 := UIConfig{Mode: ModeConsole, NoColor: false, Interactive: false}
	cfg3 := UIConfig{Mode: ModeJSON, NoColor: true, Interactive: false}

	if cfg1 != cfg2 {
		t.Error("identical configs should be equal")
	}
	if cfg1 == cfg3 {
		t.Error("different configs should not be equal")
	}
}

func TestDetectModeExplicitMode(t *testing.T) {
	tests := []struct {
		name string
		mode UIMode
		want UIMode
	}{
		{"explicit console mode", ModeConsole, ModeConsole},
		{"explicit JSON mode", ModeJSON, ModeJSON},
		{"explicit TUI mode", ModeTUI, ModeTUI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := UIConfig{Mode: tt.mode}
			got := DetectMode(cfg)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectModeCIEnvironment(t *testing.T) {
	t.Run("CI env set returns console mode", func(t *testing.T) {
		t.Setenv("CI", "true")

		cfg := UIConfig{Mode: ModeAuto}
		got := DetectMode(cfg)
		if got != ModeConsole {
			t.Errorf("DetectMode() with CI=true = %v, want ModeConsole", got)
		}
	})
}

func TestDetectModeDefault(t *testing.T) {
	cfg := UIConfig{Mode: ModeAuto}
	got := DetectMode(cfg)
	if got != ModeConsole {
		t.Errorf("DetectMode() default = %v, want ModeConsole", got)
	}
}

func TestDetectModeTTYPath(t *testing.T) {
	t.Run("TTY environment returns console mode", func(t *testing.T) {
		cfg := UIConfig{Mode: ModeAuto}
		mockTTY := func(fd uintptr) bool { return true }

		got := detectModeWithTTY(cfg, mockTTY)
		if got != ModeConsole {
			t.Errorf("TTY path = %v, want ModeConsole", got)
		}
	})

	t.Run("non-TTY environment returns console mode", func(t *testing.T) {
		cfg := UIConfig{Mode: ModeAuto}
		mockTTY := func(fd uintptr) bool { return false }

		got := detectModeWithTTY(cfg, mockTTY)
		if got != ModeConsole {
			t.Errorf("non-TTY path = %v, want ModeConsole", got)
		}
	})
}

func TestDetectModeJSON(t *testing.T) {
	tests := []struct {
		name    string
		cfg     UIConfig
		mockTTY func(fd uintptr) bool
		want    UIMode
	}{
		{
			name:    "JSON overrides explicit console mode",
			cfg:     UIConfig{Mode: ModeConsole, JSON: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name:    "JSON overrides explicit TUI mode",
			cfg:     UIConfig{Mode: ModeTUI, JSON: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name:    "JSON overrides TTY detection",
			cfg:     UIConfig{Mode: ModeAuto, JSON: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name:    "JSON works in CI environment",
			cfg:     UIConfig{Mode: ModeAuto, JSON: true},
			mockTTY: func(fd uintptr) bool { return false },
			want:    ModeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "JSON works in CI environment" {
				t.Setenv("CI", "true")
			}

			got := detectModeWithTTY(tt.cfg, tt.mockTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectModeInteractive(t *testing.T) {
	tests := []struct {
		name    string
		cfg     UIConfig
		mockTTY func(fd uintptr) bool
		want    UIMode
	}{
		{
			name:    "Interactive overrides explicit console mode",
			cfg:     UIConfig{Mode: ModeConsole, Interactive: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeTUI,
		},
		{
			name:    "Interactive overrides explicit JSON mode",
			cfg:     UIConfig{Mode: ModeJSON, Interactive: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeTUI,
		},
		{
			name:    "Interactive overrides non-TTY detection",
			cfg:     UIConfig{Mode: ModeAuto, Interactive: true},
			mockTTY: func(fd uintptr) bool { return false },
			want:    ModeTUI,
		},
		{
			name:    "Interactive works in CI environment",
			cfg:     UIConfig{Mode: ModeAuto, Interactive: true},
			mockTTY: func(fd uintptr) bool { return false },
			want:    ModeTUI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Interactive works in CI environment" {
				t.Setenv("CI", "true")
			}

			got := detectModeWithTTY(tt.cfg, tt.mockTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectModeBothSet(t *testing.T) {
	tests := []struct {
		name    string
		cfg     UIConfig
		mockTTY func(fd uintptr) bool
		want    UIMode
	}{
		{
			name: "JSON takes precedence over Interactive",
			cfg: UIConfig{
				Mode:        ModeAuto,
				JSON:        true,
				Interactive: true,
			},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name: "JSON precedence applies with explicit mode",
			cfg: UIConfig{
				Mode:        ModeConsole,
				JSON:        true,
				Interactive: true,
			},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name: "JSON precedence applies in CI environment",
			cfg: UIConfig{
				Mode:        ModeAuto,
				JSON:        true,
				Interactive: true,
			},
			mockTTY: func(fd uintptr) bool { return false },
			want:    ModeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "JSON precedence applies in CI environment" {
				t.Setenv("CI", "true")
			}

			got := detectModeWithTTY(tt.cfg, tt.mockTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectModeNoOverrides(t *testing.T) {
	tests := []struct {
		name    string
		cfg     UIConfig
		mockTTY func(fd uintptr) bool
		want    UIMode
	}{
		{
			name:    "no overrides with explicit console mode",
			cfg:     UIConfig{Mode: ModeConsole, JSON: false, Interactive: false},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeConsole,
		},
		{
			name:    "no overrides with explicit JSON mode",
			cfg:     UIConfig{Mode: ModeJSON, JSON: false, Interactive: false},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name:    "no overrides with explicit TUI mode",
			cfg:     UIConfig{Mode: ModeTUI, JSON: false, Interactive: false},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeTUI,
		},
		{
			name:    "no overrides with auto mode in TTY",
			cfg:     UIConfig{Mode: ModeAuto, JSON: false, Interactive: false},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeConsole,
		},
		{
			name:    "no overrides with auto mode in CI",
			cfg:     UIConfig{Mode: ModeAuto, JSON: false, Interactive: false},
			mockTTY: func(fd uintptr) bool { return false },
			want:    ModeConsole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "no overrides with auto mode in CI" {
				t.Setenv("CI", "true")
			}

			got := detectModeWithTTY(tt.cfg, tt.mockTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectModeNoColor(t *testing.T) {
	tests := []struct {
		name    string
		cfg     UIConfig
		mockTTY func(fd uintptr) bool
		want    UIMode
	}{
		{
			name:    "NoColor set with auto mode returns console",
			cfg:     UIConfig{Mode: ModeAuto, NoColor: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeConsole,
		},
		{
			name:    "NoColor set doesn't override explicit JSON mode",
			cfg:     UIConfig{Mode: ModeJSON, NoColor: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeJSON,
		},
		{
			name:    "NoColor set doesn't override explicit TUI mode",
			cfg:     UIConfig{Mode: ModeTUI, NoColor: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeTUI,
		},
		{
			name:    "NoColor set doesn't override Interactive flag",
			cfg:     UIConfig{Mode: ModeAuto, NoColor: true, Interactive: true},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeTUI,
		},
		{
			name:    "NoColor false with auto mode and TTY returns console",
			cfg:     UIConfig{Mode: ModeAuto, NoColor: false},
			mockTTY: func(fd uintptr) bool { return true },
			want:    ModeConsole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectModeWithTTY(tt.cfg, tt.mockTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUIConfigLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		cfg          UIConfig
		wantLogLevel string
	}{
		{
			name:         "default log level",
			cfg:          UIConfig{},
			wantLogLevel: "",
		},
		{
			name:         "debug log level",
			cfg:          UIConfig{LogLevel: "debug"},
			wantLogLevel: "debug",
		},
		{
			name:         "info log level",
			cfg:          UIConfig{LogLevel: "info"},
			wantLogLevel: "info",
		},
		{
			name:         "warn log level",
			cfg:          UIConfig{LogLevel: "warn"},
			wantLogLevel: "warn",
		},
		{
			name:         "error log level",
			cfg:          UIConfig{LogLevel: "error"},
			wantLogLevel: "error",
		},
		{
			name:         "off log level",
			cfg:          UIConfig{LogLevel: "off"},
			wantLogLevel: "off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel = %q, want %q", tt.cfg.LogLevel, tt.wantLogLevel)
			}
		})
	}
}
