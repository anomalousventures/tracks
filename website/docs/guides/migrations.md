# Migrations

Manage database schema changes with [Goose](https://github.com/pressly/goose) migrations.

## Directory Structure

```text
internal/db/
├── migrations/         # SQL migration files
├── queries/            # SQLC query files
└── generated/          # SQLC generated code
```

## Creating Migrations

```bash
make migrate-create NAME=add_posts_table
```

Creates: `internal/db/migrations/20240115143022_add_posts_table.sql`

### File Format

```sql
-- +goose Up
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS posts;
```

## Running Migrations

**Make targets:**

```bash
make migrate-up       # Apply pending
make migrate-down     # Roll back last
make migrate-status   # Check status
```

**CLI (PostgreSQL only):**

```bash
tracks db migrate            # Apply pending
tracks db migrate --dry-run  # Preview
tracks db rollback           # Roll back last
tracks db status             # Check status
tracks db reset --force      # Reset database
```

## Best Practices

**Always include down migrations:**

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN bio TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN bio;
```

**Use idempotent statements:**

```sql
CREATE TABLE IF NOT EXISTS posts (...);
DROP TABLE IF EXISTS posts;
```

**One change per migration** - keep migrations focused on a single logical change.

**Test both directions** before deploying:

```bash
make migrate-up && make migrate-down && make migrate-up
```

## Driver-Specific Notes

### PostgreSQL

- Transactional DDL (atomic schema changes)
- Supports `CREATE INDEX CONCURRENTLY`

### SQLite

Limited `ALTER TABLE` support. To drop a column, recreate the table:

```sql
-- +goose Up
CREATE TABLE posts_new (...);
INSERT INTO posts_new SELECT ... FROM posts;
DROP TABLE posts;
ALTER TABLE posts_new RENAME TO posts;
```

## Troubleshooting

**Dirty database:** Migration failed midway. Check status, manually fix, then update `goose_db_version` table.

**Order conflicts:** Coordinate timestamps when multiple developers create migrations. Rebase before merging.
