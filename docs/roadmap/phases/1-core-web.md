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
- Web build pipeline (TailwindCSS, HTMX v2)
- Templ-UI component library integration
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
- [ ] Base layout templates (internal/http/views/layouts/)
- [ ] Component directory structure (internal/http/views/components/)
- [ ] Template generation added to `make generate`

**PRD Reference:** [Templates & Assets - Template System](../../prd/7_templates_assets.md#template-system)

**Implementation Notes:**

- Create layouts, pages, components directories
- Add templ generation to `make generate` target (alongside sqlc and mockery)
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

### 1.4 Web Build Pipeline

**Description:** Set up TailwindCSS and HTMX v2 build process with extensions

**Acceptance Criteria:**

- [ ] TailwindCSS configuration and compilation
- [ ] HTMX v2 bundling with extensions (head-support, idiomorph, response-targets)
- [ ] Counter component integrated into homepage
- [ ] Asset build targets in Makefile
- [ ] Development vs production builds
- [ ] Air configuration watches .templ, .css, .js files and rebuilds assets

**PRD Reference:** [Templates & Assets - CSS & JavaScript](../../prd/7_templates_assets.md#css--javascript)

**Implementation Notes:**

- Configure Tailwind to scan .templ files
- Set up esbuild for JavaScript bundling
- Include theme switcher component
- Optimize for production builds
- Update .air.toml to watch frontend assets and trigger asset rebuilds
- Exclude generated outputs (`*_templ.go`, `internal/assets/dist`) from Air watch

### 1.5 Templ-UI Integration

**Description:** Install and configure templ-ui component library

**Decision:** [ADR-009: Templ-UI for UI Components](../../adr/009-templui-for-ui-components.md)

**Acceptance Criteria:**

- [ ] templui CLI installed as tool dependency
- [ ] templui init runs during project generation
- [ ] Full starter set components installed
- [ ] tracks ui add and tracks ui list commands implemented
- [ ] Example pages using templ-ui components
- [ ] Component customization documented

**PRD Reference:** [Templates & Assets - UI Component Library](../../prd/7_templates_assets.md#ui-component-library-templ-ui)

**Implementation Notes:**

- Install forms, layout, and feedback components
- Components go in internal/http/views/components/ui/
- Users own the component code (can customize freely)
- Update Tailwind config to include component styles

### 1.6 Middleware Stack & Documentation

**Description:** Implement core middleware and create comprehensive documentation

**Acceptance Criteria:**

**Middleware:**

- [ ] Request ID generation
- [ ] Logging middleware
- [ ] Recovery middleware
- [ ] Compression middleware
- [ ] Static file serving

**Documentation:**

- [ ] Router setup guide
- [ ] Template component guide
- [ ] Middleware documentation
- [ ] Asset management guide
- [ ] Templ-UI customization workflow

**PRD Reference:** [Web Layer - Middleware Stack](../../prd/5_web_layer.md#middleware-stack)

**Implementation Notes:**

- Order matters! Document the chain
- Use chi built-in middleware where possible
- Add custom middleware for app-specific needs
- Create comprehensive examples for all features

## Dependencies

### Prerequisites

- Phase 0 completed (CLI foundation)
- Go 1.25+ (required for tool directive support - see CLAUDE.md)

### External Dependencies

- github.com/go-chi/chi/v5
- github.com/a-h/templ
- github.com/benbjohnson/hashfs
- github.com/templui/templui (tool dependency)
- TailwindCSS (npm)
- esbuild (npm)
- HTMX v2 with extensions (npm: htmx.org, @htmx-org/htmx-head-support, @htmx-org/idiomorph, @htmx-org/htmx-response-targets)
- Standard middleware packages

### Internal Dependencies

- Basic project structure from Phase 0

## Success Criteria

1. Generated app starts HTTP server successfully
2. Routes respond with proper HTML
3. Templates compile without errors
4. Assets served with cache headers
5. Middleware chain works correctly
6. Templ-UI components render correctly
7. Build pipeline produces optimized CSS/JS
8. Dark mode theme switching works
9. Users can add/customize UI components

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
| 2025-11-13 | Split 1.4 into separate epics for Build Pipeline (1.4) and Middleware/Docs (1.6) | Better separation of concerns |
| 2025-11-13 | Added Epic 1.5 for Templ-UI Integration | [ADR-009](../../adr/009-templui-for-ui-components.md) - Adopting templ-ui as core UI library |
| 2025-11-13 | Updated dependencies to include templui, TailwindCSS, esbuild | Required for Epic 1.4 and 1.5 |
| 2025-11-13 | Enhanced success criteria with UI components and build pipeline | Reflects expanded scope with templ-ui |
| 2025-11-19 | Replaced Alpine.js with HTMX v2 + extensions in Epic 1.4 | Templ-UI provides component interactivity; HTMX for partial rendering. Alpine.js deferred to post-MVP |
| 2025-11-19 | Added counter example integration to homepage in Epic 1.4 | Demonstrates HTMX patterns after Epic 1.3 (assets) completes |

## Notes

- Middleware order is critical - document extensively
- Keep initial templates minimal but extensible
- Focus on developer experience for template authoring
- Asset pipeline should be transparent to developers

## Next Phase

[Phase 2: Data Layer →](./2-data-layer.md)
