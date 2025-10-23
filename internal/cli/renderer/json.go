// Package renderer provides output formatting implementations for CLI commands.
//
// JSONRenderer outputs machine-readable JSON for scripting and automation.
// It accumulates all output (titles, sections, tables) and writes formatted
// JSON when Flush() is called.
package renderer

import (
	"encoding/json"
	"io"
)

// JSONRenderer implements the Renderer interface for machine-readable
// JSON output.
//
// JSONRenderer accumulates all CLI output (titles, sections, tables) in memory
// and writes formatted JSON when Flush() is called. This enables CLI integration
// with scripts, CI/CD pipelines, and other tools that need structured data.
//
// Example usage:
//
//	renderer := NewJSONRenderer(os.Stdout)
//	renderer.Title("Project Created")
//	renderer.Section(Section{
//	    Title: "Configuration",
//	    Body:  "Using Chi router with templ templates",
//	})
//	renderer.Table(Table{
//	    Headers: []string{"File", "Status"},
//	    Rows:    [][]string{{"user.go", "created"}},
//	})
//	renderer.Flush()
//
// Output:
//
//	{
//	  "title": "Project Created",
//	  "sections": [
//	    {"title": "Configuration", "body": "Using Chi router with templ templates"}
//	  ],
//	  "tables": [
//	    {"headers": ["File", "Status"], "rows": [["user.go", "created"]]}
//	  ]
//	}
//
// JSONRenderer is safe for concurrent use from multiple goroutines as long
// as the underlying io.Writer is thread-safe.
type JSONRenderer struct {
	out      io.Writer
	data     *jsonOutput
}

// jsonOutput holds all accumulated data for JSON output.
type jsonOutput struct {
	Title    string    `json:"title,omitempty"`
	Sections []Section `json:"sections"`
	Tables   []Table   `json:"tables"`
}

// NewJSONRenderer creates a new JSONRenderer that writes to the
// provided io.Writer.
//
// The writer parameter is typically os.Stdout or os.Stderr, but can be
// any io.Writer for testing purposes (e.g., bytes.Buffer).
//
// Example:
//
//	renderer := NewJSONRenderer(os.Stdout)
func NewJSONRenderer(out io.Writer) *JSONRenderer {
	return &JSONRenderer{
		out: out,
		data: &jsonOutput{
			Sections: []Section{},
			Tables:   []Table{},
		},
	}
}

// Title stores a prominent title for the JSON output.
//
// The title is rendered as the "title" field in the JSON output.
// Multiple calls to Title will overwrite previous values.
//
// Example:
//
//	renderer.Title("Welcome to Tracks")
func (r *JSONRenderer) Title(s string) {
	r.data.Title = s
}

// Section adds a titled section to the JSON output.
//
// Sections are accumulated and rendered as an array in the "sections"
// field. Multiple calls to Section will append to the array.
//
// Example:
//
//	renderer.Section(Section{
//	    Title: "Database Configuration",
//	    Body:  "Using LibSQL with migrations enabled",
//	})
func (r *JSONRenderer) Section(sec Section) {
	r.data.Sections = append(r.data.Sections, sec)
}

// Table adds structured tabular data to the JSON output.
//
// Tables are accumulated and rendered as an array in the "tables"
// field. Multiple calls to Table will append to the array.
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
func (r *JSONRenderer) Table(t Table) {
	r.data.Tables = append(r.data.Tables, t)
}

// Progress returns a no-op Progress implementation for JSON output.
//
// JSON output is not suitable for incremental progress updates, so this
// method returns a Progress instance that discards all updates. Progress
// is designed for interactive terminals, not machine-readable output.
//
// Example:
//
//	progress := renderer.Progress(ProgressSpec{Label: "Downloading", Total: 100})
//	progress.Increment(50)  // No output
//	progress.Done()         // No output
func (r *JSONRenderer) Progress(spec ProgressSpec) Progress {
	return &jsonProgress{}
}

// Flush writes all accumulated data as formatted JSON.
//
// The JSON output is indented with 2 spaces for readability. After
// flushing, the renderer's internal data is not cleared, so subsequent
// calls to Flush will output the same data (potentially with additions
// from new Title/Section/Table calls).
//
// Always returns nil unless the underlying io.Writer returns an error.
//
// Example:
//
//	renderer.Title("Project Created")
//	renderer.Flush()  // Writes JSON to output
func (r *JSONRenderer) Flush() error {
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(r.data)
}

// jsonProgress is a no-op Progress implementation for JSON output.
//
// Since JSON output is designed for machine-readable structured data,
// incremental progress updates don't make sense. All methods are no-ops.
type jsonProgress struct{}

// Increment is a no-op for JSON progress.
func (p *jsonProgress) Increment(n int64) {}

// Done is a no-op for JSON progress.
func (p *jsonProgress) Done() {}
