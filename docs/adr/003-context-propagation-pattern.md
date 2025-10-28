# ADR-003: Context Propagation Pattern for Request-Scoped Values

**Status:** Accepted
**Date:** 2025-10-27
**Context:** Epic 0.5 - Architecture Alignment

## Context

During Epic 3 implementation, we needed a consistent way to pass request-scoped values (logger, tracing, etc.) through the CLI execution chain. Some code was importing the `cli` package directly to get the logger, which created import cycles.

**Problem:**

```go
// ❌ WRONG: Direct import creates cycle
// internal/generator/validator_impl.go
import "github.com/anomalousventures/tracks/internal/cli"

func (v *validatorImpl) ValidateDirectory(path string) error {
    logger := cli.GetLogger(context.Background())  // Import cycle!
    logger.Warn().Msg("cleanup failed")
}
```

We needed a pattern to:

1. Attach request-scoped values at the root
2. Retrieve values deep in the call stack
3. Avoid import cycles
4. Follow Go context best practices

## Decision

We will use Go's standard `context.Context` for propagating request-scoped values:

1. **Attach at Root** - Logger and other request-scoped values attached in `root.go`
2. **Context as First Parameter** - Always pass `context.Context` as first parameter
3. **Context Package Functions** - Use helper functions `WithLogger(ctx, logger)` and `GetLogger(ctx)`
4. **Never Store Context** - Never store context in struct fields
5. **Cobra Integration** - Use `cmd.Context()` and `cmd.SetContext(ctx)`

**Example:**

```go
// ✅ Attach in root.go
func NewRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{ ... }

    cobra.OnInitialize(func() {
        logger := cli.NewLogger(logLevel)
        ctx := cli.WithLogger(context.Background(), logger)
        rootCmd.SetContext(ctx)
    })

    return rootCmd
}

// ✅ Retrieve in command
func (c *NewCommand) run(cmd *cobra.Command, args []string) error {
    logger := cli.GetLogger(cmd.Context())
    logger.Debug().Str("project", args[0]).Msg("creating project")
    // ...
}

// ✅ Pass to dependencies via constructor
func NewValidator(logger zerolog.Logger) *Validator {
    return &Validator{logger: logger}
}
```

## Consequences

### Positive

- **No Import Cycles** - Dependencies receive logger via constructor, no cli package import
- **Standard Go Pattern** - Uses `context.Context` as intended
- **Request Scoping** - Each command execution has its own context
- **Cancellation Support** - Context can be cancelled for timeout/interrupt handling
- **Testability** - Easy to create contexts with test values

### Negative

- **Boilerplate** - Must thread context through function calls
- **Type Safety Loss** - Context values are `interface{}`, requires type assertion

### Neutral

- **Context Guidelines** - Must follow "never store in struct" rule consistently

## Implementation Rules

### Rule 1: Always Pass Context First

```go
// ✅ CORRECT
func (v *Validator) ValidateProjectName(ctx context.Context, name string) error

// ❌ WRONG
func (v *Validator) ValidateProjectName(name string) error
```

### Rule 2: Never Store Context in Struct

```go
// ✅ CORRECT: Logger stored, received via constructor
type Validator struct {
    logger zerolog.Logger
}

// ❌ WRONG: Context stored
type Validator struct {
    ctx context.Context  // Never do this!
    // Reason: Contexts are request-scoped and should flow through the call chain.
    // Storing context in a struct breaks cancellation propagation, creates confusing
    // lifetimes (which request does this context belong to?), and violates Go best practices.
    // Pass context as first parameter instead.
}
```

### Rule 3: Attach Request-Scoped Values Only

**Attach:**

- Logger
- Trace IDs
- Request IDs
- User identity (future auth)

**Don't Attach:**

- Configuration (use dependency injection)
- Database connections (use dependency injection)
- Application state

## Alternatives Considered

### 1. Global Logger Singleton

**Rejected:** Hidden dependency, difficult to test, not request-scoped.

### 2. Explicit Logger Parameter Everywhere

**Rejected:** Too much boilerplate, doesn't scale to other request-scoped values.

### 3. Thread-Local Storage

**Rejected:** Not idiomatic in Go, goroutine-unsafe, difficult to reason about.

## Implementation Notes

- Epic 0.5 Phase 4 establishes this pattern (tasks #171-174)
- Currently, logger is passed via constructor to validator
- Future: trace IDs, request IDs for better observability
- Context propagation tests will verify pattern usage (task #174)

## References

- [Epic 0.5: Architecture Alignment](../roadmap/phases/0-foundation/epics/0.5-architecture-alignment.md)
- [Go Context Package](https://pkg.go.dev/context)
- [Go Blog: Context](https://go.dev/blog/context)
- [Context Best Practices](https://go.dev/blog/context-and-structs)
