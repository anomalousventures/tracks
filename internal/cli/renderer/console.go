// Package renderer provides output formatting implementations for CLI commands.
//
// ConsoleRenderer outputs human-readable formatted text to a terminal using
// Lip Gloss styles from the Theme. It automatically respects NO_COLOR and
// other accessibility environment variables through the Theme system.
package renderer

import (
	"fmt"
	"io"

	"github.com/anomalousventures/tracks/internal/cli/ui"
	"github.com/charmbracelet/lipgloss"
)

// ConsoleRenderer implements the Renderer interface for human-readable
// terminal output.
//
// ConsoleRenderer writes formatted output to an io.Writer using Lip Gloss
// styles from the Theme. All styling automatically respects NO_COLOR and
// other accessibility environment variables.
//
// Example usage:
//
//	renderer := NewConsoleRenderer(os.Stdout)
//	renderer.Title("Project Created")
//	renderer.Section(Section{
//	    Title: "Configuration",
//	    Body:  "Using Chi router with templ templates",
//	})
//	renderer.Flush()
//
// ConsoleRenderer is safe for concurrent use from multiple goroutines as long
// as the underlying io.Writer is thread-safe.
type ConsoleRenderer struct {
	out io.Writer
}

// NewConsoleRenderer creates a new ConsoleRenderer that writes to the
// provided io.Writer.
//
// The writer parameter is typically os.Stdout or os.Stderr, but can be
// any io.Writer for testing purposes (e.g., bytes.Buffer).
//
// Example:
//
//	renderer := NewConsoleRenderer(os.Stdout)
func NewConsoleRenderer(out io.Writer) *ConsoleRenderer {
	return &ConsoleRenderer{out: out}
}

// Title displays a prominent title using the Theme.Title style.
//
// The title is rendered with bold purple styling (when colors are enabled)
// and written on its own line.
//
// Example:
//
//	renderer.Title("Welcome to Tracks")
func (r *ConsoleRenderer) Title(s string) {
	fmt.Fprintln(r.out, ui.Theme.Title.Render(s))
}

// Section displays a titled section with body content.
//
// If the section has a title, it is rendered using Theme.Title style.
// The body is rendered as plain text below the title. Both title and
// body are optional and will only be rendered if non-empty.
//
// Example:
//
//	renderer.Section(Section{
//	    Title: "Database Configuration",
//	    Body:  "Using LibSQL with migrations enabled",
//	})
func (r *ConsoleRenderer) Section(sec Section) {
	if sec.Title != "" {
		fmt.Fprintln(r.out, ui.Theme.Title.Render(sec.Title))
	}
	if sec.Body != "" {
		fmt.Fprintln(r.out, sec.Body)
	}
}

// Table displays structured tabular data with aligned columns.
//
// Headers are rendered using Theme.Title style. All columns are
// automatically sized to fit their content with proper alignment.
// Empty tables produce no output.
//
// Example:
//
//	renderer.Table(Table{
//	    Headers: []string{"File", "Status", "Lines"},
//	    Rows: [][]string{
//	        {"user.go", "created", "42"},
//	        {"user_test.go", "created", "128"},
//	    },
//	})
func (r *ConsoleRenderer) Table(t Table) {
	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return
	}

	numCols := len(t.Headers)
	if numCols == 0 && len(t.Rows) > 0 {
		numCols = len(t.Rows[0])
	}

	if numCols == 0 {
		return
	}

	colWidths := make([]int, numCols)

	for i, header := range t.Headers {
		if i < numCols {
			colWidths[i] = lipgloss.Width(header)
		}
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < numCols {
				width := lipgloss.Width(cell)
				if width > colWidths[i] {
					colWidths[i] = width
				}
			}
		}
	}

	for i := range colWidths {
		colWidths[i] += 2
	}

	if len(t.Headers) > 0 {
		headerCells := make([]string, len(t.Headers))
		for i, header := range t.Headers {
			if i < numCols {
				cellStyle := ui.Theme.Title.Width(colWidths[i])
				headerCells[i] = cellStyle.Render(header)
			}
		}
		headerRow := lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)
		fmt.Fprintln(r.out, headerRow)
	}

	for _, row := range t.Rows {
		rowCells := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cellStyle := lipgloss.NewStyle().Width(colWidths[i])
			rowCells[i] = cellStyle.Render(cell)
		}
		rowStr := lipgloss.JoinHorizontal(lipgloss.Top, rowCells...)
		fmt.Fprintln(r.out, rowStr)
	}
}

// Progress is a stub implementation for progress bar rendering.
//
// This method will be fully implemented in a future task. For now, it
// returns a stub Progress implementation that satisfies the interface
// but produces no output.
//
// See: Task 17 - Add Progress bar rendering to ConsoleRenderer
func (r *ConsoleRenderer) Progress(spec ProgressSpec) Progress {
	return &stubProgress{}
}

// Flush ensures all buffered output is written.
//
// For ConsoleRenderer, this is a no-op since fmt.Fprintln writes directly
// to the underlying io.Writer without buffering. This method exists to
// satisfy the Renderer interface and maintain consistency with other
// renderer implementations that may need flushing (like JSONRenderer).
//
// Always returns nil.
func (r *ConsoleRenderer) Flush() error {
	return nil
}

// stubProgress is a placeholder Progress implementation.
//
// This stub exists to allow Progress method to satisfy the Renderer
// interface before full progress bar functionality is implemented.
// All methods are no-ops.
type stubProgress struct{}

// Increment is a no-op stub.
func (p *stubProgress) Increment(n int64) {}

// Done is a no-op stub.
func (p *stubProgress) Done() {}
