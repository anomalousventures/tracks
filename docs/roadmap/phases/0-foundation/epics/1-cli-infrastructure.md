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
- Configuration file loading from tracks.yaml - deferred to later phases (Viper infrastructure ready)
- Advanced CLI features (autocomplete, etc.)

## Task Breakdown

The following tasks will become GitHub issues, ordered by dependency:

1. ✅ **Initialize tracks CLI Go module** (Complete)
2. ✅ **Set up Cobra root command with basic structure** (Complete)
3. ✅ **Implement version tracking and --version flag** (Complete)
4. ✅ **Define global CLI flags (--json, --no-color, --interactive)** ([#6](https://github.com/anomalousventures/tracks/issues/6))
5. ✅ **Add comprehensive help text with examples and flag documentation** ([#7](https://github.com/anomalousventures/tracks/issues/7) | [PR #26](https://github.com/anomalousventures/tracks/pull/26))
6. ✅ **Add TUI placeholder message for no-args execution** ([#8](https://github.com/anomalousventures/tracks/issues/8) | [PR #27](https://github.com/anomalousventures/tracks/pull/27))
7. ✅ **Define UIMode enum and UIConfig struct** ([#9](https://github.com/anomalousventures/tracks/issues/9) | [PR #28](https://github.com/anomalousventures/tracks/pull/28))
8. ✅ **Implement basic DetectMode with TTY and CI detection** ([#10](https://github.com/anomalousventures/tracks/issues/10) | [PR #29](https://github.com/anomalousventures/tracks/pull/29))
9. ✅ **Wire up flag support to mode detection** ([#11](https://github.com/anomalousventures/tracks/issues/11) | [PR #30](https://github.com/anomalousventures/tracks/pull/30))
10. ✅ **Wire up environment variable support (NO_COLOR, TRACKS_LOG_LEVEL, --verbose, --quiet)** ([#12](https://github.com/anomalousventures/tracks/issues/12) | [PR #31](https://github.com/anomalousventures/tracks/pull/31))
11. ✅ **Define Renderer interface and core types (Section, Table, Progress)** ([#13](https://github.com/anomalousventures/tracks/issues/13) | [PR #33](https://github.com/anomalousventures/tracks/pull/33))
12. ✅ **Create Lip Gloss theme system with color styles and NO_COLOR support** ([#14](https://github.com/anomalousventures/tracks/issues/14))
13. **Implement ConsoleRenderer with Title and Section rendering** ([#15](https://github.com/anomalousventures/tracks/issues/15))
14. **Add Table rendering to ConsoleRenderer** ([#16](https://github.com/anomalousventures/tracks/issues/16))
15. **Add Progress bar rendering to ConsoleRenderer using Bubbles ViewAs** ([#17](https://github.com/anomalousventures/tracks/issues/17))
16. **Implement JSONRenderer with all Renderer methods** ([#18](https://github.com/anomalousventures/tracks/issues/18))
17. **Create Makefile targets for building CLI** ([#19](https://github.com/anomalousventures/tracks/issues/19))
18. **Add cross-platform build configuration (Linux, macOS, Windows)** ([#20](https://github.com/anomalousventures/tracks/issues/20))
19. **Write unit tests for mode detection and flags** ([#21](https://github.com/anomalousventures/tracks/issues/21))
20. **Write unit tests for Renderer implementations** ([#22](https://github.com/anomalousventures/tracks/issues/22))
21. **Set up CLI integration test framework** ([#23](https://github.com/anomalousventures/tracks/issues/23))
22. **Document CLI development workflow and Renderer pattern** ([#24](https://github.com/anomalousventures/tracks/issues/24))

## Dependencies

### Prerequisites

- Go 1.25+ installed
- Access to external packages:
  - github.com/spf13/cobra v1.10.1 (already in go.mod)
  - github.com/spf13/viper v1.20.1 (already in go.mod)
  - github.com/charmbracelet/bubbletea v1.3.0 (already in go.mod)
  - github.com/charmbracelet/lipgloss v1.1.0 (already in go.mod)
  - github.com/charmbracelet/bubbles v0.21.0 (to be added)
  - github.com/charmbracelet/glamour v0.10.0 (to be added)
  - github.com/mattn/go-isatty v0.0.20 (already in go.mod)

Note: Versions listed are current as of October 2025. Use `go get <package>@latest` during implementation to ensure latest stable versions. Development tools will be added to the `tool` directive in go.mod.

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

### Note on Modern Go Tooling Pattern

The CLI and generated projects use the `go tool <name>` pattern for all development tools (golangci-lint, air, etc.) instead of requiring global installations. Tools are declared in `go.mod` using the `tool` directive (Go 1.25+) and invoked via `go tool <name>`. This ensures consistent tool versions across all developers and CI environments, eliminates dependency on global installs, and makes the toolchain fully reproducible.

## Testing Strategy

- Unit tests for command registration and flag parsing
- Integration tests that execute the binary and verify output
- Cross-platform build tests in CI

## Next Epic

[Epic 2: Template Engine & Embedding →](./2-template-engine.md)
