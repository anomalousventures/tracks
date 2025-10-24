# CLI Infrastructure

This package implements the Tracks CLI tool using Cobra, Viper, and the Charm TUI stack (Lip Gloss, Bubbles, Bubble Tea).

## Architecture Overview

The CLI uses the **Renderer Pattern** to separate business logic from output formatting. Commands produce data, Renderers display it.

```text
┌─────────────┐
│   Command   │  Business logic (data gathering, validation)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Renderer   │  Output formatting (console, JSON, TUI)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Output    │  stdout/stderr
└─────────────┘
```

### Key Components

#### 1. Root Command (`root.go`)

- Defines global flags (`--json`, `--no-color`, `--interactive`, `-v`, `-q`)
- Manages configuration via Viper and context
- Creates Renderer instances for commands
- Handles error flushing

#### 2. Renderer Interface (`renderer/renderer.go`)

Defines how output is displayed:

```go
type Renderer interface {
    Title(string)
    Section(Section)
    Table(Table)
    Progress(ProgressSpec) Progress
    Flush() error
}
```

**Implementations:**

- **ConsoleRenderer** - Human-readable, colored output using Lip Gloss
- **JSONRenderer** - Machine-readable JSON for scripting
- **TUIRenderer** - Interactive Bubble Tea interface (Phase 4)

#### 3. UI Mode Detection (`ui/mode.go`)

Automatically chooses the right output mode:

```text
Priority (highest to lowest):
1. --json flag           → ModeJSON
2. --interactive flag    → ModeTUI (Phase 4: returns ModeConsole)
3. CI environment        → ModeConsole
4. Non-TTY stdout        → ModeConsole
5. Default (TTY)         → ModeConsole (Phase 4: ModeTUI)
```

#### 4. Theme System (`ui/theme.go`)

Centralized Lip Gloss styles used by ConsoleRenderer and TUIRenderer:

- **Title** - Bold purple (#7D56F4)
- **Success** - Green (#04B575)
- **Error** - Red (#FF4672)
- **Warning** - Orange (#FFA657)
- **Muted** - Gray (#626262)

Automatically respects `NO_COLOR` environment variable.

## Usage Examples

### Basic Command Structure

```go
func myCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "mycommand",
        Short: "Does something useful",
        Run: func(cmd *cobra.Command, args []string) {
            // 1. Create renderer
            r := NewRendererFromCommand(cmd)

            // 2. Produce output
            r.Title("My Command")
            r.Section(renderer.Section{
                Body: "Command completed successfully",
            })

            // 3. Flush
            flushRenderer(cmd, r)
        },
    }
}
```

### Using Tables

```go
r.Table(renderer.Table{
    Headers: []string{"Name", "Version", "Status"},
    Rows: [][]string{
        {"tracks", "0.1.0", "active"},
        {"cobra", "1.10.1", "stable"},
    },
})
```

### Using Progress Bars

```go
progress := r.Progress(renderer.ProgressSpec{
    Label: "Installing dependencies",
    Total: 100,
})

for i := 0; i <= 100; i++ {
    progress.Increment(1)
    time.Sleep(50 * time.Millisecond)
}

progress.Done()
```

### JSON Output

All commands automatically support `--json`:

```bash
$ tracks version --json
{
  "title": "Tracks v0.1.0",
  "sections": [
    {
      "title": "",
      "body": "Commit: abc123\nBuilt: 2025-10-24"
    }
  ]
}
```

## Mode Detection

### Environment Variables

- **NO_COLOR** - Disables colors (standard env var)
- **TRACKS_JSON** - Forces JSON mode
- **TRACKS_NO_COLOR** - Same as NO_COLOR
- **TRACKS_INTERACTIVE** - Forces interactive TUI mode
- **TRACKS_LOG_LEVEL** - Sets verbosity (debug, info, warn, error, off)
- **CI** - Detected automatically, forces console mode

### Flags

- `--json` - Output JSON (highest priority)
- `--no-color` - Disable colors
- `--interactive` - Force interactive TUI mode
- `-v, --verbose` - Enable verbose output
- `-q, --quiet` - Suppress non-error output

**Mutual Exclusivity:** `--verbose` and `--quiet` cannot be used together.

## Adding New Commands

1. **Create command function** in `root.go` or separate file:

```go
func newFeatureCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "feature [args]",
        Short: "Does something",
        Long:  "Detailed description...",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            r := NewRendererFromCommand(cmd)

            // Your logic here

            flushRenderer(cmd, r)
        },
    }
}
```

1. **Register in NewRootCmd:**

```go
rootCmd.AddCommand(newFeatureCmd())
```

1. **Add tests** in `root_test.go` and `test/integration/cli_test.go`

## Extending the Renderer

### Adding a New Renderer Implementation

1. **Create new file** (e.g., `renderer/markdown.go`)
2. **Implement Renderer interface:**

```go
type MarkdownRenderer struct {
    w io.Writer
}

func NewMarkdownRenderer(w io.Writer) *MarkdownRenderer {
    return &MarkdownRenderer{w: w}
}

func (r *MarkdownRenderer) Title(s string) {
    fmt.Fprintf(r.w, "# %s\n\n", s)
}

func (r *MarkdownRenderer) Section(sec Section) {
    if sec.Title != "" {
        fmt.Fprintf(r.w, "## %s\n\n", sec.Title)
    }
    fmt.Fprintf(r.w, "%s\n\n", sec.Body)
}

// ... implement other methods
```

1. **Update NewRendererFromCommand** to support new mode
2. **Add tests** in `renderer/markdown_test.go`

### Adding a New UIMode

1. **Add constant** to `ui/mode.go`:

```go
const (
    ModeAuto UIMode = iota
    ModeConsole
    ModeJSON
    ModeTUI
    ModeMarkdown  // New mode
)
```

1. **Update String() method**
2. **Update DetectMode()** logic
3. **Add flag or env var** support in `root.go`

## Testing Strategies

### Unit Tests

Test individual components in isolation:

```go
func TestMyCommand(t *testing.T) {
    var buf bytes.Buffer

    cmd := myCmd()
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"arg1"})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("command failed: %v", err)
    }

    output := buf.String()
    if !strings.Contains(output, "expected") {
        t.Errorf("unexpected output: %s", output)
    }
}
```

### Integration Tests

Test the compiled binary:

```go
//go:build integration

func TestCLIVersion(t *testing.T) {
    stdout, _ := RunCLIExpectSuccess(t, "version")
    AssertContains(t, stdout, "Tracks")
}
```

Run with: `make test-integration`

### Renderer Tests

Use table-driven tests for different renderers:

```go
func TestRenderersImplementInterface(t *testing.T) {
    var buf bytes.Buffer

    renderers := map[string]renderer.Renderer{
        "console": renderer.NewConsoleRenderer(&buf),
        "json":    renderer.NewJSONRenderer(&buf),
    }

    for name, r := range renderers {
        t.Run(name, func(t *testing.T) {
            r.Title("Test")
            // assertions...
        })
    }
}
```

## Development Workflow

### Building

```bash
# Build CLI binary
make build

# Binary outputs to: ./bin/tracks
```

### Testing

```bash
# Run unit tests
make test

# Run integration tests (requires build)
make test-integration

# Run all tests
make test-all

# Run with coverage
make test-coverage
```

### Linting

```bash
# Run all linters
make lint

# Run specific linters
make lint-go
make lint-md
```

### Running Locally

```bash
# Run without installing
./bin/tracks version

# Install to GOPATH
go install ./cmd/tracks

# Run installed version
tracks version
```

### Debugging

```bash
# Enable verbose output
tracks -v version

# Output JSON for inspection
tracks --json version | jq

# Check environment variable detection
TRACKS_LOG_LEVEL=debug tracks version
```

## Common Pitfalls

### 1. Forgetting to Flush

**Problem:** Output doesn't appear

```go
// BAD
r.Title("Hello")
// Forgot to flush!
```

**Solution:** Always call `flushRenderer(cmd, r)` or `r.Flush()`

### 2. Not Using Helper Functions

**Problem:** Repeating renderer initialization

```go
// BAD - duplicated in every command
cfg := GetConfig(cmd)
uiMode := ui.DetectMode(...)
var r renderer.Renderer
if uiMode == ui.ModeJSON {
    r = renderer.NewJSONRenderer(...)
}
```

**Solution:** Use the helper

```go
// GOOD
r := NewRendererFromCommand(cmd)
```

### 3. Wrong Zero Values in Structs

**Problem:** Explicitly setting zero values

```go
// BAD - Title: "" is unnecessary
r.Section(renderer.Section{
    Title: "",
    Body:  "content",
})
```

**Solution:** Omit zero-value fields

```go
// GOOD
r.Section(renderer.Section{
    Body: "content",
})
```

### 4. Calling Flush Multiple Times

**Problem:** Calling Flush in a loop

```go
// BAD
for _, item := range items {
    r.Section(...)
    r.Flush()  // Don't flush inside loop!
}
```

**Solution:** Accumulate, then flush once

```go
// GOOD
for _, item := range items {
    r.Section(...)
}
r.Flush()  // Flush once at end
```

### 5. Forgetting Integration Tests

**Problem:** Only unit testing commands

**Solution:** Add integration tests that execute the binary:

```go
func TestCLIMyCommand(t *testing.T) {
    stdout, _ := RunCLIExpectSuccess(t, "mycommand", "arg")
    AssertContains(t, stdout, "expected output")
}
```

## Theme Customization

### Using Theme in Custom Code

```go
import "github.com/anomalousventures/tracks/internal/cli/ui"

// Render styled output
fmt.Println(ui.Theme.Title.Render("My Title"))
fmt.Println(ui.Theme.Success.Render("✓ Success"))
fmt.Println(ui.Theme.Error.Render("✗ Error"))
```

### Checking NO_COLOR

Theme automatically respects `NO_COLOR`, but you can check manually:

```go
import "github.com/charmbracelet/lipgloss"

if lipgloss.HasDarkBackground() {
    // Adjust colors for dark terminals
}
```

## Helper Functions Reference

### NewRendererFromCommand

Creates appropriate renderer based on flags/env vars.

```go
func NewRendererFromCommand(cmd *cobra.Command) renderer.Renderer
```

**Returns:** ConsoleRenderer or JSONRenderer based on configuration.

### flushRenderer

Flushes renderer and handles errors.

```go
func flushRenderer(cmd *cobra.Command, r renderer.Renderer)
```

**Effect:** Writes output, prints errors to stderr, exits on failure.

### GetConfig

Extracts CLI configuration from command context.

```go
func GetConfig(cmd *cobra.Command) Config
```

**Returns:** Config struct with all flag and env var values.

### GetViper

Retrieves Viper instance from command context.

```go
func GetViper(cmd *cobra.Command) *viper.Viper
```

**Returns:** Viper instance or new instance if none in context.

## Related Documentation

- [Renderer Deep Dive](./renderer/README.md) - Detailed Renderer pattern docs
- [Project README](../../README.md) - Overview and vision
- [Contributing Guide](../../CONTRIBUTING.md) - Development guidelines
- [Release Process](../../docs/RELEASING.md) - Version management

## Future: Interactive TUI (Phase 4)

The TUI mode will use Bubble Tea for interactive experiences:

- File browser for code generation
- Form-based configuration
- Real-time progress updates
- Keyboard navigation

The Renderer pattern is designed to make this easy - just add `TUIRenderer` and update `DetectMode()`.
