# Server Architecture Overview

Learn the architectural principles and patterns behind Tracks-generated applications.

## What Tracks Generates

When you run `tracks new myapp`, you get a production-ready Go web application with:

- **Clean layered architecture** - Clear separation between HTTP, domain logic, and data access
- **Dependency injection** - No global state, testable components
- **Interface-based design** - Flexible, mockable dependencies
- **Type-safe everything** - SQLC for SQL, templ for HTML, route constants for URLs
- **Ready for growth** - Add features incrementally without refactoring

This isn't a framework you import - it's **code you own**. Every file is readable, modifiable, and follows idiomatic Go patterns.

## Core Principles

### 1. Dependency Injection

Services receive dependencies through constructors, not global variables:

```go
// ✅ CORRECT: Dependencies injected
type Service struct {
    repo UserRepository
}

func NewService(repo UserRepository) *Service {
    return &Service{repo: repo}
}

// ❌ WRONG: Global state
var globalDB *sql.DB

type Service struct{}

func (s *Service) GetUser(id string) (*User, error) {
    return globalDB.Query(...)
}
```

**Why?** Explicit dependencies make code testable and coupling visible.

### 2. Interface-Based Design

Accept interfaces, return structs:

```go
// ✅ CORRECT: Accept interface
func NewHandler(userService interfaces.UserService) *Handler {
    return &Handler{userService: userService}
}

// ❌ WRONG: Depend on concrete type
func NewHandler(userService *users.Service) *Handler {
    return &Handler{userService: userService}
}
```

**Why?** Interfaces enable testing with mocks and decouple layers.

### 3. Context Propagation

Always pass `context.Context` as the first parameter:

```go
// ✅ CORRECT
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    return s.repo.FindByID(ctx, id)
}

// ❌ WRONG: No context
func (s *Service) GetUser(id string) (*User, error) {
    return s.repo.FindByID(id)
}
```

**Why?** Enables request cancellation, deadlines, and tracing.

### 4. Explicit Error Handling

Wrap errors with context using `%w`:

```go
// ✅ CORRECT: Wrapped errors
if err != nil {
    return nil, fmt.Errorf("getting user %s: %w", id, err)
}

// ❌ WRONG: Error chain broken
if err != nil {
    return nil, errors.New("failed to get user")
}
```

**Why?** Preserves error chain for debugging while adding context.

### 5. Type Safety

- **SQLC** generates type-safe Go code from SQL queries
- **templ** compiles HTML templates to Go at build time
- **Route constants** replace magic strings with compile-time checks

No reflection, no runtime string parsing.

## Request Flow

Here's the journey of an HTTP request through a Tracks application:

```text
┌─────────────────┐
│  HTTP Request   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│     Middleware Chain                │
│  ┌────────────────────────────────┐ │
│  │ 1. Request ID                  │ │
│  │ 2. Logging                     │ │
│  │ 3. Security (CORS, CSP, HSTS)  │ │
│  │ 4. Authentication              │ │
│  └────────────────────────────────┘ │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Router (Chi)   │  Matches route pattern
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│          Handler                    │
│  • Validates input (DTOs)           │
│  • Calls service(s) via interfaces  │
│  • Orchestrates cross-domain ops    │
│  • Formats response                 │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│          Service                    │
│  • Business logic                   │
│  • Domain validations               │
│  • Calls repository via interface   │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│        Repository                   │
│  • Wraps SQLC-generated code        │
│  • Database queries                 │
│  • Transaction management           │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────┐
│    Database     │  LibSQL / Postgres / SQLite
└─────────────────┘
```

**Key Points:**

- **Middleware**: Runs in order, can short-circuit the request
- **Handler**: Thin layer - orchestrates, doesn't contain business logic
- **Service**: Where business rules live
- **Repository**: Only layer that talks to the database

## Layer Overview

Tracks applications are organized into four main layers:

### HTTP Layer (`internal/http/`)

Handles web-facing concerns:

- **server.go** - HTTP server setup, graceful shutdown
- **routes.go** - Route registration, middleware chain
- **routes/routes.go** - Type-safe route constants
- **handlers/** - HTTP request/response handling
- **middleware/** - Composable middleware functions

**Responsibility:** Convert HTTP into domain operations and back.

### Interfaces Package (`internal/interfaces/`)

Defines contracts between layers:

- One file per domain (e.g., `user.go`, `post.go`)
- **Zero implementations** - only interface definitions
- Consumed by handlers, implemented by services

**Responsibility:** Prevent import cycles, enable mocking.

### Domain Layer (`internal/domain/`)

Contains business logic:

- **service.go** - Business rules, implements service interface
- **repository.go** - Data access, implements repository interface
- **dto.go** - Request/response data transfer objects
- Organized by domain (users/, posts/, health/)

**Responsibility:** Enforce business rules, coordinate data access.

### Database Layer (`internal/db/`)

Manages data persistence:

- **db.go** - Connection setup, transaction helpers
- **migrations/** - Goose SQL migrations
- **queries/** - Hand-written SQL (SQLC input)
- **generated/** - SQLC-generated Go code (don't edit!)

**Responsibility:** Type-safe database access.

## Dependency Flow

Dependencies flow in one direction: **inward**.

```text
┌──────────────────────────────────────────┐
│  HTTP Layer (handlers, middleware)       │
│    depends on ↓                          │
├──────────────────────────────────────────┤
│  Interfaces (service/repo contracts)     │
│    implemented by ↓                      │
├──────────────────────────────────────────┤
│  Domain Layer (services, repositories)   │
│    depends on ↓                          │
├──────────────────────────────────────────┤
│  Database Layer (queries, connections)   │
└──────────────────────────────────────────┘
```

**Rules:**

1. **HTTP can depend on domain** - Handlers use service interfaces
2. **Domain cannot depend on HTTP** - Services don't know about HTTP
3. **Interfaces bridge the gap** - Defined in consumer packages

This creates a **clean separation** where business logic has zero HTTP dependencies.

## How This Differs from Typical Go Apps

### vs. Fat Controllers

**Typical:** Controllers contain business logic

```go
// ❌ Business logic in HTTP handler
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user := parseRequest(r)

    // Validation logic in handler
    if user.Age < 18 {
        http.Error(w, "too young", 400)
        return
    }

    // Database access in handler
    err := h.db.Insert(user)
    // ...
}
```

**Tracks:** Handlers orchestrate, services contain logic

```go
// ✅ Thin handler, delegates to service
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    user, err := h.userService.Create(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

### vs. Interfaces with Implementations

**Typical:** Interfaces next to implementations

```text
users/
├── service.go          # Contains UserService interface
└── service_impl.go     # Implements UserService
```

**Tracks:** Interfaces in consumer packages

```text
interfaces/
└── user.go             # UserService interface

domain/users/
└── service.go          # Implements UserService
```

**Why?** Prevents import cycles, enables clean mocking.

### vs. Manual SQL

**Typical:** Hand-written SQL strings

```go
// ❌ Error-prone, not type-safe
row := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id)
var user User
err := row.Scan(&user.ID, &user.Name, &user.Email)
```

**Tracks:** SQLC generates type-safe code

```go
// ✅ Type-safe, compile-time checked
user, err := r.queries.GetUser(ctx, id)
```

SQL is in `.sql` files, Go code is generated from it.

## Next Steps

- [**Layer Guide**](./layer-guide.md) - Deep dive into each layer's responsibilities
- [**Patterns**](./patterns.md) - Common patterns for extending your app
- [**Testing**](./testing.md) - Testing strategies and examples

## See Also

- [CLI: tracks new](../cli/new.md) - Creating projects
- [CLI: Commands Reference](../cli/commands.md) - All CLI commands
