# Epic 1: CLI Infrastructure

[← Back to Phase 0](../0-foundation.md)

## Overview

Establish the foundational CLI tool using Cobra framework. This epic creates the basic `tracks` command that can be executed, displays version information, and provides help text. This is the absolute foundation that all other epics depend on.

## Goals

- Working `tracks` binary that can be built and executed
- Version tracking system in place
- Help system functional
- Command structure framework ready for extension
- Basic logging infrastructure

## Scope

### In Scope

- Cobra CLI initialization and configuration
- Root command setup with basic flags
- Version command and `--version` flag
- Help text system (`--help`)
- Build configuration for cross-platform compilation
- Basic structured logging with zerolog
- CLI entry point (`cmd/tracks/main.go`)

### Out of Scope

- Actual command implementations (new, generate, etc.) - those come in later epics
- TUI mode - deferred to later phase
- Configuration file loading - minimal for now
- Advanced CLI features (autocomplete, etc.)

## Task Breakdown

The following tasks will become GitHub issues:

1. **Initialize tracks CLI Go module**
2. **Set up Cobra root command with basic structure**
3. **Implement version tracking and --version flag**
4. **Add comprehensive help text and usage documentation**
5. **Configure zerolog for structured CLI logging**
6. **Create Makefile targets for building CLI**
7. **Add cross-platform build configuration (Linux, macOS, Windows)**
8. **Write unit tests for root command and version**
9. **Set up CLI integration test framework**
10. **Document CLI development workflow**

## Dependencies

### Prerequisites

- Go 1.25+ installed
- Access to github.com/spf13/cobra package

### Blocks

- Epic 2 (Template Engine) - needs CLI to run commands
- Epic 3 (Project Generation) - needs `tracks new` command structure
- All subsequent features depend on working CLI

## Acceptance Criteria

- [ ] `tracks` command builds successfully on Linux, macOS, and Windows
- [ ] `tracks --version` displays correct version information
- [ ] `tracks --help` shows usage information
- [ ] `tracks` with no arguments shows helpful message (TUI placeholder)
- [ ] Logging outputs structured JSON in production mode
- [ ] Unit tests cover root command and version logic
- [ ] CI pipeline can build the CLI binary
- [ ] README documents how to build and run the CLI

## Technical Notes

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

### Logging

Use zerolog for structured logging from the start. Makes debugging easier later.

## Testing Strategy

- Unit tests for command registration and flag parsing
- Integration tests that execute the binary and verify output
- Cross-platform build tests in CI

## Next Epic

[Epic 2: Template Engine & Embedding →](./2-template-engine.md)
