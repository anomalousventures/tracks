# Phase 1: Core Web

[← Back to Roadmap](../README.md) | [← Phase 0](./0-foundation.md) | [Phase 2 →](./2-data-layer.md)

## Overview

This phase establishes the web foundation with Chi router, templ templates, and asset management. The goal is to have a working HTTP server with type-safe templates and optimized asset serving.

**Target Version:** v0.2.0
**Estimated Duration:** 2-3 weeks
**Status:** Not Started

## Goals

- Chi router with middleware
- templ template system integration
- hashfs for asset management
- Basic middleware stack
- Hypermedia-driven routing patterns

## Features

### 1.1 Chi Router Setup

**Description:** Integrate Chi router with proper middleware chains

**Acceptance Criteria:**

- [ ] Chi router initialized in generated apps
- [ ] Basic route registration
- [ ] Middleware chain support
- [ ] Graceful shutdown handling

**PRD Reference:** [Web Layer - Router Setup](../../prd/5_web_layer.md#router-setup)

**Implementation Notes:**

- Use chi v5
- Set up proper middleware ordering
- Implement route constants

### 1.2 templ Templates

**Description:** Set up type-safe template system with templ

**Acceptance Criteria:**

- [ ] templ installed and configured
- [ ] Base layout templates
- [ ] Component structure
- [ ] Template generation command

**PRD Reference:** [Templates & Assets - Template System](../../prd/7_templates_assets.md#template-system)

**Implementation Notes:**

- Create layouts, pages, components directories
- Set up templ generate in build process
- Create helper functions for common patterns

### 1.3 hashfs Assets

**Description:** Implement content-addressed asset serving

**Acceptance Criteria:**

- [ ] Asset embedding with go:embed
- [ ] Content hashing for cache busting
- [ ] Asset helper functions
- [ ] Compression support

**PRD Reference:** [Templates & Assets - hashfs Integration](../../prd/7_templates_assets.md#hashfs-integration)

**Implementation Notes:**

- Use benbjohnson/hashfs
- Set up asset pipeline
- Configure cache headers

### 1.4 Middleware Stack

**Description:** Implement core middleware components

**Acceptance Criteria:**

- [ ] Request ID generation
- [ ] Logging middleware
- [ ] Recovery middleware
- [ ] Compression middleware
- [ ] Static file serving

**PRD Reference:** [Web Layer - Middleware Stack](../../prd/5_web_layer.md#middleware-stack)

**Implementation Notes:**

- Order matters! Document the chain
- Use chi built-in middleware where possible
- Add custom middleware for app-specific needs

## Dependencies

### Prerequisites

- Phase 0 completed (CLI foundation)
- Go 1.25+

### External Dependencies

- github.com/go-chi/chi/v5
- github.com/a-h/templ
- github.com/benbjohnson/hashfs
- Standard middleware packages

### Internal Dependencies

- Basic project structure from Phase 0

## Success Criteria

1. Generated app starts HTTP server successfully
2. Routes respond with proper HTML
3. Templates compile without errors
4. Assets served with cache headers
5. Middleware chain works correctly

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| templ learning curve | Medium | Provide good examples and docs |
| Middleware ordering | High | Clear documentation, validation |
| Asset pipeline complexity | Medium | Start simple, iterate |

## Testing Requirements

- Route registration tests
- Template rendering tests
- Middleware chain tests
- Asset serving tests
- Integration tests for full stack

## Documentation Requirements

- Router setup guide
- Template component guide
- Middleware documentation
- Asset management guide

## Future Considerations

Features that depend on this phase:

- Authentication needs sessions (middleware)
- Database handlers need routes
- Code generation needs templates
- All web features need this foundation

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| - | - | No changes yet |

## Notes

- Middleware order is critical - document extensively
- Keep initial templates minimal but extensible
- Focus on developer experience for template authoring
- Asset pipeline should be transparent to developers

## Next Phase

[Phase 2: Data Layer →](./2-data-layer.md)
