# ADR-002: Interface Placement in Consumer Packages

**Status:** Accepted
**Date:** 2025-01-27
**Context:** Epic 0.5 - Architecture Alignment

## Context

During Epic 3 implementation, we discovered interfaces were defined in provider packages (e.g., `internal/generator/validator.go` defining the `Validator` interface). This created import cycles and violated Go best practices.

**Problem:**

```go
// ❌ WRONG: Provider defines interface
// internal/generator/validator.go
type Validator interface { ValidateProjectName(string) error }

// ❌ WRONG: CLI imports generator to use interface
// internal/cli/commands/new.go (future)
import "github.com/anomalousventures/tracks/internal/generator"

// Creates import cycle:
// internal/generator/validator_impl.go imports internal/cli (for logger)
// internal/cli will import internal/generator (for Validator interface)
```

This is the "interface provider" anti-pattern and prevents proper dependency inversion.

## Decision

We will follow Go best practices and place interfaces in **consumer packages**, not provider packages:

1. **Consumer Defines Interface** - The package that uses an interface owns it
2. **CLI Interfaces Package** - `internal/cli/interfaces/` contains interfaces consumed by CLI
3. **Provider Implements Interface** - Provider packages implement consumer-defined interfaces
4. **Import Direction** - Providers can import interface packages, but not vice versa

**Example:**

```go
// ✅ CORRECT: Consumer defines interface
// internal/cli/interfaces/validator.go
package interfaces

type Validator interface {
    ValidateProjectName(string) error
}

// ✅ CORRECT: Provider implements consumer's interface
// internal/validation/validator.go
package validation

import "github.com/anomalousventures/tracks/internal/cli/interfaces"

type Validator struct { logger zerolog.Logger }

// Implements interfaces.Validator
func (v *Validator) ValidateProjectName(name string) error { ... }
```

## Consequences

### Positive

- **No Import Cycles** - Clean dependency graph with unidirectional imports
- **Proper Dependency Inversion** - High-level modules don't depend on low-level modules
- **Better Testability** - Easy to create mocks from consumer's perspective
- **Go Idiomatic** - Follows Go proverb "accept interfaces, return structs"
- **Package Independence** - Providers don't need to know about all consumers

### Negative

- **Interface Discovery** - Developers must look in consumer packages for interfaces
- **Migration Effort** - Must move existing interfaces and update imports

### Neutral

- **Multiple Consumers** - If multiple packages need same interface, use shared consumer package

## Implementation Rules

### Rule 1: Interfaces Belong to Consumers

```go
// ✅ GOOD: CLI uses validator, so CLI owns interface
internal/cli/interfaces/validator.go

// ✅ GOOD: Implementation in provider package
internal/validation/validator.go
```

### Rule 2: Avoid Generic "Interfaces" Packages

```go
// ✅ BETTER: Interface in commands package if only one consumer
internal/cli/commands/interfaces.go

// ⚠️ ACCEPTABLE: Shared interfaces package if multiple CLI consumers
internal/cli/interfaces/validator.go
```

### Rule 3: Descriptive Interface Names

```go
// ✅ GOOD
type Validator interface { ... }
type ProjectGenerator interface { ... }

// ❌ BAD
type ValidatorInterface interface { ... }
type IValidator interface { ... }
```

## Alternatives Considered

### 1. Keep Interfaces in Provider Packages

**Rejected:** Creates import cycles, prevents dependency inversion, violates Go idioms.

### 2. Central "Interfaces" Package

**Rejected:** Creates god package, couples all packages together, difficult to maintain.

### 3. Duplicate Interfaces Per Consumer

**Rejected:** Unnecessary duplication, harder to maintain, confusing.

## Implementation Notes

- Epic 0.5 Phase 1 moves interfaces to `internal/cli/interfaces/` (tasks #156-158)
- Provider packages will be reorganized in Phase 3 (tasks #167-170)
- Architecture tests will enforce interface placement (task #180)

## References

- [Epic 0.5: Architecture Alignment](../roadmap/phases/0-foundation/epics/0.5-architecture-alignment.md)
- [Go Proverbs - Accept Interfaces, Return Structs](https://go-proverbs.github.io/)
- [Effective Go - Interfaces and Other Types](https://go.dev/doc/effective_go#interfaces)
