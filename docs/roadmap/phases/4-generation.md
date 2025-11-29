# Phase 4: Generation

[← Back to Roadmap](../README.md) | [← Phase 3](./3-auth-system.md) | [Phase 5 →](./5-production.md)

## Overview

This phase implements code generation capabilities and the basic TUI interface. Developers can generate complete CRUD resources, services, and handlers through both CLI commands and interactive TUI.

**Target Version:** v0.6.0
**Estimated Duration:** 4-5 weeks
**Status:** Not Started

## Goals

- Complete resource scaffolding
- Service and handler generators
- Interactive TUI with Bubble Tea
- Hot reload development experience
- Generated code is idiomatic Go

## Features

### 4.1 Resource Generators

**Description:** Generate complete CRUD resources

**Acceptance Criteria:**

- [ ] Migration generation
- [ ] SQLC query generation
- [ ] Repository creation
- [ ] Service layer generation
- [ ] Handler generation
- [ ] Template generation
- [ ] Route registration

**PRD Reference:** [Code Generation - Interactive TUI Generators](../../prd/14_code_generation.md#interactive-tui-generators)

**Implementation Notes:**

- Support field types and relations
- Generate tests by default
- Use consistent naming patterns
- No runtime magic

### 4.2 Service Generators

**Description:** Generate service layer with dependency injection

**Acceptance Criteria:**

- [ ] Interface definitions
- [ ] Constructor with DI
- [ ] Business logic methods
- [ ] Error handling
- [ ] Test generation with mocks

**PRD Reference:** [Code Generation - Service Layer Generation](../../prd/14_code_generation.md#service-layer-generation)

**Implementation Notes:**

- Always use interfaces
- Include context propagation
- Generate comprehensive tests
- Use UUIDv7 for IDs

### 4.3 Basic TUI

**Description:** Interactive terminal interface with Bubble Tea

**Acceptance Criteria:**

- [ ] Launches when `tracks` run without args
- [ ] Interactive forms for generation
- [ ] Field selection and validation
- [ ] Real-time preview
- [ ] Navigation between screens

**PRD Reference:** [TUI Mode - TUI Architecture](../../prd/16_tui_mode.md#tui-architecture)

**Implementation Notes:**

- Start with generation features
- Use Bubble Tea components
- Keep UI responsive
- Support keyboard navigation

### 4.4 Hot Reload

**Description:** Development server with automatic reload

**Acceptance Criteria:**

- [ ] Air configuration
- [ ] File watching
- [ ] Automatic rebuild
- [ ] Browser refresh
- [ ] Error display

**PRD Reference:** [Core Architecture - Development Build](../../prd/1_core_architecture.md#development-build)

**Implementation Notes:**

- Use Air for hot reload
- Configure ignore patterns
- Handle build errors gracefully
- Preserve session state

## Dependencies

### Prerequisites

- Phases 0-3 completed
- Templates and database for generation

### External Dependencies

- github.com/charmbracelet/bubbletea
- github.com/cosmtrek/air
- Code generation libraries

### Internal Dependencies

- Database layer for migrations
- Template system for views
- Service patterns established

## Success Criteria

1. Can generate complete CRUD in < 1 minute
2. Generated code compiles without errors
3. TUI is intuitive and responsive
4. Hot reload works reliably
5. Generated tests pass

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex generation logic | High | Start simple, iterate |
| TUI complexity | Medium | Focus on core features first |
| Generated code quality | High | Review and test extensively |

## Testing Requirements

- Generator output tests
- TUI interaction tests
- Generated code compilation tests
- Hot reload functionality tests
- Template rendering tests

## Documentation Requirements

- Generator command reference
- TUI user guide
- Generated code patterns
- Customization guide

## Future Considerations

Features that depend on this phase:

- Advanced TUI features
- MCP server code generation
- Custom generator plugins

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| - | - | No changes yet |

## Notes

- Generated code should look hand-written
- Keep generators simple and predictable
- TUI should enhance, not replace CLI
- Focus on developer productivity

## Next Phase

[Phase 5: Production →](./5-production.md)
