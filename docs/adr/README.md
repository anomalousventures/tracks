# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records documenting significant architectural decisions made during the development of Tracks.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision along with its context and consequences. ADRs help team members understand why certain decisions were made and provide historical context for future changes.

## ADR Format

Each ADR follows this structure:

- **Title** - Short descriptive name (e.g., "ADR-001: Dependency Injection for CLI Commands")
- **Status** - Accepted, Proposed, Deprecated, Superseded
- **Date** - When the decision was made
- **Context** - The problem or situation that led to this decision
- **Decision** - What we decided to do
- **Consequences** - Positive, negative, and neutral outcomes
- **Alternatives Considered** - Other options we evaluated
- **Implementation Notes** - How this will be implemented
- **References** - Links to related documentation

## Decision Records

### Active

- [ADR-001: Dependency Injection Pattern for CLI Commands](./001-dependency-injection-for-cli-commands.md) - Establishes DI pattern for testable, maintainable CLI commands
- [ADR-002: Interface Placement in Consumer Packages](./002-interface-placement-consumer-packages.md) - Prevents import cycles by placing interfaces in consumer packages
- [ADR-003: Context Propagation Pattern](./003-context-propagation-pattern.md) - Uses Go context for request-scoped values like logger and trace IDs
- [ADR-004: Mockery for Test Mock Generation](./004-mockery-for-test-mock-generation.md) - Uses mockery with package discovery for automated test mock generation
- [ADR-005: HTTP Layer Architecture and Cross-Domain Orchestration](./005-http-layer-architecture.md) - Establishes HTTP layer structure and handler orchestration pattern for generated applications
- [ADR-006: Route Reference Constants Pattern](./006-route-reference-constants.md) - Defines centralized route path constants to eliminate magic strings
- [ADR-007: Configuration File Separation](./007-configuration-file-separation.md) - Separates configuration concerns into dedicated config package
- [ADR-008: Generator Template Sequencing](./008-generator-template-sequencing.md) - Sequences template rendering to support mock-dependent tests (app templates → tidy → generate → test templates)

### Superseded

None yet.

### Deprecated

None yet.

## Creating a New ADR

When making a significant architectural decision:

1. Copy the template structure from existing ADRs
2. Number it sequentially (ADR-004, ADR-005, etc.)
3. Use a descriptive filename: `NNN-short-descriptive-name.md`
4. Fill in all sections thoroughly
5. Get team review before marking as "Accepted"
6. Update this index

## When to Create an ADR

Create an ADR when:

- Making a decision that affects the system's structure
- Choosing between competing architectural patterns
- Establishing a new convention or standard
- Making a decision that's difficult to reverse
- Making a decision that will impact future development

Don't create an ADR for:

- Implementation details that don't affect architecture
- Decisions that are easily reversible
- Routine code changes
- Bug fixes

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
