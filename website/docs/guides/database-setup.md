# Database Setup

Configure and connect your Tracks application to a database.

## Supported Drivers

| Driver | Best For | CGO Required |
|--------|----------|--------------|
| `postgres` | Production workloads, teams | No |
| `sqlite3` | Development, single-server | Yes |
| `go-libsql` | Edge deployment with Turso | Yes |

```bash
tracks new myapp --db postgres
tracks new myapp --db sqlite3
tracks new myapp --db go-libsql
```

## Configuration

Database settings use environment variables with the `APP_` prefix.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_DATABASE_URL` | string | varies | Connection string |
| `APP_DATABASE_CONNECT_TIMEOUT` | duration | `10s` | Connection timeout |
| `APP_DATABASE_MAX_OPEN_CONNS` | int | `25` | Max open connections (PostgreSQL) |
| `APP_DATABASE_MAX_IDLE_CONNS` | int | `5` | Max idle connections (PostgreSQL) |
| `APP_DATABASE_CONN_MAX_IDLE_TIME` | duration | `5m` | Max idle time |
| `APP_DATABASE_CONN_MAX_LIFETIME` | duration | `1h` | Max lifetime |

### Connection Strings

**PostgreSQL:**

```bash
APP_DATABASE_URL=postgres://user:password@localhost:5432/myapp?sslmode=disable
```

**SQLite3:**

```bash
APP_DATABASE_URL=file:./data/myapp.db  # File-based
APP_DATABASE_URL=:memory:               # In-memory
```

**go-libsql (Turso):**

```bash
APP_DATABASE_URL=http://localhost:8081                        # Local
APP_DATABASE_URL=libsql://your-db.turso.io?authToken=token    # Cloud
```

## Connection Pooling

### PostgreSQL

Configurable pool settings for concurrent access:

```go
db.SetMaxOpenConns(cfg.MaxOpenConns)       // default: 25
db.SetMaxIdleConns(cfg.MaxIdleConns)       // default: 5
db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime) // default: 5m
db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // default: 1h
```

### SQLite3

Single connection with WAL mode (enabled automatically):

```go
db.SetMaxOpenConns(1)
db.SetMaxIdleConns(1)
```

### go-libsql

Connects via HTTP to Turso—connection pooling is managed server-side. Local instances use the same single-connection pattern as SQLite3 but without WAL mode (Turso handles durability).

## Local Development

**PostgreSQL:**

```bash
docker compose up -d postgres
make migrate-up
make dev
```

**SQLite3:**

```bash
mkdir -p data
make migrate-up
make dev
```

## Production

### PostgreSQL

- Use managed services (AWS RDS, Cloud SQL)
- Enable SSL: `sslmode=require`
- Consider PgBouncer for high traffic

### SQLite3

- Use persistent volumes
- Enable Litestream for backups
- Single instance only

## Health Checks

The `/health` endpoint verifies database connectivity:

```json
{"status": "ok", "database": "ok", "timestamp": "2024-01-15T10:30:00Z"}
```
