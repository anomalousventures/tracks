# Phase 2: Data Layer

[← Back to Roadmap](../README.md) | [← Phase 1](./1-core-web.md) | [Phase 3 →](./3-auth-system.md)

## Overview

This phase implements the database layer with SQLC for type-safe queries, Goose for migrations, and support for multiple database drivers. Critical foundation for all data-driven features.

**Target Version:** v0.4.0
**Estimated Duration:** 3-4 weeks
**Status:** Not Started

## Goals

- Multi-database driver support
- SQLC for type-safe SQL
- Goose migration system
- UUIDv7 identifiers
- Repository pattern implementation

## Features

### 2.1 Database Drivers

**Description:** Support for go-libsql, sqlite3, and postgres drivers

**Acceptance Criteria:**

- [ ] Driver selection during project creation
- [ ] Connection management
- [ ] Driver-specific SQL handling
- [ ] Connection pooling configuration

**PRD Reference:** [Database Layer - Database Driver Support](../../prd/2_database_layer.md#database-driver-support)

**Implementation Notes:**

- Default to go-libsql for edge deployments
- Handle CGO requirements appropriately
- Abstract driver differences

### 2.2 SQLC Setup

**Description:** Configure SQLC for type-safe query generation

**Acceptance Criteria:**

- [ ] sqlc.yaml configuration
- [ ] Query files structure
- [ ] Generated code organization
- [ ] Type mappings configured

**PRD Reference:** [Database Layer - SQLC Configuration](../../prd/2_database_layer.md#sqlc-configuration)

**Implementation Notes:**

- Separate queries by domain
- Use consistent naming conventions
- Generate interfaces for testing

### 2.3 Goose Migrations

**Description:** Database migration system with Goose

**Acceptance Criteria:**

- [ ] Migration file structure
- [ ] Up/down migration support
- [ ] Embedded migrations
- [ ] Migration commands in CLI

**PRD Reference:** [Database Layer - Migration System](../../prd/2_database_layer.md#migration-system)

**Implementation Notes:**

- Use timestamp-based naming
- Separate migrations by driver if needed
- Include rollback capability

### 2.4 UUIDv7 & Slugs

**Description:** Implement UUIDv7 for IDs and slug generation

**Acceptance Criteria:**

- [ ] UUIDv7 generation functions
- [ ] Slug generation and sanitization
- [ ] Dual identifier pattern
- [ ] Validation helpers

**PRD Reference:** [Database Layer - UUID Implementation](../../prd/2_database_layer.md#uuid-implementation-uuidv7)

**Implementation Notes:**

- Use gofrs/uuid for UUIDv7
- Implement slug uniqueness checking
- Create helper packages

## Dependencies

### Prerequisites

- Phase 0 & 1 completed
- Database drivers installed

### External Dependencies

- github.com/pressly/goose/v3
- sqlc
- github.com/gofrs/uuid/v5
- Database drivers (libsql, sqlite3, pgx)

### Internal Dependencies

- Web layer for database endpoints

## Success Criteria

1. Can create and run migrations
2. SQLC generates correct Go code
3. Database connections work for all drivers
4. UUIDv7 IDs generated correctly
5. Repository pattern implemented

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Driver incompatibilities | High | Test all drivers thoroughly |
| Migration failures | High | Always test rollbacks |
| SQLC complexity | Medium | Start with simple queries |

## Testing Requirements

- Migration up/down tests
- SQLC generated code tests
- Connection pool tests
- UUID generation tests
- Cross-database compatibility tests

## Documentation Requirements

- Database setup guide
- Migration guide
- SQLC query writing guide
- Repository pattern docs

## Future Considerations

Features that depend on this phase:

- Authentication (user tables)
- Authorization (permission tables)
- All business logic needing persistence
- Code generation for resources

## Adjustments Log

This section tracks changes made to the original plan.

| Date | Change | Reason |
|------|--------|--------|
| - | - | No changes yet |

## Notes

- Database layer is critical - test extensively
- Keep SQL simple and portable where possible
- Document driver-specific SQL clearly
- UUIDv7 is mandatory, not UUIDv4

## Next Phase

[Phase 3: Auth System →](./3-auth-system.md)
