# Database Troubleshooting

Common issues and solutions for Tracks database operations.

## Connection Issues

### Connection Refused

```bash
# PostgreSQL - check if running
docker compose ps
docker compose up -d postgres

# SQLite - check data directory
mkdir -p data
```

### Authentication Failed

Verify credentials match between `.env` and `docker-compose.yml`:

```bash
APP_DATABASE_URL=postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable
```

### Database Does Not Exist

```bash
docker compose exec postgres createdb -U postgres myapp
```

## Migration Issues

### Dirty Database

Migration failed midway. Check status and manually fix:

```bash
tracks db status
```

```sql
SELECT * FROM goose_db_version ORDER BY id DESC LIMIT 5;
UPDATE goose_db_version SET is_applied = true WHERE version_id = <version>;
```

### No Migration Files Found

```bash
ls -la internal/db/migrations/
```

## SQLite3 Issues

### Database Is Locked

Multiple processes trying to write. Tracks mitigates this with single-connection mode and WAL (enabled automatically).

If still experiencing locks, add busy timeout:

```sql
PRAGMA busy_timeout=5000;   -- Wait 5s before failing
```

### Disk I/O Error

```bash
df -h data/                                    # Check disk space
sqlite3 data/myapp.db "PRAGMA integrity_check" # Check integrity
```

### Large WAL File

```sql
PRAGMA wal_checkpoint(TRUNCATE);
```

## PostgreSQL Issues

### Too Many Connections

```bash
APP_DATABASE_MAX_OPEN_CONNS=10  # Reduce pool size
```

```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'myapp';
```

### Slow Queries

```sql
-- Enable logging
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- Check for missing indexes
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
```

### Connection Timeout

```bash
APP_DATABASE_CONNECT_TIMEOUT=30s
```

## SQLC Issues

### Column Does Not Exist

Query references a column not in schema:

```bash
make migrate-up  # Apply pending migrations
make generate    # Regenerate SQLC
```

### Type Mismatch

Add overrides in `sqlc.yaml`:

```yaml
gen:
  go:
    overrides:
      - db_type: "text"
        go_type: "string"
```

## Health Check Failing

```bash
# Check logs
docker compose logs app

# Test connectivity
psql $APP_DATABASE_URL -c "SELECT 1"
sqlite3 data/myapp.db "SELECT 1"
```
