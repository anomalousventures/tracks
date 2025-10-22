// Package renderer provides an abstraction layer for CLI output formatting.
//
// The Renderer pattern separates business logic from output formatting,
// enabling multiple output modes (console, JSON, TUI) without duplicating code.
// Commands produce data, Renderers display it in the appropriate format.
package renderer

// Renderer is the interface that wraps the basic CLI output methods.
//
// Implementations of this interface handle different output modes:
//   - ConsoleRenderer: Human-friendly output with colors and formatting
//   - JSONRenderer: Machine-readable JSON output for scripting
//   - TUIRenderer: Interactive terminal UI (future implementation)
//
// All Renderer methods are designed to be called sequentially during command
// execution, with Flush called at the end to ensure all output is written.
type Renderer interface {
	// Title displays a prominent title or heading.
	// Used for the main title of command output.
	Title(s string)

	// Section displays a titled section with body content.
	// Used for grouping related information under a heading.
	Section(sec Section)

	// Table displays tabular data with headers and rows.
	// Used for structured data like lists of items or summaries.
	Table(t Table)

	// Progress creates and returns a Progress tracker for long-running operations.
	// The returned Progress interface allows incremental updates and completion.
	Progress(spec ProgressSpec) Progress

	// Flush ensures all buffered output is written.
	// Should be called after all other methods to guarantee output visibility.
	// Returns an error if the flush operation fails.
	Flush() error
}
