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
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

const (
	columnPadding = 2
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
	if numCols == 0 && len(t.Rows) > 0 && len(t.Rows[0]) > 0 {
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
		if i < len(colWidths)-1 {
			colWidths[i] += columnPadding
		}
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

// Progress creates a progress bar for tracking long-running operations.
//
// Returns a ConsoleProgress instance that renders using Bubbles progress
// component with gradient colors. Uses ViewAs for standalone rendering
// without requiring a Bubble Tea event loop.
//
// The progress bar updates in-place using \r (carriage return) and
// completes with a newline when Done() is called.
//
// Example:
//
//	progress := renderer.Progress(ProgressSpec{Label: "Downloading", Total: 100})
//	progress.Increment(25)  // 25%
//	progress.Increment(50)  // 75%
//	progress.Increment(25)  // 100%
//	progress.Done()         // Adds newline
func (r *ConsoleRenderer) Progress(spec ProgressSpec) Progress {
	bar := progress.New(progress.WithScaledGradient("#7D56F4", "#04B575"))
	return &ConsoleProgress{
		out:   r.out,
		bar:   bar,
		label: spec.Label,
		total: spec.Total,
	}
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

// ConsoleProgress implements Progress interface using Bubbles progress component.
//
// Renders a progress bar using ViewAs for standalone rendering without
// Bubble Tea event loop. Updates are written in-place using \r prefix.
type ConsoleProgress struct {
	out     io.Writer
	bar     progress.Model
	label   string
	total   int64
	current int64
	done    bool
}

// Increment updates the progress bar by the specified amount.
//
// Calculates the new percentage and renders the progress bar in-place
// using \r (carriage return). Handles edge cases like zero total and
// overflow gracefully. If a label was provided, it is displayed before
// the progress bar using the Muted theme style.
func (p *ConsoleProgress) Increment(n int64) {
	p.current += n

	var percent float64
	if p.total > 0 {
		percent = float64(p.current) / float64(p.total)
		if percent > 1.0 {
			percent = 1.0
		}
	} else {
		percent = 1.0
	}

	output := "\r"
	if p.label != "" {
		output += ui.Theme.Muted.Render(p.label+": ")
	}
	output += p.bar.ViewAs(percent)

	fmt.Fprint(p.out, output)
}

// Done completes the progress bar and adds a newline.
//
// Marks the progress as complete and writes a final newline to move
// to the next line. Subsequent calls are idempotent (no additional output).
func (p *ConsoleProgress) Done() {
	if p.done {
		return
	}
	p.done = true
	fmt.Fprintln(p.out)
}
