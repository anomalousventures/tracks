# Renderer Pattern Deep Dive

The Renderer pattern is the core abstraction for CLI output in Tracks. It separates **what** to display from **how** to display it.

## Design Philosophy

### Separation of Concerns

```text
Command Logic          Renderer Implementation
┌──────────────┐      ┌────────────────────┐
│              │      │                    │
│ • Validation │──────▶│ • Format output   │
│ • Business   │ data │ • Apply styles    │
│ • Data fetch │──────▶│ • Handle errors   │
│              │      │                    │
└──────────────┘      └────────────────────┘
```

Commands produce **data** (titles, sections, tables). Renderers format that data for different contexts (terminal, JSON, TUI).

### Benefits

1. **Testability** - Mock renderers for unit tests
2. **Flexibility** - Add new output formats without changing commands
3. **Consistency** - All commands use the same output primitives
4. **Maintainability** - Output logic isolated from business logic

## Interface

```go
type Renderer interface {
    // Title displays a prominent heading
    Title(string)

    // Section displays titled content blocks
    Section(Section)

    // Table displays structured data in rows/columns
    Table(Table)

    // Progress creates a tracker for long-running operations
    Progress(ProgressSpec) Progress

    // Flush writes all accumulated output
    Flush() error
}
```

### Supporting Types

```go
// Section represents a content block with optional title
type Section struct {
    Title string  // Optional heading
    Body  string  // Main content
}

// Table represents structured tabular data
type Table struct {
    Headers []string    // Column headers
    Rows    [][]string  // Data rows
}

// ProgressSpec configures a progress tracker
type ProgressSpec struct {
    Label string   // Display label
    Total int64    // Total items
}

// Progress tracks incremental updates
type Progress interface {
    Increment(int64)  // Update progress
    Done()            // Mark complete
}
```

## Implementations

### ConsoleRenderer

Human-readable terminal output using Lip Gloss for styling.

**Features:**

- Colored output (respects NO_COLOR)
- Themed styles (Title, Success, Error, Warning, Muted)
- Table alignment with lipgloss.Table
- Progress bars using Bubbles progress component
- Direct writes (no buffering)

**Usage:**

```go
r := renderer.NewConsoleRenderer(os.Stdout)

r.Title("Installation Complete")
r.Section(renderer.Section{
    Body: "Successfully installed 15 packages",
})
r.Table(renderer.Table{
    Headers: []string{"Package", "Version"},
    Rows: [][]string{
        {"cobra", "1.10.1"},
        {"viper", "1.20.1"},
    },
})

progress := r.Progress(renderer.ProgressSpec{
    Label: "Downloading",
    Total: 100,
})
for i := 0; i <= 100; i++ {
    progress.Increment(1)
    time.Sleep(10 * time.Millisecond)
}
progress.Done()

if err := r.Flush(); err != nil {
    log.Fatal(err)
}
```

**Output:**

```text
Installation Complete

Successfully installed 15 packages

Package         Version
──────────────────────
cobra           1.10.1
viper           1.20.1

Downloading [████████████████████] 100%
```

### JSONRenderer

Machine-readable JSON for scripting and automation.

**Features:**

- Structured JSON output
- Accumulates all data before writing
- Pretty-printed with 2-space indentation
- No-op progress (JSON not suitable for incremental updates)

**Usage:**

```go
r := renderer.NewJSONRenderer(os.Stdout)

r.Title("Installation Complete")
r.Section(renderer.Section{
    Body: "Successfully installed 15 packages",
})
r.Table(renderer.Table{
    Headers: []string{"Package", "Version"},
    Rows: [][]string{
        {"cobra", "1.10.1"},
        {"viper", "1.20.1"},
    },
})

if err := r.Flush(); err != nil {
    log.Fatal(err)
}
```

**Output:**

```json
{
  "title": "Installation Complete",
  "sections": [
    {
      "title": "",
      "body": "Successfully installed 15 packages"
    }
  ],
  "tables": [
    {
      "headers": ["Package", "Version"],
      "rows": [
        ["cobra", "1.10.1"],
        ["viper", "1.20.1"]
      ]
    }
  ]
}
```

### TUIRenderer (Phase 4)

Interactive Bubble Tea interface for immersive experiences.

**Planned Features:**

- Full-screen TUI application
- Keyboard navigation
- Real-time updates
- Form inputs
- File browsers
- Confirmation dialogs

## Implementation Guide

### Creating a New Renderer

Let's create a Markdown renderer as an example.

#### 1. Create File

```go
// internal/cli/renderer/markdown.go
package renderer

import (
    "fmt"
    "io"
    "strings"
)

type MarkdownRenderer struct {
    w io.Writer
}

func NewMarkdownRenderer(w io.Writer) *MarkdownRenderer {
    return &MarkdownRenderer{w: w}
}
```

#### 2. Implement Title

```go
func (r *MarkdownRenderer) Title(s string) {
    fmt.Fprintf(r.w, "# %s\n\n", s)
}
```

#### 3. Implement Section

```go
func (r *MarkdownRenderer) Section(sec Section) {
    if sec.Title != "" {
        fmt.Fprintf(r.w, "## %s\n\n", sec.Title)
    }
    if sec.Body != "" {
        fmt.Fprintf(r.w, "%s\n\n", sec.Body)
    }
}
```

#### 4. Implement Table

```go
func (r *MarkdownRenderer) Table(t Table) {
    if len(t.Headers) == 0 {
        return
    }

    // Headers
    fmt.Fprintf(r.w, "| %s |\n", strings.Join(t.Headers, " | "))

    // Separator
    sep := make([]string, len(t.Headers))
    for i := range sep {
        sep[i] = "---"
    }
    fmt.Fprintf(r.w, "| %s |\n", strings.Join(sep, " | "))

    // Rows
    for _, row := range t.Rows {
        cells := make([]string, len(t.Headers))
        for i := range t.Headers {
            if i < len(row) {
                cells[i] = row[i]
            }
        }
        fmt.Fprintf(r.w, "| %s |\n", strings.Join(cells, " | "))
    }
    fmt.Fprintln(r.w)
}
```

#### 5. Implement Progress

```go
type markdownProgress struct{}

func (p *markdownProgress) Increment(n int64) {}
func (p *markdownProgress) Done()             {}

func (r *MarkdownRenderer) Progress(spec ProgressSpec) Progress {
    // Markdown doesn't support progress bars
    return &markdownProgress{}
}
```

#### 6. Implement Flush

```go
func (r *MarkdownRenderer) Flush() error {
    // Markdown writes immediately, nothing to flush
    return nil
}
```

#### 7. Add Tests

```go
// internal/cli/renderer/markdown_test.go
package renderer_test

import (
    "bytes"
    "strings"
    "testing"

    "github.com/anomalousventures/tracks/internal/cli/renderer"
)

func TestMarkdownRenderer(t *testing.T) {
    var buf bytes.Buffer
    r := renderer.NewMarkdownRenderer(&buf)

    r.Title("Test Title")
    r.Section(renderer.Section{
        Title: "Section",
        Body:  "Content",
    })
    r.Table(renderer.Table{
        Headers: []string{"A", "B"},
        Rows:    [][]string{{"1", "2"}},
    })

    if err := r.Flush(); err != nil {
        t.Fatalf("Flush failed: %v", err)
    }

    output := buf.String()

    if !strings.Contains(output, "# Test Title") {
        t.Error("Missing title")
    }
    if !strings.Contains(output, "## Section") {
        t.Error("Missing section title")
    }
    if !strings.Contains(output, "| A | B |") {
        t.Error("Missing table")
    }
}
```

#### 8. Integrate with CLI

Update `internal/cli/root.go`:

```go
// Add mode constant in ui/mode.go
const (
    ModeAuto UIMode = iota
    ModeConsole
    ModeJSON
    ModeTUI
    ModeMarkdown  // New
)

// Update DetectMode() in ui/mode.go
func DetectMode(cfg UIConfig) UIMode {
    // ... existing logic ...
    if cfg.Markdown {  // Add Markdown flag
        return ModeMarkdown
    }
    // ... rest of detection ...
}

// Update NewRendererFromCommand() in root.go
func NewRendererFromCommand(cmd *cobra.Command) renderer.Renderer {
    cfg := GetConfig(cmd)
    uiMode := ui.DetectMode(...)

    switch uiMode {
    case ui.ModeJSON:
        return renderer.NewJSONRenderer(cmd.OutOrStdout())
    case ui.ModeMarkdown:
        return renderer.NewMarkdownRenderer(cmd.OutOrStdout())
    default:
        return renderer.NewConsoleRenderer(cmd.OutOrStdout())
    }
}

// Add flag in NewRootCmd()
rootCmd.PersistentFlags().Bool("markdown", false, "Output in Markdown format")
```

## Design Patterns

### Accumulate-Then-Flush

Some renderers need to see all data before rendering (e.g., JSON needs complete structure).

```go
type AccumulatingRenderer struct {
    data  OutputData
    w     io.Writer
}

func (r *AccumulatingRenderer) Title(s string) {
    r.data.Title = s  // Store, don't write yet
}

func (r *AccumulatingRenderer) Flush() error {
    // Now write everything as complete structure
    return json.NewEncoder(r.w).Encode(r.data)
}
```

### Immediate Write

Other renderers write immediately (e.g., Console for interactive feedback).

```go
type ImmediateRenderer struct {
    w io.Writer
}

func (r *ImmediateRenderer) Title(s string) {
    fmt.Fprintln(r.w, s)  // Write immediately
}

func (r *ImmediateRenderer) Flush() error {
    return nil  // Nothing buffered
}
```

### No-Op Progress

If a renderer can't support progress bars, provide a no-op implementation:

```go
type noOpProgress struct{}

func (p *noOpProgress) Increment(n int64) {}
func (p *noOpProgress) Done()             {}

func (r *MyRenderer) Progress(spec ProgressSpec) Progress {
    return &noOpProgress{}
}
```

## Testing Strategies

### Test All Implementations

Use table-driven tests to ensure all renderers work:

```go
func TestRendererImplementations(t *testing.T) {
    tests := []struct {
        name string
        r    renderer.Renderer
    }{
        {"console", renderer.NewConsoleRenderer(&bytes.Buffer{})},
        {"json", renderer.NewJSONRenderer(&bytes.Buffer{})},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.r.Title("Test")
            if err := tt.r.Flush(); err != nil {
                t.Errorf("Flush failed: %v", err)
            }
        })
    }
}
```

### Mock Renderer for Command Tests

```go
type MockRenderer struct {
    Titles   []string
    Sections []renderer.Section
    Flushed  bool
}

func (m *MockRenderer) Title(s string) {
    m.Titles = append(m.Titles, s)
}

func (m *MockRenderer) Section(s renderer.Section) {
    m.Sections = append(m.Sections, s)
}

func (m *MockRenderer) Flush() error {
    m.Flushed = true
    return nil
}

// Use in tests
func TestMyCommand(t *testing.T) {
    mock := &MockRenderer{}
    myCommandLogic(mock)

    if len(mock.Titles) != 1 {
        t.Error("Expected 1 title")
    }
    if !mock.Flushed {
        t.Error("Renderer not flushed")
    }
}
```

### Integration Tests

Test complete flow with real renderers:

```go
func TestConsoleOutput(t *testing.T) {
    var buf bytes.Buffer
    r := renderer.NewConsoleRenderer(&buf)

    r.Title("Test")
    r.Section(renderer.Section{Body: "Content"})
    r.Flush()

    output := buf.String()
    if !strings.Contains(output, "Test") {
        t.Error("Title missing from output")
    }
}
```

## Common Pitfalls

### 1. Forgetting Flush

```go
// BAD - no output appears
r.Title("Hello")
// Forgot r.Flush()!
```

### 2. Flushing Too Early

```go
// BAD - JSON would be incomplete
r.Title("Test")
r.Flush()       // Too early!
r.Section(...)  // Won't be included
```

### 3. Assuming IO Never Fails

```go
// BAD - ignoring error
r.Flush()

// GOOD - checking error
if err := r.Flush(); err != nil {
    return fmt.Errorf("output failed: %w", err)
}
```

### 4. Testing Output Format

```go
// BAD - brittle test
if output != "exact string" { ... }

// GOOD - semantic test
if !strings.Contains(output, "key content") { ... }
```

### 5. Not Implementing All Methods

```go
// BAD - panic at runtime
func (r *MyRenderer) Progress(...) Progress {
    panic("not implemented")
}

// GOOD - no-op implementation
func (r *MyRenderer) Progress(...) Progress {
    return &noOpProgress{}
}
```

## Performance Considerations

### Buffering

For network/file writes, use buffering:

```go
func NewFileRenderer(filename string) (*FileRenderer, error) {
    f, err := os.Create(filename)
    if err != nil {
        return nil, err
    }
    return &FileRenderer{
        w: bufio.NewWriter(f),  // Buffered!
        f: f,
    }, nil
}

func (r *FileRenderer) Flush() error {
    // Flush buffer first
    if err := r.w.(*bufio.Writer).Flush(); err != nil {
        return err
    }
    // Then sync file
    return r.f.Sync()
}
```

### Large Tables

Stream large tables row-by-row instead of buffering:

```go
func (r *StreamRenderer) Table(t Table) {
    r.writeHeaders(t.Headers)
    for _, row := range t.Rows {
        r.writeRow(row)  // Write immediately
    }
}
```

## Future Enhancements

### Planned Features (Phase 4)

- **TUIRenderer** - Full-screen Bubble Tea interface
- **HTMLRenderer** - For generated documentation
- **MarkdownRenderer** - For README generation
- **Streaming renderers** - For long-running commands

### Extension Points

The Renderer interface may grow to support:

- Prompts/Input (`Prompt(question string) (answer string)`)
- Confirmations (`Confirm(message string) bool`)
- Multi-select (`Select(options []string) []int`)

## Related Documentation

- [CLI Architecture](../README.md) - Overall CLI design
- [Theme System](../ui/theme.go) - Lip Gloss styles
- [Mode Detection](../ui/mode.go) - Output mode logic
