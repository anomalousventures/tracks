# Layer Guide

Detailed guide to each layer in a Tracks-generated application.

## HTTP Layer (`internal/http/`)

The HTTP layer handles all web-facing concerns. It converts HTTP requests into domain operations and domain results back into HTTP responses.

### Server (`server.go`)

Sets up the HTTP server with dependency injection:

```go
type Server struct {
    cfg    *config.ServerConfig
    logger interfaces.Logger
    router chi.Router

    // Service dependencies (injected)
    healthService interfaces.HealthService
    // userService interfaces.UserService  (added incrementally)
}

func NewServer(cfg *config.ServerConfig, logger interfaces.Logger) *Server {
    return &Server{
        cfg:    cfg,
        logger: logger,
        router: chi.NewRouter(),
    }
}

// Builder pattern for dependency injection
func (s *Server) WithHealthService(svc interfaces.HealthService) *Server {
    s.healthService = svc
    return s
}

func (s *Server) RegisterRoutes() *Server {
    registerRoutes(s)
    return s
}

func (s *Server) Start(ctx context.Context) error {
    // Graceful shutdown logic
}
```

**Key Points:**

- Builder pattern allows incremental service registration
- Graceful shutdown with context cancellation
- No global state - everything is passed in

### Routes (`routes.go`)

**File structure:**

- `internal/http/routes.go` - Route registration and middleware chain
- `internal/http/routes/` - Domain-based route files (health.go, users.go, etc.) with constants and helpers

Registers routes and applies middleware chain. Routes serve HTML by default (HYPERMEDIA-first):

```go
func registerRoutes(s *Server) {
    r := s.router

    // Global middleware (runs for all requests)
    r.Use(middleware.RequestID)
    r.Use(middleware.NewLogging(s.logger))
    r.Use(middleware.Recoverer)
    r.Use(middleware.CORS())
    r.Use(middleware.Security())

    // Health check (no auth required)
    r.Get(routes.APIHealth, handlers.NewHealthHandler(s.healthService).Handle)

    // User routes (HYPERMEDIA - serve HTML via templ)
    userHandler := handlers.NewUserHandler(s.userService)
    r.Get(routes.UserIndex, userHandler.HandleIndex)       // List users
    r.Get(routes.UserShow, userHandler.HandleShow)         // Show user profile
    r.Get(routes.UserNew, userHandler.HandleNew)           // New user form
    r.Post(routes.UserCreate, userHandler.HandleCreate)    // Create user
    r.Get(routes.UserEdit, userHandler.HandleEdit)         // Edit user form
    r.Post(routes.UserUpdate, userHandler.HandleUpdate)    // Update user
    r.Post(routes.UserDelete, userHandler.HandleDelete)    // Delete user
}
```

**Key Points:**

- **Middleware order matters** - RequestID first, auth last
- Group routes with `r.Route()` to share middleware
- Route patterns use route constants (not magic strings)

### Route Constants (`routes/`)

Routes are organized by domain in separate files. Each file contains route constants, slug constants (for parameterized routes), and helper functions for type-safe URL generation.

**Simple Domain (`routes/health.go` - API endpoint, no parameters):**

```go
package routes

const (
    APIHealth = "/api/health"  // JSON endpoint
)
```

**Complex Domain (`routes/users.go` - HYPERMEDIA routes with helpers):**

```go
package routes

import "net/url"

// UserSlugParam is exported so handlers can extract parameters without magic strings.
// usersPath remains unexported as it's an internal routing detail.
const (
    usersPath     = "users"
    UserSlugParam = "username"
)

// HYPERMEDIA route constants (serve HTML via templ)
const (
    UserIndex  = "/" + usersPath                                 // GET  /users
    UserShow   = "/" + usersPath + "/:" + UserSlugParam          // GET  /users/:username
    UserNew    = "/" + usersPath + "/new"                        // GET  /users/new
    UserCreate = "/" + usersPath                                 // POST /users
    UserEdit   = "/" + usersPath + "/:" + UserSlugParam + "/edit" // GET  /users/:username/edit
    UserUpdate = "/" + usersPath + "/:" + UserSlugParam          // POST /users/:username
    UserDelete = "/" + usersPath + "/:" + UserSlugParam          // POST /users/:username
)

// RouteURL substitutes parameters and URL-encodes values
func RouteURL(route string, params ...string) string {
    // ... implementation with url.PathEscape
}

// Typed helper functions for type safety
func UserShowURL(username string) string {
    return RouteURL(UserShow, UserSlugParam, username)
}

func UserEditURL(username string) string {
    return RouteURL(UserEdit, UserSlugParam, username)
}
// ... other helpers
```

**Key Benefits:**

- **Domain-based organization** - All user routes in one file
- **Type safety** - Compile-time checks prevent typos
- **URL encoding** - Automatic via RouteURL helper
- **Typed helpers** - IDE autocomplete, refactoring support
- **HYPERMEDIA-first** - Form routes (/new, /edit) for HTML

See the [Routing Guide](./routing-guide.md) for complete details on domain-based routing patterns.

### Handlers (`handlers/`)

Convert HTTP to domain operations:

```go
type UserHandler struct {
    userService interfaces.UserService
}

func NewUserHandler(userService interfaces.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    // 1. Extract parameters - use exported constant to avoid magic strings
    username := chi.URLParam(r, routes.UserSlugParam)
    if username == "" {
        http.Error(w, "username required", http.StatusBadRequest)
        return
    }

    // 2. Call service
    user, err := h.userService.GetByUsername(r.Context(), username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

**Handler Responsibilities:**

- ✅ Extract and validate request parameters
- ✅ Call service methods
- ✅ Format responses (JSON, HTML, redirects)
- ✅ Orchestrate multiple services for complex operations
- ❌ **NO** business logic
- ❌ **NO** direct database access

### Middleware (`middleware/`)

Single-responsibility composable functions:

```go
// Logging middleware
func NewLogging(logger interfaces.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            logger.Info("request started",
                "method", r.Method,
                "path", r.URL.Path,
                "remote_addr", r.RemoteAddr,
            )

            // Wrap response writer to capture status
            ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
            next.ServeHTTP(ww, r)

            logger.Info("request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "status", ww.Status(),
                "duration_ms", time.Since(start).Milliseconds(),
            )
        })
    }
}
```

**Middleware Best Practices:**

- One responsibility per middleware
- Composable - can be reordered
- Pass dependencies via closure (logger, config, etc.)
- Use context for request-scoped values

## Interfaces Package (`internal/interfaces/`)

The interfaces package defines contracts between layers. This is a **critical architectural choice** that prevents import cycles.

### Why Separate Interfaces?

**Problem:** If interfaces are with implementations, you get import cycles:

```text
❌ WITHOUT interfaces/ package:

handlers/ imports domain/users/
domain/users/ imports domain/posts/
domain/posts/ imports domain/users/  ← CYCLE!
```

**Solution:** Interfaces in a separate package breaks the cycle:

```text
✅ WITH interfaces/ package:

handlers/ imports interfaces/
domain/users/ implements interfaces.UserService
domain/posts/ implements interfaces.PostService
No cycles!
```

### Interface Organization

One file per domain:

```text
interfaces/
├── health.go    # HealthService, HealthRepository
├── user.go      # UserService, UserRepository
└── post.go      # PostService, PostRepository
```

**Example** (`interfaces/user.go`):

```go
package interfaces

import "context"

//go:generate mockery --name=UserService --outpkg=mocks --output=../../tests/mocks

type UserService interface {
    Create(ctx context.Context, req CreateUserRequest) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context, limit, offset int) ([]*User, error)
    Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
    Delete(ctx context.Context, id string) error
}

type UserRepository interface {
    Insert(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindAll(ctx context.Context, limit, offset int) ([]*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

**Key Rules:**

- **Zero implementations** - only interface definitions
- **go:generate directive** - for automatic mock generation
- **Context first** - always first parameter
- **Return errors** - explicit error handling

### Interface Compliance

Ensure implementations satisfy interfaces at compile time:

```go
// In domain/users/service.go
var _ interfaces.UserService = (*Service)(nil)

// In domain/users/repository.go
var _ interfaces.UserRepository = (*Repository)(nil)
```

If the interface changes and implementation doesn't match, you get a **compile error**.

## Domain Layer (`internal/domain/`)

Business logic lives here, organized by domain (feature area).

### Directory Structure

```text
domain/
├── health/
│   ├── service.go
│   ├── service_test.go
│   ├── repository.go
│   └── dto.go
├── users/
│   ├── service.go
│   ├── service_test.go
│   ├── repository.go
│   ├── repository_test.go
│   └── dto.go
└── posts/
    ├── service.go
    ├── service_test.go
    ├── repository.go
    └── dto.go
```

**Organize by domain, not by layer** - all user-related code in `users/`.

### Service (`service.go`)

Contains business logic:

```go
type Service struct {
    repo interfaces.UserRepository
}

func NewService(repo interfaces.UserRepository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
    if req.Age < 18 {
        return nil, ErrUserTooYoung
    }

    user := &User{
        ID:        uuid.New().String(),
        Name:      req.Name,
        Email:     req.Email,
        Age:       req.Age,
        CreatedAt: time.Now(),
    }

    if err := s.repo.Insert(ctx, user); err != nil {
        return nil, fmt.Errorf("inserting user: %w", err)
    }

    return user, nil
}
```

**Service Responsibilities:**

- ✅ Business rules and validations
- ✅ Coordinate repository calls
- ✅ Transaction boundaries
- ❌ **NO** HTTP knowledge (no http.Request, http.Response)
- ❌ **NO** direct database access (use repository)

### Repository (`repository.go`)

Wraps SQLC-generated code:

```go
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

func (r *Repository) Insert(ctx context.Context, user *User) error {
    params := db.CreateUserParams{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Age:       int32(user.Age),
        CreatedAt: user.CreatedAt,
    }

    if err := r.queries.CreateUser(ctx, params); err != nil {
        return fmt.Errorf("creating user: %w", err)
    }

    return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
    dbUser, err := r.queries.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("getting user: %w", err)
    }

    return &User{
        ID:        dbUser.ID,
        Name:      dbUser.Name,
        Email:     dbUser.Email,
        Age:       int(dbUser.Age),
        CreatedAt: dbUser.CreatedAt,
    }, nil
}
```

**Repository Responsibilities:**

- ✅ Wrap SQLC-generated queries
- ✅ Convert between SQLC types and domain types
- ✅ Handle `sql.ErrNoRows` → domain errors
- ❌ **NO** business logic
- ❌ **NO** manual SQL (use SQLC)

### DTOs (`dto.go`)

Request/response data transfer objects:

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}

type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Age       int       `json:"age"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Why separate DTOs?**

- HTTP layer uses DTOs (JSON tags, validation)
- Domain layer uses entities (business logic)
- Decouples HTTP representation from domain model

## Database Layer (`internal/db/`)

Manages database connections and SQL queries.

### Connection Setup (`db.go`)

```go
func New(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
    db, err := sql.Open(cfg.Driver, cfg.URL)
    if err != nil {
        return nil, fmt.Errorf("opening database: %w", err)
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("pinging database: %w", err)
    }

    return db, nil
}
```

### Migrations (`migrations/`)

Goose SQL migrations with timestamp prefixes:

```text
migrations/
├── 20250108120000_create_users_table.sql
└── 20250108120100_create_posts_table.sql
```

**Example migration:**

```sql
-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    age INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

### Queries (`queries/`)

Hand-written SQL queries for SQLC:

```sql
-- name: GetUser :one
SELECT id, name, email, age, created_at
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, name, email, age, created_at
FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateUser :exec
INSERT INTO users (id, name, email, age, created_at)
VALUES (?, ?, ?, ?, ?);
```

**SQLC generates type-safe Go code from these queries.**

### Generated Code (`generated/`)

**DO NOT EDIT** - Generated by SQLC from `queries/*.sql`:

```go
// Code generated by sqlc. DO NOT EDIT.

type GetUserRow struct {
    ID        string
    Name      string
    Email     string
    Age       int32
    CreatedAt time.Time
}

func (q *Queries) GetUser(ctx context.Context, id string) (GetUserRow, error) {
    row := q.db.QueryRowContext(ctx, getUser, id)
    var i GetUserRow
    err := row.Scan(&i.ID, &i.Name, &i.Email, &i.Age, &i.CreatedAt)
    return i, err
}
```

## Dependency Injection in main.go

The `cmd/server/main.go` file wires everything together:

```go
func main() {
    ctx := context.Background()

    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    logger := logging.New(cfg.Logging.Level)

    // TRACKS:DB:BEGIN
    database, err := db.New(ctx, cfg.Database)
    if err != nil {
        logger.Fatal("failed to connect to database", "error", err)
    }
    defer database.Close()
    // TRACKS:DB:END

    // TRACKS:REPOSITORIES:BEGIN
    healthRepo := health.NewRepository(database)
    // userRepo := users.NewRepository(database)  (added incrementally)
    // TRACKS:REPOSITORIES:END

    // TRACKS:SERVICES:BEGIN
    healthService := health.NewService(healthRepo)
    // userService := users.NewService(userRepo)  (added incrementally)
    // TRACKS:SERVICES:END

    srv := http.NewServer(&cfg.Server, logger).
        WithHealthService(healthService).
        // WithUserService(userService).  (added incrementally)
        RegisterRoutes()

    if err := srv.Start(ctx); err != nil {
        logger.Fatal("server error", "error", err)
    }
}
```

**Marker comments** (`// TRACKS:X:BEGIN` / `// TRACKS:X:END`) enable incremental code generation.

## Next Steps

- [**Architecture Overview**](./architecture-overview.md) - Core principles and request flow
- [**Routing Guide**](./routing-guide.md) - HYPERMEDIA-first routing and domain-based organization
- [**Patterns**](./patterns.md) - Common patterns for extending your app
- [**Testing**](./testing.md) - Testing strategies and examples

## See Also

- [CLI: tracks new](../cli/new.mdx) - Creating projects
