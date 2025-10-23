package renderer

// Section represents a titled content block.
//
// Sections are used to organize output into logical groups,
// each with a heading (Title) and descriptive content (Body).
//
// Example:
//
//	Section{
//	    Title: "Project Configuration",
//	    Body: "Using Chi router with templ templates and SQLC for database access.",
//	}
type Section struct {
	// Title is the section heading, displayed prominently.
	Title string `json:"title"`

	// Body is the section content, may span multiple lines.
	Body string `json:"body"`
}

// Table represents structured tabular data.
//
// Tables display data in rows and columns with headers.
// All rows should have the same number of columns as Headers.
//
// Example:
//
//	Table{
//	    Headers: []string{"Name", "Type", "Status"},
//	    Rows: [][]string{
//	        {"user.go", "model", "created"},
//	        {"user_test.go", "test", "created"},
//	    },
//	}
type Table struct {
	// Headers are the column names displayed at the top of the table.
	Headers []string `json:"headers"`

	// Rows contains the table data, where each row is a slice of cell values.
	// Rows with fewer cells than Headers will be padded with empty strings.
	// Extra cells beyond the number of headers are ignored.
	Rows [][]string `json:"rows"`
}

// ProgressSpec specifies the configuration for a progress tracker.
//
// Used to create a Progress instance for tracking long-running operations
// such as file downloads, code generation, or batch processing.
//
// Example:
//
//	spec := ProgressSpec{
//	    Label: "Generating files",
//	    Total: 42,
//	}
//	progress := renderer.Progress(spec)
type ProgressSpec struct {
	// Label is the human-readable description of the operation being tracked.
	Label string

	// Total is the maximum value representing 100% completion.
	// Progress updates increment toward this value.
	Total int64
}

// Progress tracks incremental updates for long-running operations.
//
// Progress instances are created by calling Renderer.Progress() with a ProgressSpec.
// Use Increment to update progress and Done to mark completion.
//
// Example:
//
//	progress := renderer.Progress(ProgressSpec{Label: "Processing", Total: 100})
//	for i := 0; i < 100; i++ {
//	    // do work
//	    progress.Increment(1)
//	}
//	progress.Done()
type Progress interface {
	// Increment advances the progress by the specified amount.
	// The total of all increments should equal ProgressSpec.Total.
	Increment(n int64)

	// Done marks the progress as complete and finalizes the display.
	// Should be called after all increments are complete.
	Done()
}
