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

	t.Run("CI env empty string returns console mode", func(t *testing.T) {
		t.Setenv("CI", "")

		cfg := UIConfig{Mode: ModeAuto}
		got := DetectMode(cfg)
		if got != ModeConsole {
			t.Errorf("DetectMode() with CI='' = %v, want ModeConsole", got)
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
