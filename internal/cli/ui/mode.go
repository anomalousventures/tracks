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

	// JSON enables JSON output mode for scripting and automation.
	// Takes highest precedence in mode detection.
	JSON bool

	// Interactive forces interactive TUI mode even in non-TTY environments.
	// Takes precedence over auto-detection but lower than JSON.
	Interactive bool

	// LogLevel controls developer debug logging (debug, info, warn, error, off).
	// Defaults to "off" to hide debugging from end users.
	LogLevel string
}

// ttyDetector checks if a file descriptor is a TTY.
type ttyDetector func(fd uintptr) bool

// defaultTTYDetector uses isatty for production.
var defaultTTYDetector ttyDetector = isatty.IsTerminal

// DetectMode determines the appropriate UI mode based on configuration and environment.
// Detection priority (highest to lowest):
//  1. JSON set → returns ModeJSON (for scripting/automation)
//  2. Interactive set → returns ModeTUI (force interactive)
//  3. cfg.Mode (if not ModeAuto) → returns explicitly set mode
//  4. NO_COLOR, CI environment, or non-TTY → returns ModeConsole
//  5. Default → returns ModeConsole (TUI coming in Phase 4)
func DetectMode(cfg UIConfig) UIMode {
	return detectModeWithTTY(cfg, defaultTTYDetector)
}

// detectModeWithTTY is an internal helper that allows TTY detection to be mocked for testing.
func detectModeWithTTY(cfg UIConfig, isTTY ttyDetector) UIMode {
	// Highest priority: JSON output mode
	if cfg.JSON {
		return ModeJSON
	}
	// Force interactive TUI mode
	if cfg.Interactive {
		return ModeTUI
	}

	// Respect explicit mode setting
	if cfg.Mode != ModeAuto {
		return cfg.Mode
	}

	// NO_COLOR, CI environment, or non-TTY output uses console mode
	if cfg.NoColor || os.Getenv("CI") != "" || !isTTY(os.Stdout.Fd()) {
		return ModeConsole
	}

	// Default to console mode (TUI coming in Phase 4)
	return ModeConsole
}
