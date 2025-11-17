# Common Patterns

Learn how to extend and enhance Tracks-generated applications.

## Adding a New Domain

Future versions of Tracks will include `tracks generate resource`, but for now, follow the health check pattern.

### Step 1: Define Interfaces

Create `internal/interfaces/user.go`:

```go
package interfaces

import "context"

//go:generate mockery --name=UserService --outpkg=mocks --output=../../tests/mocks
//go:generate mockery --name=UserRepository --outpkg=mocks --output=../../tests/mocks

type UserService interface {
    Create(ctx context.Context, name, email string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context) ([]*User, error)
}

type UserRepository interface {
    Insert(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindAll(ctx context.Context) ([]*User, error)
}

type User struct {
    ID    string
    Name  string
    Email string
}
```

### Step 2: Create Migration

Create `internal/db/migrations/YYYYMMDDHHMMSS_create_users.sql`:

```sql
-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP TABLE users;
```

Run migration:

```bash
make db-migrate
```

### Step 3: Add SQL Queries

Create `internal/db/queries/users.sql`:

```sql
-- name: GetUser :one
SELECT id, name, email, created_at
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC;

-- name: CreateUser :exec
INSERT INTO users (id, name, email, created_at)
VALUES (?, ?, ?, ?);
```

Generate code:

```bash
make generate
```

### Step 4: Implement Repository

Create `internal/domain/users/repository.go`:

```go
package users

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/youruser/yourproject/internal/db"
    "github.com/youruser/yourproject/internal/interfaces"
)

var _ interfaces.UserRepository = (*Repository)(nil)

type Repository struct {
    db      *sql.DB
    queries *db.Queries
}

func NewRepository(database *sql.DB) *Repository {
    return &Repository{
        db:      database,
        queries: db.New(database),
    }
}

func (r *Repository) Insert(ctx context.Context, user *interfaces.User) error {
    params := db.CreateUserParams{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        CreatedAt: time.Now(),
    }

    if err := r.queries.CreateUser(ctx, params); err != nil {
        return fmt.Errorf("creating user: %w", err)
    }

    return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*interfaces.User, error) {
    row, err := r.queries.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("getting user: %w", err)
    }

    return &interfaces.User{
        ID:    row.ID,
        Name:  row.Name,
        Email: row.Email,
    }, nil
}
```

### Step 5: Implement Service

Create `internal/domain/users/service.go`:

```go
package users

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/youruser/yourproject/internal/interfaces"
)

var _ interfaces.UserService = (*Service)(nil)

type Service struct {
    repo interfaces.UserRepository
}

func NewService(repo interfaces.UserRepository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, email string) (*interfaces.User, error) {
    // Validation
    if name == "" || email == "" {
        return nil, ErrInvalidInput
    }

    user := &interfaces.User{
        ID:    uuid.New().String(),
        Name:  name,
        Email: email,
    }

    if err := s.repo.Insert(ctx, user); err != nil {
        return nil, fmt.Errorf("inserting user: %w", err)
    }

    return user, nil
}
```

### Step 6: Add Handler

Create `internal/http/handlers/user.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/youruser/yourproject/internal/interfaces"
)

type UserHandler struct {
    userService interfaces.UserService
}

func NewUserHandler(userService interfaces.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    user, err := h.userService.Create(r.Context(), req.Name, req.Email)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### Step 7: Register Routes (Domain-Based)

Create `internal/http/routes/users.go` with domain-specific routes:

```go
package routes

const userSlug = "users"

// HYPERMEDIA routes serve HTML via templ (default pattern)
const (
    UserIndex  = "/" + userSlug
    UserShow   = "/" + userSlug + "/:" + userSlug
    UserNew    = "/" + userSlug + "/new"
    UserCreate = "/" + userSlug
    UserEdit   = "/" + userSlug + "/:" + userSlug + "/edit"
    UserUpdate = "/" + userSlug + "/:" + userSlug
)

func UserShowURL(username string) string {
    return RouteURL(UserShow, userSlug, username)
}
// ... other helpers
```

Update `internal/http/routes.go`:

```go
func registerRoutes(s *Server) {
    r := s.router

    // ... existing middleware ...

    // Health (API endpoint)
    r.Get(routes.APIHealth, handlers.NewHealthHandler(s.healthService).Handle)

    // Users (HYPERMEDIA routes - serve HTML)
    userHandler := handlers.NewUserHandler(s.userService)
    r.Get(routes.UserIndex, userHandler.HandleIndex)
    r.Get(routes.UserShow, userHandler.HandleShow)
    r.Get(routes.UserNew, userHandler.HandleNew)
    r.Post(routes.UserCreate, userHandler.HandleCreate)
}
```

See [Routing Guide](./routing-guide.md) for complete domain-based routing patterns.

### Step 8: Wire Dependencies

Update `cmd/server/main.go`:

```go
// TRACKS:REPOSITORIES:BEGIN
healthRepo := health.NewRepository(database)
userRepo := users.NewRepository(database)
// TRACKS:REPOSITORIES:END

// TRACKS:SERVICES:BEGIN
healthService := health.NewService(healthRepo)
userService := users.NewService(userRepo)
// TRACKS:SERVICES:END

srv := http.NewServer(&cfg.Server, logger).
    WithHealthService(healthService).
    WithUserService(userService).
    RegisterRoutes()
```

### Step 9: Generate Mocks

```bash
make generate-mocks
```

This creates `tests/mocks/mock_UserService.go` and `mock_UserRepository.go`.

## Cross-Domain Orchestration

Handlers can use multiple services from different domains.

**Example:** Dashboard handler using multiple services

```go
type DashboardHandler struct {
    userService  interfaces.UserService
    postService  interfaces.PostService
    statsService interfaces.StatsService
}

func NewDashboardHandler(
    userService interfaces.UserService,
    postService interfaces.PostService,
    statsService interfaces.StatsService,
) *DashboardHandler {
    return &DashboardHandler{
        userService:  userService,
        postService:  postService,
        statsService: statsService,
    }
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get current user
    user, err := h.userService.GetCurrent(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    // Get user's recent posts
    posts, err := h.postService.ListByAuthor(ctx, user.ID, 5)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Get user stats
    stats, err := h.statsService.GetForUser(ctx, user.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := DashboardResponse{
        User:  user,
        Posts: posts,
        Stats: stats,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

**Key Point:** Handlers orchestrate, services contain logic. This is safe because all dependencies use interfaces.

## Adding Middleware

Middleware are composable functions that wrap HTTP handlers.

**Example:** Rate limiting middleware

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.Mutex
    requests map[string][]time.Time
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr

        rl.mu.Lock()
        defer rl.mu.Unlock()

        now := time.Now()
        windowStart := now.Add(-rl.window)

        // Get recent requests for this IP
        requests := rl.requests[ip]

        // Remove old requests
        var recent []time.Time
        for _, t := range requests {
            if t.After(windowStart) {
                recent = append(recent, t)
            }
        }

        // Check limit
        if len(recent) >= rl.limit {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        // Add this request
        recent = append(recent, now)
        rl.requests[ip] = recent

        next.ServeHTTP(w, r)
    })
}
```

**Add to routes:**

```go
func registerRoutes(s *Server) {
    r := s.router

    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.NewLogging(s.logger))
    r.Use(middleware.NewRateLimiter(100, time.Minute).Middleware)  // NEW

    // ... routes ...
}
```

## Transaction Boundaries

For operations spanning multiple repositories, use transactions.

**Example:** Create user with initial profile

```go
func (s *Service) CreateUserWithProfile(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Start transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("starting transaction: %w", err)
    }
    defer tx.Rollback()  // No-op if tx.Commit() succeeds

    // Create user
    user := &User{
        ID:    uuid.New().String(),
        Name:  req.Name,
        Email: req.Email,
    }
    if err := s.userRepo.InsertTx(ctx, tx, user); err != nil {
        return nil, fmt.Errorf("inserting user: %w", err)
    }

    // Create profile
    profile := &Profile{
        UserID: user.ID,
        Bio:    req.Bio,
    }
    if err := s.profileRepo.InsertTx(ctx, tx, profile); err != nil {
        return nil, fmt.Errorf("inserting profile: %w", err)
    }

    // Commit
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("committing transaction: %w", err)
    }

    return user, nil
}
```

**Repository with transaction support:**

```go
func (r *Repository) InsertTx(ctx context.Context, tx *sql.Tx, user *User) error {
    queries := db.New(tx)  // Use transaction instead of database
    // ... rest of insert logic
}
```

## Complete Example: Get User Profile

Here's the complete flow from HTTP request to database query.

### Request

```http
GET /u/johndoe HTTP/1.1
```

### Handler

```go
// Route is registered using domain route constant: routes.UserShow
func (h *UserHandler) HandleShow(w http.ResponseWriter, r *http.Request) {
    username := chi.URLParam(r, "users")  // Matches slug constant
    user, err := h.userService.GetByUsername(r.Context(), username)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Use helper function to generate edit URL for template
    editURL := routes.UserEditURL(user.Username)

    // Render HYPERMEDIA response (HTML via templ)
    views.UserProfile(user, editURL).Render(r.Context(), w)
}
```

### Service

```go
func (s *Service) GetByUsername(ctx context.Context, username string) (*interfaces.User, error) {
    if username == "" {
        return nil, ErrInvalidUsername
    }
    return s.repo.FindByUsername(ctx, username)
}
```

### Repository

```go
func (r *Repository) FindByUsername(ctx context.Context, username string) (*interfaces.User, error) {
    row, err := r.queries.GetUserByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("getting user: %w", err)
    }
    return &interfaces.User{
        ID:       row.ID,
        Username: row.Username,
        Name:     row.Name,
        Email:    row.Email,
    }, nil
}
```

### SQLC Query

```sql
-- name: GetUserByUsername :one
SELECT id, username, name, email FROM users WHERE username = ?;
```

### Generated Code

```go
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (GetUserByUsernameRow, error) {
    row := q.db.QueryRowContext(ctx, getUserByUsername, username)
    var i GetUserByUsernameRow
    err := row.Scan(&i.ID, &i.Username, &i.Name, &i.Email)
    return i, err
}
```

## Best Practices

### DO

- ✅ Use dependency injection for all dependencies
- ✅ Define interfaces in `internal/interfaces/`
- ✅ Pass context as first parameter
- ✅ Wrap errors with `%w` for error chains
- ✅ Use route constants instead of magic strings
- ✅ Keep handlers thin (orchestration only)
- ✅ Put business logic in services
- ✅ Use SQLC for all database queries
- ✅ Generate mocks after interface changes
- ✅ Write tests before implementation

### DON'T

- ❌ Store context in struct fields
- ❌ Use global variables or singletons
- ❌ Put business logic in handlers
- ❌ Define interfaces in implementation packages
- ❌ Use string literals for route paths
- ❌ Write raw SQL in repository methods
- ❌ Skip error wrapping
- ❌ Ignore `golangci-lint` warnings
- ❌ Commit without running tests

## Troubleshooting

### Import cycle detected

**Problem:**

```text
import cycle not allowed
package yourmodule/internal/http/handlers
    imports yourmodule/internal/domain/users
    imports yourmodule/internal/domain/posts
    imports yourmodule/internal/domain/users
```

**Solution:** Move interfaces to `internal/interfaces/`. Handlers import interfaces, services implement them.

### Mock generation fails

**Problem:**

```bash
$ make generate-mocks
Error: could not import yourmodule/internal/interfaces
```

**Solution:**

1. Ensure interfaces package has no implementation code
2. Run `go mod tidy`
3. Check `//go:generate` directives are correct
4. Run `make generate-mocks` again

### Tests fail after interface change

**Problem:** Changed interface signature, tests now fail.

**Solution:**

1. Update interface in `internal/interfaces/`
2. Update implementation in service/repository
3. Regenerate mocks: `make generate-mocks`
4. Update test code to match new signature

### Route not found

**Problem:** Getting 404 for a route you just added.

**Solution:**

1. Check route is registered in `routes.go`
2. Verify route constant matches pattern
3. Check middleware isn't blocking the route
4. Use `chi.Walk()` to debug registered routes:

```go
chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
    fmt.Printf("%s %s\n", method, route)
    return nil
})
```

## Next Steps

- [**Architecture Overview**](./architecture-overview.md) - Core principles
- [**Layer Guide**](./layer-guide.md) - Deep dive on each layer
- [**Routing Guide**](./routing-guide.md) - HYPERMEDIA-first routing and domain-based organization
- [**Testing**](./testing.md) - Testing strategies

## See Also

- [CLI: tracks new](../cli/new.mdx) - Creating projects
