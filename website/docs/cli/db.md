# tracks db

Manage database migrations for Tracks applications.

## Usage

```bash
tracks db <command> [flags]
```

These commands must be run from within a Tracks project directory (where `.tracks.yaml` exists). They read database configuration from environment variables, typically loaded from your `.env` file.

## Subcommands

| Command | Description |
|---------|-------------|
| `migrate` | Apply pending migrations |
| `rollback` | Roll back last migration |
| `status` | Show migration status |
| `reset` | Reset database |

## tracks db migrate

Apply all pending database migrations in order. Migrations are SQL files in `internal/db/migrations/` that define schema changes.

```bash
tracks db migrate [--dry-run]
```

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview migrations without applying |

Use `--dry-run` to see what would run before making changes:

```bash
tracks db migrate --dry-run
```

## tracks db rollback

Roll back the most recently applied migration. Useful for undoing a migration during development or fixing issues.

```bash
tracks db rollback
```

## tracks db status

Show which migrations have been applied and which are pending. Helpful for debugging migration issues or verifying deployment state.

```bash
tracks db status
```

## tracks db reset

Reset the database by rolling back all migrations then reapplying them. This destroys all data - use with caution.

```bash
tracks db reset [--force]
```

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

In CI or scripts, use `--force` to skip the interactive confirmation:

```bash
tracks db reset --force
```

## Supported Drivers

| Driver | Status |
|--------|--------|
| `postgres` | Supported |
| `sqlite3` | Use `make migrate-*` targets |
| `go-libsql` | Use `make migrate-*` targets |

For SQLite and go-libsql projects, use the generated Makefile targets instead:

```bash
make migrate-up
make migrate-down
make migrate-status
```

## Environment

| Variable | Description |
|----------|-------------|
| `APP_DATABASE_URL` | Database connection string (required) |

The command loads `.env` automatically, or you can set the variable directly:

```bash
APP_DATABASE_URL=postgres://localhost/myapp tracks db migrate
```

## Common Errors

**Not in a project:** Run from your project root where `.tracks.yaml` exists.

```text
Error: not in a Tracks project directory (missing .tracks.yaml)
```

**Missing database URL:** Create a `.env` file or set the variable.

```text
Error: APP_DATABASE_URL environment variable is not set
```

**Connection failed:** Check that your database is running and the URL is correct.

```text
Error: failed to connect to database: connection refused
```

## See Also

- [Database Setup](../guides/database-setup.md) - Configuration and drivers
- [Migrations Guide](../guides/migrations.md) - Writing migration files
