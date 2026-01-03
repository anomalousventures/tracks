# SQLC Queries

Write type-safe SQL queries with [SQLC](https://sqlc.dev/).

## Directory Structure

```text
internal/db/
├── queries/      # Your SQL query files
├── generated/    # SQLC output (do not edit)
└── sqlc.yaml     # Configuration
```

## Query Annotations

| Annotation | Returns | Use Case |
|------------|---------|----------|
| `:one` | Single row or error | Get by ID |
| `:many` | Slice of rows | List, search |
| `:exec` | Only error | INSERT, UPDATE, DELETE |
| `:execrows` | Row count + error | Batch operations |

## Writing Queries

Create `internal/db/queries/users.sql`:

```sql
-- name: GetUser :one
SELECT id, email, name FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, name FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (id, email, name) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET name = $2 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

Generate code:

```bash
make generate
```

## Using Generated Code

```go
type UserRepository struct {
    queries *generated.Queries
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{queries: generated.New(db)}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    row, err := r.queries.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return &User{ID: row.ID, Email: row.Email, Name: row.Name}, nil
}
```

### Transactions

```go
func (r *UserRepository) CreateWithProfile(ctx context.Context, user *User) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    queries := r.queries.WithTx(tx)
    // ... use queries ...
    return tx.Commit()
}
```

## Query Patterns

**Nullable fields** use `sql.NullString`, `sql.NullInt64`, etc.

**Dynamic filters:**

```sql
-- name: SearchUsers :many
SELECT * FROM users
WHERE ($1::text IS NULL OR email LIKE '%' || $1 || '%')
ORDER BY created_at DESC;
```

## Configuration

The generated `sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "migrations/"
    gen:
      go:
        package: "generated"
        out: "generated"
        emit_json_tags: true
        emit_empty_slices: true
```

### Type Overrides

```yaml
gen:
  go:
    overrides:
      - db_type: "uuid"
        go_type: "github.com/google/uuid.UUID"
```

## Best Practices

- Use descriptive names: `GetUserByEmail`, `ListActiveUsers`
- Organize queries by domain: `users.sql`, `posts.sql`
- Validate before generating - SQLC checks SQL against your schema
