# Phase 2: Data Layer

[← Back to Roadmap](../README.md) | [← Phase 1](./1-core-web.md) | [Phase 3 →](./3-auth-system.md)

## Overview

This phase implements the database layer with SQLC for type-safe queries, Goose for migrations, and support for multiple database drivers. Critical foundation for all data-driven features.

**Target Version:** v0.4.0
**Estimated Effort:** 135 tasks across 5 epics
**Status:** Not Started

## Goals

- Multi-database driver support (go-libsql, sqlite3, postgres)
- SQLC for type-safe SQL queries
- Goose migration system with embedded migrations
- UUIDv7 identifiers and slug generation utilities
- Database CLI commands for migration management

## Epic Structure

```text
Epic 2.1: Identifier Utilities (UUIDv7 & Slugs)  [Foundation - No deps]
    │
Epic 2.2: Goose Migration System                 [Needs 2.1 for UUIDs in schema]
    │
    ├──► Epic 2.3: SQLC Query Enhancement        [Needs 2.2 for schema]
    │
    └──► Epic 2.4: Database CLI Commands         [Needs 2.2 for migrations]
         │
         └──► Epic 2.5: Audit, Integration & Docs [Needs 2.1-2.4 complete]
```

## Features

### 2.1 Identifier Utilities (UUIDv7 & Slugs)

**Status:** Not Started

**Description:** Provide ready-to-use identifier generation and validation utilities

**Epic Document:** [2.1 Identifier Utilities](./2-data-layer/epics/2.1-identifier-utilities.md)

**Acceptance Criteria:**

- [ ] `identifier.NewID()` returns valid UUIDv7 (RFC 9562)
- [ ] `identifier.ValidateID(id)` returns error for invalid UUIDs
- [ ] `identifier.ExtractTimestamp(id)` returns creation time
- [ ] `slug.Generate()` returns URL-safe 8-char string
- [ ] `slug.Sanitize(input)` returns clean slug
- [ ] `slug.ValidateUsername(username)` enforces rules
- [ ] Dependencies added to go.mod template
- [ ] Generated project tests pass

**PRD Reference:** [Database Layer - UUID Implementation](../../prd/2_database_layer.md#uuid-implementation-uuidv7)

**Task Estimate:** 18 tasks

### 2.2 Goose Migration System

**Status:** Not Started

**Description:** Database migration infrastructure with driver-specific schemas

**Epic Document:** [2.2 Goose Migrations](./2-data-layer/epics/2.2-goose-migrations.md)

**Acceptance Criteria:**

- [ ] Migration directory structure per driver
- [ ] Migration runner with embedded FS
- [ ] Initial schema migration template
- [ ] `make migrate-up` runs all pending migrations
- [ ] `make migrate-down` rolls back last migration
- [ ] `make migrate-status` shows applied/pending
- [ ] `make migrate-create NAME=xxx` creates timestamped file
- [ ] Migration up/down tested for each driver

**PRD Reference:** [Database Layer - Migration System](../../prd/2_database_layer.md#migration-system)

**Task Estimate:** 32 tasks

### 2.3 SQLC Query Enhancement

**Status:** Not Started

**Description:** Enhance SQLC setup with proper type mappings and query patterns

**Epic Document:** [2.3 SQLC Enhancement](./2-data-layer/epics/2.3-sqlc-enhancement.md)

**Acceptance Criteria:**

- [ ] sqlc.yaml correctly maps driver to SQLC engine
- [ ] Health check query works with all 3 drivers
- [ ] `make generate` produces valid Go code
- [ ] Generated code compiles without errors
- [ ] SQLC generation is idempotent
- [ ] Parameter syntax correct per driver

**PRD Reference:** [Database Layer - SQLC Configuration](../../prd/2_database_layer.md#sqlc-configuration)

**Task Estimate:** 25 tasks

### 2.4 Database CLI Commands

**Status:** Not Started

**Description:** Add `tracks db` subcommands for database management

**Epic Document:** [2.4 Database CLI](./2-data-layer/epics/2.4-database-cli.md)

**Acceptance Criteria:**

- [ ] `tracks db migrate` runs from project directory
- [ ] `tracks db rollback` undoes last migration
- [ ] `tracks db status` shows correct state
- [ ] `tracks db reset --force` drops and recreates
- [ ] Commands read DATABASE_URL from .env
- [ ] Works with all 3 database drivers

**PRD Reference:** [Database Layer - Migration Commands](../../prd/2_database_layer.md#migration-system)

**Task Estimate:** 28 tasks

### 2.5 Audit, Integration & Documentation

**Status:** Not Started

**Description:** Audit existing templates, verify integration, complete documentation

**Epic Document:** [2.5 Audit & Documentation](./2-data-layer/epics/2.5-audit-documentation.md)

**Acceptance Criteria:**

- [ ] All existing templates audited against PRD
- [ ] Database setup guide complete
- [ ] Migration guide with examples
- [ ] SQLC query patterns documented
- [ ] Generated README includes database section
- [ ] Driver-specific gotchas documented

**PRD Reference:** [Database Layer](../../prd/2_database_layer.md)

**Task Estimate:** 32 tasks

## Dependencies

### Prerequisites

- Phase 0 & 1 completed
- Database drivers installed

### External Dependencies

- github.com/pressly/goose/v3
- sqlc
- github.com/gofrs/uuid/v5
- github.com/jaevor/go-nanoid
- Database drivers (libsql, sqlite3, pgx)

### Internal Dependencies

- Web layer for database endpoints

## Success Criteria

1. Can create and run migrations for all 3 drivers
2. SQLC generates correct Go code for all 3 drivers
3. Database connections work for all drivers
4. UUIDv7 IDs generated correctly
5. Slug utilities work correctly
6. `tracks db` commands work from project directories

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Driver incompatibilities | High | Test all drivers thoroughly in CI |
| Migration failures | High | Always test rollbacks |
| SQLC driver differences | Medium | Driver-specific query syntax |

## Testing Requirements

- Migration up/down tests for each driver
- SQLC generated code tests
- Connection pool tests
- UUID generation tests
- Slug generation tests
- Cross-database compatibility tests
- CLI command integration tests

## Documentation Requirements

- Database setup guide
- Migration writing guide
- SQLC query patterns guide
- Driver selection guide
- Troubleshooting guide

## Future Considerations

Features that depend on this phase:

- Authentication (user tables)
- Authorization (permission tables)
- All business logic needing persistence
- Code generation for resources (Phase 4)

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| 2025-11-28 | Created epic structure with 5 epics | Phase 2 planning complete |
| 2025-11-28 | Merged Database Drivers into other epics | Driver support distributed across relevant epics |
| 2025-11-28 | Added Database CLI as Epic 2.4 | PRD requires `tracks db` commands |

## Notes

- Database layer is critical - test extensively
- Keep SQL simple and portable where possible
- Document driver-specific SQL clearly
- UUIDv7 is mandatory, not UUIDv4
- Minimal generated projects (health check only)
- Full domain generation via Phase 4 (`tracks generate resource`)

## Next Phase

[Phase 3: Auth System →](./3-auth-system.md)
