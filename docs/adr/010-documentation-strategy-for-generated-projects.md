# ADR-010: Documentation Strategy for Generated Projects

**Status:** Accepted
**Date:** 2025-11-18
**Deciders:** Development Team
**Context:** Phase 1 - Core Web Layer

## Context

Tracks generates complete web applications including documentation (README, guides, troubleshooting). As the framework evolves, generated documentation becomes stale, users miss critical updates, and maintenance burden increases.

**Current Problems:**

1. **Documentation Drift** - Generated docs reflect version at creation time, never update
2. **User Confusion** - Users reference outdated patterns, missing new features/best practices
3. **Maintenance Burden** - Every doc change requires template updates, regeneration, user migration
4. **Mixed Ownership** - Unclear which docs are auto-generated vs user-created
5. **Traffic Dispersion** - Documentation scattered across thousands of user repos instead of centralized

**Alternatives Considered:**

1. **Comprehensive Generated Docs** ❌
   - Pros: Self-contained projects, offline access
   - Cons: Goes stale immediately, no updates, maintenance nightmare

2. **Generated Docs with Update Command** ❌
   - Pros: Users can refresh docs via `tracks update-docs`
   - Cons: Still requires regeneration, merge conflicts, complex ownership

3. **Minimal README + Links to Official Docs** ✅
   - Pros: Always current, single source of truth, drives traffic to docs site
   - Cons: Requires internet for docs, initial README must be carefully designed

4. **No README, Only Links** ❌
   - Pros: Zero stale content
   - Cons: Poor first impression, users expect basic project context

## Decision

Generated projects will contain **minimal documentation** with all comprehensive content living on the official docs site at `https://go-tracks.io/docs/`.

**README Content (Moderate Scope):**

```markdown
# {{.ProjectName}}

Web application built with [Tracks](https://go-tracks.io).

## Quick Start

\`\`\`bash
make dev          # Start development server
make test         # Run tests
make build        # Build production binary
\`\`\`

## Architecture

- **HTTP Layer**: Chi router + templ templates + HTMX
- **Database**: {{.DBDriver}} with SQLC type-safe queries
- **Sessions**: SCS session management
- **Logging**: Zerolog structured logging

See [Architecture Guide](https://go-tracks.io/docs/architecture) for details.

## Development

- [Getting Started](https://go-tracks.io/docs/getting-started)
- [Database Migrations](https://go-tracks.io/docs/database)
- [Adding Resources](https://go-tracks.io/docs/generators/resource)
- [Testing](https://go-tracks.io/docs/testing)
- [Deployment](https://go-tracks.io/docs/deployment)

## Documentation

Full documentation available at [go-tracks.io/docs](https://go-tracks.io/docs):

- [Handlers & Routing](https://go-tracks.io/docs/handlers)
- [Templates & Components](https://go-tracks.io/docs/templates)
- [Database Access](https://go-tracks.io/docs/database)
- [Configuration](https://go-tracks.io/docs/configuration)
- [Troubleshooting](https://go-tracks.io/docs/troubleshooting)

Built with Tracks v{{.TracksVersion}}
\`\`\`

**What NOT to Include:**

- ❌ No `docs/` folder in generated projects
- ❌ No component pattern documentation
- ❌ No troubleshooting guides
- ❌ No performance optimization docs
- ❌ No extensive code examples (beyond basic usage)
- ❌ No API reference documentation

**Docs Site Responsibility:**

All comprehensive documentation lives at `https://go-tracks.io/docs/`:

- Handler-to-template integration patterns
- Component composition examples
- Form validation strategies
- HTMX partial rendering patterns
- Performance optimization guides
- Troubleshooting guides
- Security best practices
- Deployment guides

## Rationale

**1. Single Source of Truth**

Centralizing documentation prevents fragmentation. Users always see current best practices, new features, and security updates.

**2. Reduced Maintenance Burden**

Updating one docs site is simpler than maintaining template documentation that generates into thousands of projects. Template changes require careful versioning, migration paths, and user communication.

**3. Improved User Experience**

Users get accurate, up-to-date documentation with search, cross-references, and community contributions. Stale generated docs create confusion and support burden.

**4. Traffic Benefits**

Driving users to the official docs site:
- Increases visibility into common issues
- Enables better analytics on popular topics
- Creates community hub for discussions
- Improves SEO for framework discovery

**5. Clear Ownership**

Generated README is "project metadata" (architecture, quick start). All "how-to" content lives on docs site. No ambiguity about what's auto-generated vs user-created.

## Consequences

**Positive:**

- **Always Current** - Users see latest docs, get immediate value from framework improvements
- **Easier Maintenance** - One docs site update reaches all users instantly
- **Better Analytics** - Track which docs users reference most, identify gaps
- **Community Building** - Centralized hub for examples, discussions, contributions
- **Clear Boundaries** - README for project context, docs site for comprehensive guides

**Negative:**

- **Internet Dependency** - Users need connection to access comprehensive docs
- **Link Rot Risk** - If docs site restructures, links break (mitigated by redirects)
- **Initial Friction** - Users must navigate away from project to learn patterns
- **Offline Development** - Reduced docs access in offline environments

**Mitigation Strategies:**

1. **Comprehensive README Quick Reference** - Include enough context for basic development (make commands, architecture overview)
2. **Stable URL Structure** - Design docs URLs to be stable across versions, use redirects when restructuring
3. **Clear Link Labels** - Every link describes what users will learn (not just "click here")
4. **Offline Fallback** - Future: `tracks docs` command to fetch/cache docs locally
5. **Version Pinning** - Link to version-appropriate docs when versioned docs become available

## Implementation

**Phase 1 Immediate Changes:**

1. **Epic 1.2 (Templ Templates)** - Close 14 issues related to generating documentation/components that templui provides:
   - Close #322 (handler-to-template integration docs)
   - Close #331 (component props pattern docs)
   - Close #332 (component composition examples)
   - Close #341 (template performance docs)
   - Close #342 (templ troubleshooting guide)
   - Close #325-#330, #334, #335, #337 (components templui provides)

2. **README Template** - Update `internal/templates/project/README.md.tmpl` with moderate-scope content (architecture, quick start, links)

3. **Docs Site Content** - Create comprehensive guides for critical topics:
   - Component patterns and composition
   - Form validation strategies
   - Troubleshooting guide
   - Template performance optimization

**Template Structure:**

```go
// internal/templates/project/README.md.tmpl
type ReadmeData struct {
    ProjectName    string
    DBDriver       string
    TracksVersion  string
    GoVersion      string
}
```

**Documentation Links:**

All links use format: `https://go-tracks.io/docs/<page>`

Current structure (no versioning yet):

- `/docs/getting-started`
- `/docs/architecture`
- `/docs/handlers`
- `/docs/templates`
- `/docs/database`
- `/docs/testing`
- `/docs/deployment`
- `/docs/troubleshooting`
- `/docs/configuration`
- `/docs/generators/resource`

When versioned docs launch, links will evolve to `/docs/v1/...` or `/latest/...`.

## References

- Tracks Docs Site: https://go-tracks.io/docs/
- ADR-009: Templ-UI for UI Components (component strategy)
- ADR-005: HTTP Layer Architecture (generated structure)
- PRD: Phase 1 Epic 1.2 (Templ Templates)
