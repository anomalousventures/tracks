# ADR-001: Dependency Injection Pattern for CLI Commands

**Status:** Accepted
**Date:** 2025-10-27
**Context:** Epic 0.5 - Architecture Alignment

## Context

During Epic 3 Phases 1-2 implementation, we identified that CLI commands were using direct instantiation of dependencies, making them untestable and tightly coupled. This pattern violated the clean architecture principles we established for generated applications.

**Problem:**

```go
// ❌ Direct instantiation, untestable
func newCmd() *cobra.Command {
    return &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            validator := generator.NewValidator()  // Hard-coded dependency
            // Can't mock, can't test
        },
    }
}
```

The CLI commands were inconsistent with generated server code which uses dependency injection throughout.

## Decision

We will adopt a dependency injection pattern for all CLI commands:

1. **Command Struct Pattern** - Each command is a struct with injected dependencies
2. **Constructor Functions** - Commands are created via `New<Command>Command(deps...)` constructors
3. **Command() Method** - Commands expose a `Command() *cobra.Command` method
4. **Interface-Based Dependencies** - All dependencies use interfaces defined in consumer packages

**Example:**

```go
// ✅ Dependency injection, testable
type NewCommand struct {
    validator interfaces.Validator
    generator interfaces.ProjectGenerator
}

func NewNewCommand(v interfaces.Validator, g interfaces.ProjectGenerator) *NewCommand {
    return &NewCommand{validator: v, generator: g}
}

func (c *NewCommand) Command() *cobra.Command {
    return &cobra.Command{
        Use:   "new [project-name]",
        Short: "Create a new Tracks application",
        RunE:  c.run,
    }
}

func (c *NewCommand) run(cmd *cobra.Command, args []string) error {
    // Use injected dependencies
    return c.validator.ValidateProjectName(args[0])
}
```

## Consequences

### Positive

- **Testability** - Commands can be unit tested with mocked dependencies
- **Consistency** - CLI follows same patterns as generated server code
- **Flexibility** - Easy to swap implementations or add new dependencies
- **Maintainability** - Clear dependency graph, no hidden dependencies
- **Documentation** - Constructor signature documents required dependencies

### Negative

- **More Boilerplate** - Requires struct definition + constructor + Command() method
- **Migration Effort** - Existing commands must be refactored (2 commands: new, version)

### Neutral

- **Wiring Complexity** - Dependencies are wired in `root.go` (centralized, explicit)

## Alternatives Considered

### 1. Keep Direct Instantiation

**Rejected:** Violates clean architecture, makes testing impossible, inconsistent with generated code.

### 2. Global Singletons

**Rejected:** Hidden dependencies, difficult to test, state management issues.

### 3. Functional Options Pattern

**Rejected:** Overly complex for CLI commands, makes dependencies less explicit.

## Implementation Notes

- All new commands MUST use this pattern
- Existing commands will be migrated in Epic 0.5 Phase 2 (tasks #161-166)
- `root.go` is responsible for dependency wiring
- Architecture tests will enforce this pattern (Epic 0.5 Phase 6)

## References

- [Epic 0.5: Architecture Alignment](../roadmap/phases/0-foundation/epics/0.5-architecture-alignment.md)
- [CLAUDE.md - CLI Tool Architecture](../../CLAUDE.md#cli-tool-architecture-tracks-itself)
