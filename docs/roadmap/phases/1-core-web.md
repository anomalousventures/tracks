# Phase 1: Core Web

[← Back to Roadmap](../README.md) | [← Phase 0](./0-foundation.md) | [Phase 2 →](./2-data-layer.md)

## Overview

This phase establishes the web foundation with Chi router, templ templates, and asset management. The goal is to have a working HTTP server with type-safe templates and optimized asset serving.

**Target Version:** v0.3.0
**Estimated Duration:** 2-3 weeks
**Status:** Complete ✅ (All epics finished)

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

- [x] Chi router initialized in generated apps
- [x] Basic route registration
- [x] Middleware chain support
- [x] Graceful shutdown handling

**PRD Reference:** [Web Layer - Router Setup](../../prd/5_web_layer.md#router-setup)

**Implementation Notes:**

- Use chi v5
- Set up proper middleware ordering
- Implement route constants

### 1.2 templ Templates

**Description:** Set up type-safe template system with templ

**Acceptance Criteria:**

- [x] templ installed and configured
- [x] Base layout templates (internal/http/views/layouts/)
- [x] Component directory structure (internal/http/views/components/)
- [x] Template generation added to `make generate`

**PRD Reference:** [Templates & Assets - Template System](../../prd/7_templates_assets.md#template-system)

**Implementation Notes:**

- Create layouts, pages, components directories
- Add templ generation to `make generate` target (alongside sqlc and mockery)
- Create helper functions for common patterns

### 1.3 Asset Build & Serving Pipeline (merged 1.3+1.4)

**Status:** Complete

**Description:** Complete asset pipeline with TailwindCSS, JavaScript bundling, HTMX v2, and content-addressed serving

**Completed:**

- [x] Web/ directory structure created
- [x] Basic assets.go template with embed.FS
- [x] Static file handler template
- [x] MIME type handling
- [x] .gitignore template for asset pipeline
- [x] TailwindCSS v4 configuration and compilation
- [x] JavaScript bundling with esbuild
- [x] HTMX v2 with extensions (head-support, idiomorph, response-targets)
- [x] Counter component example
- [x] Content-addressed serving with hashfs
- [x] Asset compression and caching
- [x] Air live reload for assets
- [x] ProjectGenerator integration
- [x] Comprehensive testing

**Epic Document:** [1.3 Asset Pipeline](./epics/1.3-asset-pipeline.md) (62 total tasks)

**Deprecated Epics:** Original Epic 1.3 (HashFS) and Epic 1.4 (Web Build Pipeline) were merged on 2025-11-19 to eliminate artificial boundaries. See [1.3-hashfs-assets.md](./epics/1.3-hashfs-assets.md) and [1.4-web-build-pipeline.md](./epics/1.4-web-build-pipeline.md) for historical reference.

**PRD Reference:** [Templates & Assets](../../prd/7_templates_assets.md)

**Implementation Notes:**

- Phase 1 provides basic static file serving (no hashfs yet)
- Future phases add TailwindCSS, esbuild, HTMX v2, and hashfs
- Air configuration will watch .templ, .css, .js files (Phase 7)
- All builds use minification (no dev/prod mode split)

### 1.5 Templ-UI Integration

**Description:** Install and configure templ-ui component library

**Decision:** [ADR-009: Templ-UI for UI Components](../../adr/009-templui-for-ui-components.md)

**Acceptance Criteria:**

- [x] templui CLI installed as tool dependency
- [x] templui init runs during project generation
- [x] Full starter set components installed
- [x] tracks ui add and tracks ui list commands implemented
- [x] Example pages using templ-ui components
- [x] Component customization documented

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

- [x] Request ID generation
- [x] Logging middleware
- [x] Recovery middleware
- [x] Compression middleware
- [x] Static file serving
- [x] Timeout middleware
- [x] Throttle middleware
- [x] Security headers (X-Frame-Options, X-Content-Type-Options, CSP, etc.)
- [x] CORS middleware (configurable)

**Documentation:**

- [x] Router setup guide
- [x] Template component guide
- [x] Middleware documentation
- [x] Asset management guide
- [x] Templ-UI customization workflow

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
8. Dark mode theme switching works (via Templ-UI components and CSS classes)
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
