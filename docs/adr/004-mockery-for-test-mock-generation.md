# ADR-004: Mockery for Test Mock Generation

**Status:** Accepted
**Date:** 2025-10-29
**Context:** Epic 0.5 - Architecture Alignment

## Context

After implementing ADR-001 (Dependency Injection) and ADR-002 (Interface Placement), we have multiple interfaces that need to be mocked for testing:

- `interfaces.Renderer` - Used by all commands for output
- `interfaces.BuildInfo` - Used by version command
- `interfaces.Progress` - Used by renderer for progress tracking
- Future: `interfaces.Validator`, `interfaces.ProjectGenerator`, etc.

**Problem:**

Currently, we have hand-rolled mocks duplicated across test files:

```go
// internal/cli/commands/new_test.go
type mockRenderer struct {
    titleCalls   []string
    sectionCalls []interfaces.Section
    flushed      bool
}
func (m *mockRenderer) Title(text string) { ... }
// ... more methods

// internal/cli/renderer/renderer_test.go
type mockRenderer struct{}
func (m *mockRenderer) Title(s string) {}
// ... different implementation
```

This creates maintenance burden:

- **Duplication** - Same interface mocked in multiple places
- **Inconsistency** - Different mock implementations for same interface
- **Manual Updates** - Interface changes require updating multiple mocks
- **Not Eating Our Own Dogfood** - Tracks generates apps with mockery, but doesn't use it itself

## Decision

We will use **Mockery** to automatically generate test mocks from interfaces:

1. **Add Mockery as Tool** - Use Go 1.25+ tool directive in `go.mod`
2. **Package-Based Discovery** - Configure mockery to discover all interfaces in `interfaces/` packages
3. **Convention Over Configuration** - Auto-generate mocks for all interfaces, no manual registration needed
4. **Generated Mocks Package** - Mocks live in `<package>/mocks/` subdirectories
5. **Testify Integration** - Use `github.com/stretchr/testify/mock` for mock assertions

**Configuration:**

```yaml
# .mockery.yaml
with-expecter: true
dir: "{{.InterfaceDir}}/mocks"
outpkg: "{{.PackageName}}"
filename: "mock_{{.InterfaceName}}.go"
packages:
  github.com/anomalousventures/tracks/internal/cli/interfaces:
    interfaces:
      # Auto-discover all interfaces
  github.com/anomalousventures/tracks/internal/generator/interfaces:
    interfaces:
      # Auto-discover all interfaces (future)
```

**Usage:**

```go
// Test with generated mock
import (
    "github.com/anomalousventures/tracks/internal/cli/interfaces/mocks"
)

func TestNewCommand_Run(t *testing.T) {
    mockRenderer := mocks.NewMockRenderer(t)
    mockRenderer.EXPECT().Title("Creating new Tracks application: myapp").Once()
    mockRenderer.EXPECT().Section(mock.Anything).Once()
    mockRenderer.EXPECT().Flush().Return(nil).Once()

    // Test command with mock
}
```

## Consequences

### Positive

- **Consistency** - Single source of truth for mocks
- **Maintainability** - Interface changes automatically update mocks
- **Eating Our Own Dogfood** - Tracks uses the same tools it generates for users
- **Type Safety** - Generated mocks are type-checked at compile time
- **Better Assertions** - Testify mock provides powerful expectation system
- **Less Boilerplate** - No manual mock implementations needed
- **Convention Over Configuration** - Package discovery eliminates manual interface registration

### Negative

- **Build Complexity** - Adds code generation step to workflow
- **Learning Curve** - Developers must learn testify/mock API
- **Generated Code** - Adds generated files to repository (or requires generation before tests)

### Neutral

- **Tooling Dependency** - Relies on mockery being maintained
- **Test Migration** - Existing tests must be refactored to use generated mocks

## Alternatives Considered

### 1. Continue Hand-Rolling Mocks

**Rejected:** Doesn't scale as interfaces grow, creates duplication, inconsistent with generated apps.

### 2. gomock (Official Go Mocking)

**Rejected:** More verbose, requires reflection mode or source mode, less idiomatic than testify.

### 3. Manual Interface Registration in Config

**Rejected:** Requires updating config every time an interface is added, violates DRY principle.

**Our Approach:** Use package-based discovery - mockery automatically finds all interfaces in configured packages.

## Implementation Notes

### Makefile Integration

```makefile
.PHONY: generate-mocks
generate-mocks: ## Generate test mocks from interfaces
	@echo "Generating mocks..."
	go tool mockery
```

### Pre-commit Hook (Optional)

Consider adding to `.githooks/pre-commit`:

```bash
# Ensure mocks are up-to-date
make generate-mocks
git add internal/*/interfaces/mocks/
```


### Migration Strategy

1. **Add Dependencies** - mockery + testify to go.mod
2. **Create Config** - .mockery.yaml with package discovery
3. **Generate Mocks** - Run `make generate-mocks`
4. **Refactor Tests** - Replace hand-rolled mocks with generated ones
5. **Remove Old Mocks** - Delete manual mock implementations

### Package Discovery Pattern

Mockery will discover interfaces in packages matching:

- `internal/cli/interfaces/`
- `internal/generator/interfaces/` (future)
- Any package named `interfaces/`

No need to register individual interfaces - adding a new interface automatically generates its mock.

## References

- [Mockery Documentation](https://vektra.github.io/mockery/)
- [Testify Mock Documentation](https://pkg.go.dev/github.com/stretchr/testify/mock)
- [ADR-001: Dependency Injection](./001-dependency-injection-for-cli-commands.md)
- [ADR-002: Interface Placement](./002-interface-placement-consumer-packages.md)
- [Epic 0.5: Architecture Alignment](../roadmap/phases/0-foundation/epics/0.5-architecture-alignment.md)
