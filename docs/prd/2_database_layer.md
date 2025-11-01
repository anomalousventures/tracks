# Database Layer

**[← Back to Summary](./0_summary.md) | [← Core Architecture](./1_core_architecture.md) | [Authentication →](./3_authentication.md)**

## Overview

The database layer provides zero-overhead, type-safe database access through SQLC with support for multiple database drivers. Applications use UUIDv7 for internal identifiers and human-readable slugs for public URLs.

## Goals

- Zero runtime SQL errors through compile-time query validation
- Database-agnostic architecture with adapter pattern
- UUIDv7 primary keys for security and ordering, slugs for public URLs
- Transactional consistency for complex operations
- Built-in observability and performance monitoring
- Support for go-libsql (default), sqlite3, or postgres database drivers
- User selects database driver at application creation

## User Stories

- As a developer, I want compile-time SQL validation so I never ship broken queries
- As a developer, I want UUIDv7 for internal IDs to prevent enumeration attacks and maintain order
- As a developer, I want human-readable slugs for public URLs
- As a developer, I want to choose between go-libsql, sqlite3, or postgres at project creation
- As a DevOps engineer, I want database metrics exported to Prometheus
- As a developer, I want automatic query timing and slow query logging
- As a developer, I want consistent behavior between development and production databases

## Database Driver Support

Tracks supports three database drivers, chosen at project creation:

| Driver | Use Case | CGO Required | Production Ready |
|--------|----------|--------------|------------------|
| `go-libsql` | Edge deployments (Turso) | Yes | Yes |
| `sqlite3` | Single-server apps | Yes | Yes |
| `postgres` | Traditional deployments | No | Yes |

## Data Model Conventions

### Deletion Policy

**Hard delete by default.** If a feature needs recoverability, document it explicitly per-table and add `deleted_at` only where required, updating queries accordingly.

### Identifier Strategy

All resources use a dual identifier pattern:

- **UUID (v7)**: Internal references, foreign keys, API operations
- **Slug/Username**: Human-readable URLs, user-facing identifiers

## UUID Implementation (UUIDv7)

> **CRITICAL:** Use UUIDv7, not UUIDv4, for timestamp-ordering benefits

```go
// internal/pkg/identifier/uuid.go
package identifier

import (
    "github.com/gofrs/uuid/v5"  // Supports UUIDv7
)

// NewID generates a UUIDv7 with embedded timestamp
// UUIDv7 provides:
// - Timestamp ordering (better index performance)
// - Roughly sortable by creation time
// - Still cryptographically random
func NewID() string {
    return uuid.Must(uuid.NewV7()).String()
}

// Validate checks if a string is a valid UUID
func ValidateID(id string) error {
    _, err := uuid.FromString(id)
    if err != nil {
        return fmt.Errorf("invalid UUID: %w", err)
    }
    return nil
}

// Extract timestamp from UUIDv7 (useful for debugging)
func ExtractTimestamp(id string) (time.Time, error) {
    u, err := uuid.FromString(id)
    if err != nil {
        return time.Time{}, err
    }

    // UUIDv7 has timestamp in first 48 bits
    // Implementation details depend on uuid library version
    return u.Time(), nil
}
```

## Database-Specific SQL

### SQLite / go-libsql Migration

```sql
-- migrations/2024_01_15_14_30_create_users.sql
-- SQLite/go-libsql version

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,              -- UUIDv7 stored as TEXT
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT,
    password_hash TEXT,                -- nullable for passwordless
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- SQLite trigger for updated_at
CREATE TRIGGER trg_users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

### PostgreSQL Migration

```sql
-- migrations/2024_01_15_14_30_create_users.sql
-- PostgreSQL version

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,              -- UUIDv7 as TEXT for consistency
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT,
    password_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- PostgreSQL function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

### Posts Table (Multi-Database Compatible)

```sql
-- Use conditional comments for database-specific features
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    author_id TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT,

    -- SQLite version
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- PostgreSQL version (in separate file)
    -- published_at TIMESTAMPTZ,
    -- created_at TIMESTAMPTZ DEFAULT NOW(),

    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_posts_slug ON posts(slug);
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_published ON posts(published_at) WHERE published_at IS NOT NULL;
```

## Slug Generation Service

```go
// internal/pkg/slug/slug.go
package slug

import (
    "errors"
    "regexp"
    "strings"

    "github.com/jaevor/go-nanoid"
)

var (
    generator, _  = nanoid.Standard(10) // 10 chars, URL-safe
    shortGen, _   = nanoid.Standard(6)  // 6 chars for short slugs
    usernameRe    = regexp.MustCompile(`^[a-z0-9_-]{3,32}$`)
    slugSanitizer = regexp.MustCompile(`[^a-z0-9-]`)
)

// Generate creates a URL-safe slug for content
func Generate() string {
    return generator()
}

// GenerateShort for non-critical resources
func GenerateShort() string {
    return shortGen()
}

// Sanitize user-provided slugs/usernames
func Sanitize(input string) string {
    slug := strings.ToLower(input)
    slug = slugSanitizer.ReplaceAllString(slug, "-")
    slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
    slug = strings.Trim(slug, "-")

    if slug == "" {
        return Generate() // Fallback to random slug
    }
    return slug
}

// ValidateUsername checks if username meets requirements
func ValidateUsername(username string) error {
    if !usernameRe.MatchString(username) {
        return errors.New("username must be 3-32 characters and contain only letters, numbers, hyphens, and underscores")
    }
    return nil
}
```

## Migration System

### Naming Convention

Migrations use timestamp format: `YYYY_MM_DD_HH_MM_<descriptive_name>.sql`

Example: `2024_01_15_14_30_create_users.sql`

### Migration Organization

```text
db/migrations/
├── sqlite/                    # SQLite/go-libsql specific
│   ├── 2024_01_15_14_30_create_users.sql
│   └── 2024_01_15_14_35_create_posts.sql
├── postgres/                  # PostgreSQL specific
│   ├── 2024_01_15_14_30_create_users.sql
│   └── 2024_01_15_14_35_create_posts.sql
└── common/                    # Shared migrations (if any)
    └── 2024_01_15_14_40_create_settings.sql
```

### Goose Configuration

```go
// internal/db/migrate.go
package db

import (
    "database/sql"
    "embed"

    "github.com/pressly/goose/v3"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func Migrate(db *sql.DB, driver string) error {
    var fsys embed.FS
    var dir string

    switch driver {
    case "sqlite3", "go-libsql":
        fsys = sqliteMigrations
        dir = "migrations/sqlite"
    case "postgres", "pgx":
        fsys = postgresMigrations
        dir = "migrations/postgres"
    default:
        return fmt.Errorf("unsupported driver: %s", driver)
    }

    goose.SetBaseFS(fsys)

    if err := goose.SetDialect(driver); err != nil {
        return err
    }

    return goose.Up(db, dir)
}
```

## SQLC Configuration

### Project Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "sqlite"  # or "postgresql"
    queries: "db/queries"
    schema: "db/migrations/sqlite"  # or postgres
    gen:
      go:
        package: "generated"
        out: "db/generated"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
        overrides:
          - db_type: "TEXT"
            go_type: "string"
            nullable: false
```

## SQLC Queries

### User Queries

```sql
-- db/queries/users.sql

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateUser :one
INSERT INTO users (id, username, email, display_name, password_hash)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = ?, display_name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
```

### Post Queries

```sql
-- db/queries/posts.sql

-- name: GetPostBySlug :one
SELECT
    p.*,
    u.username as author_username,
    u.display_name as author_display_name
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.slug = ?
LIMIT 1;

-- name: ListPostsByAuthor :many
SELECT * FROM posts
WHERE author_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreatePost :one
INSERT INTO posts (id, slug, author_id, title, content)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePost :one
UPDATE posts
SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: PublishPost :one
UPDATE posts
SET published_at = CURRENT_TIMESTAMP
WHERE id = ? AND published_at IS NULL
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ?;
```

## Repository Pattern

### Interface Definition

```go
// internal/interfaces/user.go
package interfaces

import (
    "context"
    "database/sql"
)

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    List(ctx context.Context, limit, offset int) ([]*User, error)
    Update(ctx context.Context, id string, updates UserUpdates) (*User, error)
    Delete(ctx context.Context, id string) error
    Count(ctx context.Context) (int64, error)
}

// internal/interfaces/post.go
type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    GetByID(ctx context.Context, id string) (*Post, error)
    GetBySlug(ctx context.Context, slug string) (*PostWithAuthor, error)
    ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*Post, error)
    Update(ctx context.Context, id string, updates PostUpdates) (*Post, error)
    Publish(ctx context.Context, id string) (*Post, error)
    Delete(ctx context.Context, id string) error
}
```

### Implementation

```go
// internal/domain/users/repository.go
package users

import "myapp/internal/interfaces"

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "myapp/db/generated"  // SQLC generated
    "myapp/internal/pkg/identifier"
    "myapp/internal/pkg/slug"
)

type userRepository struct {
    db      *sql.DB
    queries *generated.Queries
}

func NewUserRepository(database *sql.DB) interfaces.UserRepository {
    return &userRepository{
        db:      database,
        queries: generated.New(database),
    }
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
    // Validate ID format (must be UUIDv7)
    if user.ID == "" {
        return errors.New("user ID is required")
    }

    if err := identifier.ValidateID(user.ID); err != nil {
        return fmt.Errorf("invalid user ID: %w", err)
    }

    // Validate username
    if err := slug.ValidateUsername(user.Username); err != nil {
        return fmt.Errorf("invalid username: %w", err)
    }

    // Create user via SQLC
    dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
        ID:          user.ID,
        Username:    user.Username,
        Email:       user.Email,
        DisplayName: sql.NullString{String: user.DisplayName, Valid: user.DisplayName != ""},
        PasswordHash: sql.NullString{String: user.PasswordHash, Valid: user.PasswordHash != ""},
    })

    if err != nil {
        if isUniqueConstraintError(err) {
            return ErrDuplicateUser
        }
        return fmt.Errorf("creating user: %w", err)
    }

    // Map back to domain model
    *user = mapDBUserToDomain(dbUser)
    return nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
    dbUser, err := r.queries.GetUserByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("getting user by username: %w", err)
    }

    user := mapDBUserToDomain(dbUser)
    return &user, nil
}

// Transaction support
func (r *userRepository) WithTx(tx *sql.Tx) UserRepository {
    return &userRepository{
        db:      tx,
        queries: db.New(tx),
    }
}
```

## Transaction Management

```go
// internal/domain/users/service.go
package users

import "myapp/internal/interfaces"

type service struct {
    db          *sql.DB
    userRepo    interfaces.UserRepository
    profileRepo interfaces.ProfileRepository
}

func (s *service) CreateUserWithProfile(ctx context.Context, input CreateUserInput) (*User, error) {
    // Start transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("starting transaction: %w", err)
    }
    defer tx.Rollback() // Safe to call even after commit

    // Use repositories with transaction
    userRepo := s.userRepo.WithTx(tx)
    profileRepo := s.profileRepo.WithTx(tx)

    // Create user
    user := &User{
        ID:       identifier.NewID(),  // UUIDv7
        Username: slug.Sanitize(input.Username),
        Email:    input.Email,
    }

    if err := userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    // Create profile
    profile := &Profile{
        ID:     identifier.NewID(),
        UserID: user.ID,
        Bio:    input.Bio,
    }

    if err := profileRepo.Create(ctx, profile); err != nil {
        return nil, err
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("committing transaction: %w", err)
    }

    return user, nil
}
```

## Database Observability

```go
// internal/db/instrumented.go
package db

import (
    "database/sql"

    "github.com/XSAM/otelsql"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func OpenInstrumented(driverName, dsn string) (*sql.DB, error) {
    // Register instrumented driver
    driverName, err := otelsql.Register(driverName,
        otelsql.WithAttributes(semconv.DBSystemKey.String(driverName)),
        otelsql.WithAttributes(semconv.DBNameKey.String("tracks")),
        otelsql.WithQueryFormatter(sanitizeQuery),
    )
    if err != nil {
        return nil, err
    }

    db, err := sql.Open(driverName, dsn)
    if err != nil {
        return nil, err
    }

    // Record connection pool metrics
    if err := otelsql.RecordStats(db); err != nil {
        return nil, err
    }

    return db, nil
}

// Sanitize queries for tracing (remove sensitive data)
func sanitizeQuery(query string) string {
    // Implementation to remove PII from queries
    return query
}
```

## Performance Considerations

### Indexing Strategy

1. **Primary keys**: All tables use TEXT UUIDs as primary keys
2. **Foreign keys**: Index all foreign key columns
3. **Query patterns**: Index columns used in WHERE clauses
4. **Composite indexes**: For multi-column queries
5. **Partial indexes**: For filtered queries (PostgreSQL)

### Connection Pool Settings

```go
// internal/db/connection.go
func SetupConnectionPool(db *sql.DB, driver string) {
    switch driver {
    case "postgres":
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(5)
        db.SetConnMaxIdleTime(5 * time.Minute)
        db.SetConnMaxLifetime(1 * time.Hour)
    case "sqlite3", "go-libsql":
        db.SetMaxOpenConns(1) // SQLite doesn't handle concurrency well
        db.SetMaxIdleConns(1)
    }
}
```

## Common Patterns

### Pagination

```go
type PageRequest struct {
    Limit  int
    Offset int
}

func (p PageRequest) Validate() error {
    if p.Limit < 1 || p.Limit > 100 {
        return errors.New("limit must be between 1 and 100")
    }
    if p.Offset < 0 {
        return errors.New("offset must be non-negative")
    }
    return nil
}

type PageResponse[T any] struct {
    Items      []T   `json:"items"`
    Total      int64 `json:"total"`
    Limit      int   `json:"limit"`
    Offset     int   `json:"offset"`
    HasNext    bool  `json:"has_next"`
    HasPrev    bool  `json:"has_prev"`
}
```

### Error Handling

```go
// internal/pkg/errors/errors.go
package errors

import "errors"

var (
    ErrNotFound       = errors.New("resource not found")
    ErrDuplicateUser  = errors.New("user already exists")
    ErrDuplicateSlug  = errors.New("slug already in use")
    ErrInvalidID      = errors.New("invalid identifier format")
)

func isUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }

    errStr := err.Error()
    return strings.Contains(errStr, "UNIQUE constraint failed") || // SQLite
           strings.Contains(errStr, "duplicate key value")          // PostgreSQL
}
```

## Next Steps

- Continue to [Authentication →](./3_authentication.md)
- Back to [Core Architecture](./1_core_architecture.md)
- Back to [Summary](./0_summary.md)
