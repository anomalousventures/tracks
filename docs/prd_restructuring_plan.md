# PRD Restructuring and Fix Tracking Document

## Status: ✅ Complete

**Created:** 2025-01-16
**Last Updated:** 2025-01-17
**Completion:** 100% - All 20 files created

## Overview

This document tracks the restructuring of the monolithic `technical_prd.md` file into modular, LLM-friendly documents, while applying critical fixes identified during the review.

## Progress Tracker

| Step | File | Status | Lines Extracted | Fixes Applied | Notes |
|------|------|--------|----------------|---------------|-------|
| 1 | Directory Structure | ✅ Complete | - | - | Created `docs/prd/` |
| 2 | `0_summary.md` | ✅ Complete | 1-53 | Framework/User choices added | |
| 3 | `1_core_architecture.md` | ✅ Complete | 37-133 | CGO requirements added | |
| 4 | `2_database_layer.md` | ✅ Complete | 135-385 | UUIDv7, DB-specific SQL | **CRITICAL FIXES APPLIED** |
| 5 | `3_authentication.md` | ✅ Complete | 387-602 | UUIDv7, complete imports | |
| 6 | `4_authorization_rbac.md` | ✅ Complete | 604-1363 | UUIDv7 fixes applied | |
| 7 | `5_web_layer.md` | ✅ Complete | 1403-1451, 2784-3058 | Middleware ordering critical fix noted | |
| 8 | `6_security.md` | ✅ Complete | 1453-1541 | File validation 261→512 bytes fixed | **CRITICAL FIX APPLIED** |
| 9 | `7_templates_assets.md` | ✅ Complete | 1543-1650, 1856-1998, 2700-2770 | All sections merged | |
| 10 | `8_external_services.md` | ✅ Complete | 1652-1842 | Circuit breaker patterns documented | |
| 11 | `9_configuration.md` | ✅ Complete | 3060-3406 | Framework vs User choices clarified | |
| 12 | `10_background_jobs.md` | ✅ Complete | 3408-3742 | UUIDv7 fixes applied | |
| 13 | `11_storage.md` | ✅ Complete | 2327-2427, 3744-3826 | Duplicates merged, 261→512 fix applied | **CRITICAL FIX APPLIED** |
| 14 | `12_observability.md` | ✅ Complete | 2079-2244 | OpenTelemetry setup documented | |
| 15 | `13_testing.md` | ✅ Complete | 3829-3926 | Comprehensive test framework documented | |
| 16 | `14_code_generation.md` | ✅ Complete | 2429-2698 | UUIDv7 fixes, TUI generation documented | |
| 17 | `15_mcp_server.md` | ✅ Complete | 3928-3996 | MCP tools expanded, UUIDv7 applied | |
| 18 | `16_tui_mode.md` | ✅ Complete | 3998-4098 | Interactive TUI with all screens documented | |
| 19 | `17_deployment.md` | ✅ Complete | NEW + 2247-2325 | Graceful shutdown, Docker, K8s, monitoring | **NEW FILE** |
| 20 | `18_dependencies.md` | ✅ Complete | NEW | Complete go.mod with all deps documented | **NEW FILE** |

Legend: ⬜ Not Started | 🟨 In Progress | ✅ Complete | ❌ Blocked

## Critical Fixes to Apply

### 🔴 P0 - Critical (Must Fix)

#### 1. Database Compatibility (Affects: `2_database_layer.md`)

- [ ] Separate SQL examples for each database type
- [ ] Remove PostgreSQL-specific syntax from "universal" examples
- [ ] Fix trigger syntax for each database
- [ ] Add clear compatibility matrix

#### 2. UUID Implementation (Affects: Multiple files)

- [x] Replace all `github.com/google/uuid` with `github.com/gofrs/uuid/v5` (✅ Done in 2_database_layer.md)
- [x] Update all examples to use UUIDv7: `uuid.Must(uuid.NewV7())` (✅ Done in 2_database_layer.md)
- [x] Document benefits of UUIDv7 over UUIDv4 (✅ Done in 2_database_layer.md)

#### 3. Storage Unification (Affects: `11_storage.md`)

- [ ] Merge duplicate implementations from lines 2327-2427 and 3744-3826
- [ ] Use single Storage interface throughout
- [ ] Remove redundant code

### 🟡 P1 - Important

#### 4. Security Improvements (Affects: `6_security.md`, `5_web_layer.md`)

- [ ] Change file validation from 261 to 512 bytes
- [ ] Add WARNING box for CSP nonce middleware ordering
- [ ] Add complete security checklist

#### 5. Configuration Clarity (Affects: `0_summary.md`, `9_configuration.md`)

- [x] Create clear Framework vs User choices section (✅ Done in 0_summary.md)
- [ ] Add complete .env.example
- [ ] Document Viper as app config tool, not framework restriction

### 🟢 P2 - Nice to Have

#### 6. Missing Content

- [ ] Add complete go.mod example (`18_dependencies.md`)
- [ ] Add deployment strategies (`17_deployment.md`)
- [ ] Add performance benchmarks
- [ ] Add troubleshooting guides

## File-Specific Fix Details

### `0_summary.md`

**Create from:** Lines 1-53 of original
**Additional content to add:**

```markdown
## Framework Choices (Built-in, Non-configurable)
- Router: Chi
- Templates: templ
- Database queries: SQLC
- Configuration: Viper (for parsing only)
- Authorization: Casbin
- Testing: testify

## User Choices (Configurable)
- Database driver: go-libsql (default), sqlite3, postgres
- Email provider: mailpit, ses, sendgrid, smtp
- SMS provider: log, sns, twilio
- Storage provider: local, s3, r2
- Queue provider: memory, sqs, pubsub
```

### `2_database_layer.md`

**Critical fixes:**

**Replace UUID code:**

```go
// OLD (WRONG)
import "github.com/google/uuid"
user.ID = uuid.New().String()

// NEW (CORRECT)
import "github.com/gofrs/uuid/v5"
user.ID = uuid.Must(uuid.NewV7()).String()
```

**Split SQL examples:**

```sql
-- Section: SQLite/go-libsql Examples
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Section: PostgreSQL Examples
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### `6_security.md`

**Critical fix:**

```go
// OLD (WRONG)
head := make([]byte, 261)

// NEW (CORRECT)
head := make([]byte, 512)  // 512 bytes for complete magic number detection
```

**Add warning:**

```markdown
⚠️ **CRITICAL: Middleware Order**
The nonce middleware MUST be registered before SecureHeaders:
```

### `11_storage.md`

**Unification approach:**

- Keep interface from lines 3765-3779
- Keep S3Storage from lines 3785-3826
- Merge upload handler from lines 2380-2426
- Add virus scanning hook point

### `18_dependencies.md` (NEW)

**Complete go.mod to create:**

```go
module github.com/user/myapp

go 1.23

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/a-h/templ v0.2.513
    github.com/spf13/viper v1.18.2
    github.com/spf13/cobra v1.8.0
    github.com/casbin/casbin/v2 v2.82.0
    github.com/gofrs/uuid/v5 v5.0.0  // UUIDv7 support
    github.com/alexedwards/scs/v2 v2.7.0
    github.com/tursodatabase/libsql-client-go v0.0.0-20240416075003-747366ff79c4
    github.com/lib/pq v1.10.9  // PostgreSQL
    github.com/mattn/go-sqlite3 v1.14.19  // Alternative SQLite
    github.com/jaevor/go-nanoid v1.3.0
    github.com/benbjohnson/hashfs v0.2.1
    github.com/sony/gobreaker v0.5.0
    github.com/markbates/goth v1.78.0  // OAuth
    github.com/gorilla/sessions v1.2.2
    github.com/go-playground/validator/v10 v10.16.0
    github.com/microcosm-cc/bluemonday v1.0.26
    github.com/invopop/ctxi18n v0.8.1
    github.com/rs/zerolog v1.31.0
    github.com/go-chi/httprate v0.8.0
    github.com/riandyrn/otelchi v0.5.1
    github.com/XSAM/otelsql v0.27.0
    github.com/prometheus/client_golang v1.18.0
    github.com/aws/aws-sdk-go-v2 v1.24.1
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/stretchr/testify v1.8.4
)
```

## Validation Checklist

### Pre-Extraction

- [ ] Original PRD backed up
- [ ] All team members notified

### During Extraction

- [ ] Each section under 1000 lines
- [ ] All code examples tested for compilation
- [ ] Cross-references updated to relative links
- [ ] Navigation footer added to each doc

### Post-Extraction

- [ ] All links in summary work
- [ ] No duplicate content
- [ ] Database examples work for all drivers
- [ ] Original PRD archived
- [ ] README updated

## Command Log

```bash
# 2025-01-16 13:28
mkdir -p /home/ashmortar/anomalousventures/tracks/docs/prd
ls -la /home/ashmortar/anomalousventures/tracks/docs/prd  # Verified creation
```

## Issues Encountered

| Date | Issue | Resolution | Impact |
|------|-------|------------|--------|
| | | | |

## Notes

- Keep this document updated after EVERY file creation/modification
- Mark items with ✅ only when fully complete and validated
- Use 🟨 for work in progress
- Add notes about any deviations from plan

## Next Action

**Current Step:** Create `0_summary.md` with framework/user choices clarification
**Blocked by:** Nothing
**Ready to proceed:** Yes
