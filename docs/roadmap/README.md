# Tracks Framework Roadmap

## Overview

This roadmap outlines the development phases for the Tracks framework across **7 phases (Phase 0 through Phase 6)**. It is designed to be **flexible and adjustable** as we learn during implementation. Each phase builds on the previous ones, but the specifics can be revised based on discoveries and changing priorities.

**Last Updated:** 2025-10-17
**Roadmap Version:** 1.0.0

## Quick Navigation

- [Phase 0: Foundation](./phases/0-foundation.md) - CLI & Project Scaffolding
- [Phase 1: Core Web](./phases/1-core-web.md) - Router, Templates, Assets
- [Phase 2: Data Layer](./phases/2-data-layer.md) - Database, SQLC, Migrations
- [Phase 3: Auth System](./phases/3-auth-system.md) - Sessions, Authentication, RBAC
- [Phase 4: Generation](./phases/4-generation.md) - Code Generators & Basic TUI
- [Phase 5: Production](./phases/5-production.md) - Observability, Testing, Deployment
- [Phase 6: Advanced](./phases/6-advanced.md) - MCP Server, Advanced TUI, Jobs
- [How to Adjust](./ADJUSTING.md) - Guidelines for updating this roadmap

## Progress Tracking

Track progress across all 7 phases below.

### Phase 0: Foundation - ✅ Complete

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 0.1 | CLI with Cobra | Complete | 2025-10-21 | 2025-11-13 | [Core Architecture](../prd/1_core_architecture.md#cli-tool-structure) | Released in v0.2.0 |
| 0.2 | Project scaffolding | Complete | 2025-10-21 | 2025-11-13 | [Core Architecture](../prd/1_core_architecture.md#generated-application-structure) | Released in v0.2.0 |
| 0.3 | Build system | Complete | 2025-10-21 | 2025-11-13 | [Core Architecture](../prd/1_core_architecture.md#build-system) | Released in v0.2.0 |
| 0.4 | Basic documentation | Complete | 2025-10-21 | 2025-11-13 | - | Released in v0.2.0 |

### Phase 1: Core Web - 🚧 In Progress

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 1.1 | Chi router setup | Not Started | - | - | [Web Layer](../prd/5_web_layer.md#router-setup) | - |
| 1.2 | templ templates | Not Started | - | - | [Templates & Assets](../prd/7_templates_assets.md#template-system) | - |
| 1.3 | hashfs assets | Not Started | - | - | [Templates & Assets](../prd/7_templates_assets.md#hashfs-integration) | - |
| 1.4 | Middleware stack | Not Started | - | - | [Web Layer](../prd/5_web_layer.md#middleware-stack) | - |

### Phase 2: Data Layer - Not Started

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 2.1 | Database drivers | Not Started | - | - | [Database Layer](../prd/2_database_layer.md#database-driver-support) | - |
| 2.2 | SQLC setup | Not Started | - | - | [Database Layer](../prd/2_database_layer.md#sqlc-configuration) | - |
| 2.3 | Goose migrations | Not Started | - | - | [Database Layer](../prd/2_database_layer.md#migration-system) | - |
| 2.4 | UUIDv7 & slugs | Not Started | - | - | [Database Layer](../prd/2_database_layer.md#uuid-implementation-uuidv7) | - |

### Phase 3: Auth System - Not Started

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 3.1 | Session management | Not Started | - | - | [Authentication](../prd/3_authentication.md#session-management) | - |
| 3.2 | OTP/Magic links | Not Started | - | - | [Authentication](../prd/3_authentication.md#1-otp-one-time-password---default) | - |
| 3.3 | OAuth2 providers | Not Started | - | - | [Authentication](../prd/3_authentication.md#3-oauth2-providers) | - |
| 3.4 | Casbin RBAC | Not Started | - | - | [Authorization & RBAC](../prd/4_authorization_rbac.md) | - |

### Phase 4: Generation - Not Started

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 4.1 | Resource generators | Not Started | - | - | [Code Generation](../prd/14_code_generation.md#interactive-tui-generators) | - |
| 4.2 | Service generators | Not Started | - | - | [Code Generation](../prd/14_code_generation.md#service-layer-generation) | - |
| 4.3 | Basic TUI | Not Started | - | - | [TUI Mode](../prd/16_tui_mode.md#tui-architecture) | - |
| 4.4 | Hot reload | Not Started | - | - | [Core Architecture](../prd/1_core_architecture.md#development-build) | - |

### Phase 5: Production - Not Started

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 5.1 | OpenTelemetry | Not Started | - | - | [Observability](../prd/12_observability.md) | - |
| 5.2 | Testing framework | Not Started | - | - | [Testing](../prd/13_testing.md) | - |
| 5.3 | Security headers | Not Started | - | - | [Security](../prd/6_security.md) | - |
| 5.4 | Service adapters | Not Started | - | - | [External Services](../prd/8_external_services.md) | - |
| 5.5 | Deployment configs | Not Started | - | - | [Deployment](../prd/17_deployment.md) | - |

### Phase 6: Advanced - Not Started

| ID  | Feature | Status | Started | Completed | PRD Link | Notes |
|-----|---------|--------|---------|-----------|----------|-------|
| 6.1 | MCP server | Not Started | - | - | [MCP Server](../prd/15_mcp_server.md) | - |
| 6.2 | Advanced TUI | Not Started | - | - | [TUI Mode](../prd/16_tui_mode.md#tui-screens) | - |
| 6.3 | Background jobs | Not Started | - | - | [Background Jobs](../prd/10_background_jobs.md) | - |
| 6.4 | i18n support | Not Started | - | - | [Templates & Assets](../prd/7_templates_assets.md#internationalization) | - |
| 6.5 | Storage adapters | Not Started | - | - | [Storage](../prd/11_storage.md) | - |

## Status Definitions

- **Not Started** - Feature not yet begun
- **In Progress** - Active development
- **Complete** - Feature implemented and tested
- **Blocked** - Cannot proceed due to dependency or issue
- **Revised** - Plan changed based on learnings

## Dependencies

### Critical Path

```text
Foundation → Core Web → Data Layer → Auth System → Generation
                            ↓
                     Production Ready
                            ↓
                    Advanced Features
```

### Feature Dependencies

- **Auth System** requires: Data Layer (for user tables), Core Web (for sessions)
- **Code Generation** requires: Data Layer (SQLC), Templates (templ)
- **RBAC** requires: Authentication (users must exist)
- **MCP Server** requires: Most other features to be implemented
- **Background Jobs** requires: Services to queue jobs

## Adjustments & Revisions

### Version History

| Version | Date | Changes | Reason |
|---------|------|---------|--------|
| 1.0.0 | 2025-10-17 | Initial roadmap | Project inception |

### Planned Review Points

- After each phase completion
- When blockers are encountered
- Monthly progress review

## Notes

- Documentation and example apps evolve continuously with each feature
- Docusaurus versions will align with minor releases (0.5, 0.7, 1.0)
- Each feature includes its documentation as part of "done"
- This roadmap is a living document - see [ADJUSTING.md](./ADJUSTING.md) for update guidelines

## Success Metrics

Each phase has specific success criteria defined in its detailed documentation. Generally:

1. **Phase Success** = All features implemented, tested, and documented
2. **Feature Success** = Works as specified in PRD, has tests, has documentation
3. **Overall Success** = Developers can build production apps with Tracks

## Next Steps

1. Review [Phase 0: Foundation](./phases/0-foundation.md) details
2. Set up development environment per [CONTRIBUTING.md](../../CONTRIBUTING.md)
3. Begin implementation of Phase 0 features
