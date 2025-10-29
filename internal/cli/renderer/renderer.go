// Package renderer provides implementations of CLI output formatting.
//
// This package provides concrete implementations of the interfaces.Renderer
// interface per ADR-002 (Interface Placement in Consumer Packages).
//
// The Renderer pattern separates business logic from output formatting,
// enabling multiple output modes (console, JSON, TUI) without duplicating code.
// Commands produce data, Renderers display it in the appropriate format.
//
// Available implementations:
//   - ConsoleRenderer: Human-friendly output with colors and formatting
//   - JSONRenderer: Machine-readable JSON output for scripting
//   - TUIRenderer: Interactive terminal UI (future implementation)
package renderer
