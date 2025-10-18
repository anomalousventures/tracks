# Epic 1: CLI Infrastructure

[← Back to Phase 0](../0-foundation.md)

## Overview

Establish the foundational CLI tool using Cobra framework. This epic creates the basic `tracks` command that can be executed, displays version information, and provides help text. This is the absolute foundation that all other epics depend on.

## Goals

- Working `tracks` binary that can be built and executed
- Version tracking system in place
- Help system functional
- Command structure framework ready for extension
- CLI output infrastructure with Renderer pattern
- Consistent styling across console and TUI modes

## Scope

### In Scope

- Cobra CLI initialization and configuration
- Root command setup with basic flags
- Version command and `--version` flag
- Help text system (`--help`)
- Build configuration for cross-platform compilation
- Renderer pattern implementation
- UIMode detection (auto, console, JSON, TUI-ready)
- Lip Gloss theme system
- Bubbles components for console (standalone mode)
- Environment variables (NO_COLOR, TRACKS_LOG_LEVEL)
- Flags (--json, --no-color, --interactive)
- CLI entry point (`cmd/tracks/main.go`)

### Out of Scope

- Actual command implementations (new, generate, etc.) - those come in later epics
- Full TUI implementation - deferred to Phase 4
- Structured logging with zerolog - only for generated apps, not CLI tool
- Configuration file loading - minimal for now
- Advanced CLI features (autocomplete, etc.)

## Task Breakdown

The following tasks will become GitHub issues:

1. **Initialize tracks CLI Go module**
2. **Set up Cobra root command with basic structure**
3. **Implement version tracking and --version flag**
4. **Add comprehensive help text and TUI placeholder message**
5. **Implement mode detection and Renderer interfaces**
6. **Implement ConsoleRenderer with Charm stack**
7. **Implement JSONRenderer**
8. **Create Makefile targets for building CLI**
9. **Add cross-platform build configuration (Linux, macOS, Windows)**
10. **Write unit tests for root command, version, and Renderer**
11. **Set up CLI integration test framework**
12. **Document CLI development workflow and Renderer pattern**

## Dependencies

### Prerequisites

- Go 1.25+ installed
- Access to github.com/spf13/cobra package
- Access to Charm stack packages (lipgloss, bubbles, glamour)
- Access to github.com/mattn/go-isatty

### Blocks

- Epic 2 (Template Engine) - needs CLI to run commands
- Epic 3 (Project Generation) - needs `tracks new` command structure
- All subsequent features depend on working CLI

## Acceptance Criteria

- [ ] `tracks` command builds successfully on Linux, macOS, and Windows
- [ ] `tracks --version` displays correct version information
- [ ] `tracks --help` shows usage information with examples
- [ ] `tracks` with no arguments shows TUI placeholder message
- [ ] Renderer interface defined and documented
- [ ] UIMode detection works (TTY, CI, flags, env vars)
- [ ] ConsoleRenderer works with Lip Gloss styling
- [ ] Bubbles progress bar renders in console mode (standalone)
- [ ] JSONRenderer outputs valid JSON
- [ ] NO_COLOR env var disables colors
- [ ] TRACKS_LOG_LEVEL env var controls verbosity
- [ ] --json flag outputs JSON
- [ ] --no-color flag disables colors
- [ ] Theme defined once, used everywhere
- [ ] Unit tests cover root command, version, and Renderer modes
- [ ] CI pipeline can build the CLI binary
- [ ] README documents Renderer pattern and output modes

## Technical Notes

### Renderer Pattern

The Renderer pattern separates business logic from output formatting. Commands return data, Renderer displays it.

```go
// internal/cli/renderer/renderer.go
type Renderer interface {
    Title(s string)
    Section(sec Section)
    Table(t Table)
    Progress(spec ProgressSpec) Progress
    Flush() error
}

type Section struct{ Title string; Body string }
type Table struct{ Headers []string; Rows [][]string }
type ProgressSpec struct{ Label string; Total int64 }

type Progress interface {
    Increment(n int64)
    Done()
}
```

### Mode Detection

```go
// internal/cli/ui/mode.go
type UIMode int
const (
    ModeAuto UIMode = iota
    ModeConsole
    ModeJSON
    ModeTUI
)

func DetectMode(cfg UIConfig) UIMode {
    if cfg.Mode != ModeAuto { return cfg.Mode }
    if os.Getenv("CI") != "" || !isatty.IsTerminal(os.Stdout.Fd()) {
        return ModeConsole
    }
    return ModeConsole // TUI in Phase 4
}
```

### Theme System

Define styles once with Lip Gloss, reuse everywhere (console and TUI).

```go
// internal/cli/ui/theme.go
var Theme = struct {
    Title    lipgloss.Style
    Success  lipgloss.Style
    Error    lipgloss.Style
    Warning  lipgloss.Style
    Muted    lipgloss.Style
}{
    Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
    Success: lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")),
    Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672")),
    Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA657")),
    Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
}

// Usage in console renderer
func (r *ConsoleRenderer) Title(s string) {
    fmt.Println(Theme.Title.Render(s))
}
```

### Console Progress (Bubbles Standalone)

Bubbles components work without Bubble Tea event loop using ViewAs:

```go
// Render progress bar directly in console mode
progressBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
for current := 0; current <= total; current++ {
    percent := float64(current) / float64(total)
    fmt.Print("\r" + progressBar.ViewAs(percent))
    // do work
}
fmt.Println() // newline after progress completes
```

### Version Tracking

Use Go build flags to embed version information:

```bash
go build -ldflags "-X main.Version=v0.1.0 -X main.Commit=$(git rev-parse HEAD)"
```

### Command Structure

Keep root command minimal. Structure for extensibility:

```go
// cmd/tracks/main.go
func main() {
    cmd.Execute()
}

// internal/cli/root.go
var rootCmd = &cobra.Command{
    Use:   "tracks",
    Short: "Tracks - Rails-inspired Go web framework",
    Long:  `Tracks is a CLI tool for generating production-ready Go web applications.`,
}
```

### Note on Logging vs Output

The tracks CLI tool uses the Renderer pattern for user-facing output (progress bars, success messages, table data). Structured logging with zerolog is for generated applications (web servers), not the CLI tool itself. This separation keeps the CLI tool friendly for developers while generated apps remain production-ready with structured logs.

## Testing Strategy

- Unit tests for command registration and flag parsing
- Integration tests that execute the binary and verify output
- Cross-platform build tests in CI

## Next Epic

[Epic 2: Template Engine & Embedding →](./2-template-engine.md)
