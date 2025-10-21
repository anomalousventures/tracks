package ui

import (
	"os"

	"github.com/mattn/go-isatty"
)

// UIMode represents the output mode for the CLI.
// The mode determines how output is rendered to the user.
type UIMode int

const (
	// ModeAuto detects the appropriate mode based on environment
	// (TTY detection, CI environment, flags).
	ModeAuto UIMode = iota

	// ModeConsole renders output using Lip Gloss styled text
	// for human-friendly console output.
	ModeConsole

	// ModeJSON renders output as structured JSON for scripting
	// and automation.
	ModeJSON

	// ModeTUI launches interactive Terminal UI using Bubble Tea
	// (coming in Phase 4).
	ModeTUI
)

// String returns the string representation of UIMode.
func (m UIMode) String() string {
	switch m {
	case ModeAuto:
		return "auto"
	case ModeConsole:
		return "console"
	case ModeJSON:
		return "json"
	case ModeTUI:
		return "tui"
	default:
		return "unknown"
	}
}

// UIConfig holds configuration for UI rendering.
type UIConfig struct {
	// Mode specifies the output mode.
	Mode UIMode

	// NoColor disables color output (respects NO_COLOR env var).
	NoColor bool

	// Interactive forces interactive mode even in non-TTY environments.
	Interactive bool
}

// DetectMode determines the appropriate UI mode based on configuration and environment.
// Detection logic:
//   - If cfg.Mode is not ModeAuto, returns the explicitly set mode
//   - If CI environment variable is set, returns ModeConsole
//   - If stdout is not a TTY (piped/redirected), returns ModeConsole
//   - Otherwise returns ModeConsole (TUI mode deferred to Phase 4)
func DetectMode(cfg UIConfig) UIMode {
	// Respect explicit mode setting
	if cfg.Mode != ModeAuto {
		return cfg.Mode
	}

	// CI environment or non-TTY output uses console mode
	_, isCi := os.LookupEnv("CI")
	if isCi || !isatty.IsTerminal(os.Stdout.Fd()) {
		return ModeConsole
	}

	// Default to console mode (TUI coming in Phase 4)
	return ModeConsole
}
