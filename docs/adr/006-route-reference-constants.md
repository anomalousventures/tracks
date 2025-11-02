# ADR-006: Route Reference Constants Pattern

**Status:** Accepted
**Date:** 2025-11-02
**Context:** Epic 3 - Project Generation

## Context

During Epic 3 implementation for project generation, we encountered a common problem in web applications: route paths scattered throughout the codebase as magic strings. This creates brittleness when refactoring routes and makes it difficult to find all references to a particular route.

**Problem:**

```go
// ❌ WRONG: Magic strings scattered across codebase

// internal/http/handlers/health_handler.go
func (h *HealthHandler) Show(w http.ResponseWriter, r *http.Request) {
    // Handler doesn't know its own route
}

// internal/http/routes.go
r.Get("/api/health", healthHandler.Show)

// internal/http/views/components/nav.templ
<a href="/api/health">Health Check</a>

// internal/monitoring/metrics.go
httpDuration.WithLabelValues("/api/health", "GET").Observe(duration)

// Problem: If we change the route path, we must find and update all these strings
// No compile-time safety, no IDE refactoring support, easy to miss references
```

When a route path changes, developers must:

1. Search the entire codebase for string literals
2. Manually verify each match is the correct route
3. Hope they didn't miss any references
4. Risk runtime errors from typos or missed updates

We needed a pattern that provides:

1. Compile-time safety for route references
2. Single source of truth for route paths
3. Easy refactoring with IDE support
4. Consistent usage across handlers, templates, metrics, and logging

## Decision

We will use a **route constants package** where all route paths are defined as typed constants, and all route references throughout the codebase MUST use these constants instead of magic strings.

### Key Decisions

1. **Constants Location:** `internal/http/routes/routes.go` (part of HTTP layer per ADR-005)

2. **All References Use Constants:** Handlers, templates, middleware, metrics, logging, documentation

3. **No Magic Strings:** Route paths as string literals are prohibited except in the routes constants definition

4. **Compile-Time Safety:** Typos in route names cause compile errors, not runtime failures

5. **Package Structure:**

   ```text
   internal/http/
   ├── routes/
   │   └── routes.go         # Route path constants
   ├── handlers/             # Import routes package
   ├── middleware/           # Import routes package
   └── views/                # Import routes package (for templ)
   ```

### Complete Architecture Example

```go
// ✅ CORRECT: Route constants as single source of truth

// internal/http/routes/routes.go
package routes

const (
    // API routes
    APIHealth       = "/api/health"
    APIUsersIndex   = "/api/users"
    APIUsersShow    = "/api/users/{id}"
    APIUsersCreate  = "/api/users"

    // Web routes
    WebHome         = "/"
    WebDashboard    = "/dashboard"
    WebUserProfile  = "/users/{username}"
)

// internal/http/routes.go (route registration)
package http

import (
    "myapp/internal/http/handlers"
    "myapp/internal/http/routes"
)

func (s *Server) registerRoutes() {
    r := chi.NewRouter()

    // Use constants for route registration
    r.Get(routes.APIHealth, s.healthHandler.Show)
    r.Get(routes.APIUsersIndex, s.userHandler.Index)
    r.Get(routes.APIUsersShow, s.userHandler.Show)
    r.Post(routes.APIUsersCreate, s.userHandler.Create)

    s.router = r
}

// internal/http/views/components/nav.templ
package components

import "myapp/internal/http/routes"

templ Nav() {
    <nav>
        <a href={ templ.URL(routes.WebHome) }>Home</a>
        <a href={ templ.URL(routes.WebDashboard) }>Dashboard</a>
    </nav>
}

// internal/monitoring/metrics.go
package monitoring

import "myapp/internal/http/routes"

func RecordRequest(route string, method string, duration time.Duration) {
    httpDuration.WithLabelValues(route, method).Observe(duration)
}

// Usage in middleware
RecordRequest(routes.APIHealth, "GET", duration)

// internal/http/handlers/health_handler_test.go
package handlers_test

import "myapp/internal/http/routes"

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", routes.APIHealth, nil)
    // ...
}
```

### Benefits Demonstrated

**Before (Magic Strings):**

```go
// Changing route path requires finding all string literals
// Typos cause runtime errors
// No IDE refactoring support

// Search results for "/api/health":
// - route definition
// - handler test
// - nav template
// - metrics call
// - documentation
// - old commented code (false positive)
// - log message (false positive)
```

**After (Constants):**

```go
// Changing route path: update one constant
// Typos cause compile errors
// IDE "Find Usages" shows all real references

// 1. Update constant
const APIHealth = "/api/v2/health"  // Changed

// 2. Compile catches all references automatically
// 3. No need to search, no false positives
```

## Consequences

### Positive

- **Compile-Time Safety** - Typos in route references cause build failures, not runtime errors
- **Single Source of Truth** - Route paths defined in exactly one place
- **Refactoring Safety** - IDE "Find Usages" and "Rename" work correctly
- **Type Safety** - Routes are string constants, not arbitrary strings
- **Documentation** - Constants serve as route inventory
- **Consistency** - Same pattern everywhere (handlers, tests, templates, metrics)
- **Code Review** - Easy to spot magic strings during PR review

### Negative

- **Extra Import** - Must import `routes` package in all consumers
- **Indirection** - One extra step to see actual path (though IDE "Go to Definition" helps)
- **Constants Proliferation** - Large apps will have many route constants

### Neutral

- **Naming Convention** - Need consistent naming for route constants (addressed in Implementation Rules)
- **Generated Code** - Code generator must create route constants when scaffolding resources

## Implementation Rules

### Rule 1: All Route References Must Use Constants

Route paths as string literals are prohibited except in `internal/http/routes/routes.go`.

```go
// ✅ CORRECT
import "myapp/internal/http/routes"

r.Get(routes.APIHealth, handler.Show)

// ❌ WRONG
r.Get("/api/health", handler.Show)  // Magic string prohibited
```

### Rule 2: Constant Naming Convention

Route constants follow this pattern:

- **Prefix:** `API` for API routes, `Web` for web page routes
- **Resource:** Capitalized resource name (singular for REST, plural for collections)
- **Action:** HTTP verb or page name

```go
// ✅ CORRECT
const (
    // API REST endpoints
    APIUsersIndex  = "/api/users"        // GET collection
    APIUsersShow   = "/api/users/{id}"   // GET single
    APIUsersCreate = "/api/users"        // POST
    APIUsersUpdate = "/api/users/{id}"   // PUT/PATCH
    APIUsersDelete = "/api/users/{id}"   // DELETE

    // Web pages
    WebHome           = "/"
    WebUserProfile    = "/users/{username}"
    WebAdminDashboard = "/admin/dashboard"
)

// ❌ WRONG
const (
    GetUsers      = "/api/users"       // Unclear if API or web
    api_users     = "/api/users"       // Wrong case
    UsersListPage = "/api/users"       // Ambiguous API vs web
)
```

### Rule 3: Constants for All Routes

Every route in the application MUST have a corresponding constant, even simple ones.

```go
// ✅ CORRECT
const WebHome = "/"
r.Get(routes.WebHome, homeHandler.Show)

// ❌ WRONG
r.Get("/", homeHandler.Show)  // Even simple routes need constants
```

### Rule 4: Path Parameters in Constants

Route constants include Chi/gorilla-style path parameters.

```go
// ✅ CORRECT
const (
    APIUsersShow    = "/api/users/{id}"
    WebUserProfile  = "/users/{username}"
)

// Usage in handler
func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")  // Extract parameter
    // ...
}

// Usage in template (construct with parameter)
<a href={ templ.URL(fmt.Sprintf("/users/%s", username)) }>Profile</a>

// Or use helper function
<a href={ templ.URL(routes.UserProfilePath(username)) }>Profile</a>
```

### Rule 5: Route Helper Functions (Optional)

For routes with parameters, provide helper functions for path construction.

```go
// ✅ CORRECT: Optional helpers for parameterized routes
package routes

import "fmt"

const (
    APIUsersShow   = "/api/users/{id}"
    WebUserProfile = "/users/{username}"
)

// Helper functions for constructing paths with parameters
func APIUsersShowPath(id string) string {
    return fmt.Sprintf("/api/users/%s", id)
}

func WebUserProfilePath(username string) string {
    return fmt.Sprintf("/users/%s", username)
}

// Usage in template
<a href={ templ.URL(routes.WebUserProfilePath(user.Username)) }>Profile</a>
```

## Alternatives Considered

### 1. Magic Strings Everywhere

**Rejected:** No compile-time safety, brittle refactoring, difficult to find all route references. This is the problem we're solving.

### 2. Route Constants Colocated with Handlers

**Structure:**

```text
internal/http/handlers/
├── user_handler.go
├── user_routes.go      # Route constants with handler
```

**Rejected:** Creates coupling between handlers and route definitions. Cross-cutting concerns (middleware, metrics, logging) would need to import handler packages. Also violates ADR-005 HTTP layer organization.

### 3. Route Registry/Router Pattern

**Example:**

```go
router := NewRouter()
router.GET("api.health", "/api/health", handler)
router.Redirect("api.health")  // Use symbolic name
```

**Rejected:** Over-engineered for the problem. Simple constants provide compile-time safety without additional abstraction layer. Also requires runtime registry lookup instead of compile-time constant.

### 4. Generated Route Constants from OpenAPI Spec

**Rejected for initial implementation:** Adds tooling complexity and build dependencies. We can add this later as an enhancement without changing the pattern. Constants can still be the source of truth with manual definition initially.

## Implementation Notes

### Epic 3 Integration

- **Task 34 (Issue #123):** Create `internal/templates/project/internal/http/routes/routes.go.tmpl`
- **Task 35 (Issue #124):** Tests for routes template
- **Initial constant:** `APIHealth = "/api/health"`
- **Path Fix:** Epic 3 documentation incorrectly specified `internal/routes/` - corrected to `internal/http/routes/` per ADR-005

### Code Generator Behavior

When generating resources (e.g., `tracks generate resource user`):

1. Generate route constants in `internal/http/routes/routes.go` with TRACKS markers
2. Use constants in generated handler tests
3. Use constants in generated route registration
4. Update documentation with route constant references

### Migration Path

For existing code using magic strings:

1. Define all routes as constants in `internal/http/routes/routes.go`
2. Update route registration to use constants
3. Update handlers, middleware, templates to import and use constants
4. Add linter rule to prevent new magic string routes (future enhancement)

### Documentation

Route constants serve as route inventory. Add godoc to explain routes:

```go
// API routes for health checks and system status
const (
    // APIHealth returns application health status
    APIHealth = "/api/health"

    // APIVersion returns application version info
    APIVersion = "/api/version"
)
```

## References

- [ADR-005: HTTP Layer Architecture](./005-http-layer-architecture.md)
- [Epic 3: Project Generation](../roadmap/phases/0-foundation/epics/3-project-generation.md)
- [Issue #123: Create routes constants template](https://github.com/anomalousventures/tracks/issues/123)
- [Issue #124: Routes template tests](https://github.com/anomalousventures/tracks/issues/124)
- [Chi Router Documentation](https://github.com/go-chi/chi)
- [Go Proverb: "Clear is better than clever"](https://go-proverbs.github.io/)
