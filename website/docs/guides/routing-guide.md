# Routing Guide

This guide explains Tracks' routing architecture, which emphasizes **HYPERMEDIA-first development** with domain-based organization and type-safe route patterns.

## HYPERMEDIA-First Philosophy

Tracks generates applications that serve **HTML by default** using [templ](https://templ.guide/) templates. This is the opposite of most modern frameworks that default to JSON APIs:

**Default (HYPERMEDIA):**

- Routes serve HTML pages via templ templates
- Forms submit to endpoints that process and redirect
- Progressive enhancement with HTMX for interactivity
- Server-rendered, accessible by default

**Exception (JSON APIs):**

- Only for specific use cases (health checks, webhooks, metrics)
- Generated with `--api` flag when needed
- Always prefixed with `/api/` for clarity

This approach provides:

- **Simpler development** - No separate frontend/backend coordination
- **Better performance** - No SPA JavaScript bundle overhead
- **Accessibility** - HTML works everywhere, JavaScript enhances
- **SEO-friendly** - Content is server-rendered

## Domain-Based Route Organization

Routes are organized by domain into separate files under `internal/http/routes/`:

```text
internal/http/routes/
├── routes.go        # Shared routes (sitemap, robots.txt, API prefix)
├── health.go        # Health check domain (simple, API endpoint)
└── users.go         # User domain (complex, HYPERMEDIA routes with helpers)
```

Each domain file defines its own route constants and helper functions.

### Why Domain-Based Organization?

1. **Scalability** - Each domain is self-contained
2. **Discoverability** - Easy to find all routes for a domain
3. **Type safety** - Domain-specific helper functions
4. **Testing** - Domain routes tested independently

## Simple Domains (API Endpoints)

Some domains have simple routes with no parameters. The health check is a good example:

**`internal/http/routes/health.go`:**

```go
package routes

const (
    APIHealth = "/api/health"
)
```

**Characteristics:**

- No parameters (static route)
- JSON response (API endpoint)
- Prefixed with `/api/` for clarity
- No helper functions needed

**Usage:**

```go
func (s *Server) routes() {
    s.router.Get(health.APIHealth, s.handleHealthCheck())
}
```

## Complex Domains (HYPERMEDIA Routes)

Most domains serve HTML and have parameterized routes. The user domain demonstrates the full pattern:

**`internal/http/routes/users.go`:**

```go
package routes

import (
    "fmt"
    "net/url"
)

// Private to prevent external dependencies on this implementation detail.
const userSlug = "users"

// HYPERMEDIA-First Pattern (Default for all generated resources):
// Routes serve HTML via templ and include form routes (/new, /edit) for RESTful HTML.
//
// API Alternative: Use --api flag to generate JSON routes with /api prefix and no forms.
//
// Type-Safe URL Generation:
// Use typed helpers (UserShowURL, UserEditURL, etc.) instead of manual string concatenation.
// Provides compile-time safety and automatic URL encoding to prevent injection attacks.
const (
    UserIndex  = "/" + userSlug
    UserShow   = "/" + userSlug + "/:" + userSlug
    UserNew    = "/" + userSlug + "/new"
    UserCreate = "/" + userSlug
    UserEdit   = "/" + userSlug + "/:" + userSlug + "/edit"
    UserUpdate = "/" + userSlug + "/:" + userSlug
    UserDelete = "/" + userSlug + "/:" + userSlug
)

// Exists to reduce code duplication across typed helpers while ensuring
// consistent URL encoding. Prefer the typed helper functions (UserShowURL, etc.)
// for type safety and clarity.
func RouteURL(route string, params ...string) string {
    if len(params) == 0 {
        return route
    }

    result := route
    for i := 0; i < len(params); i += 2 {
        if i+1 >= len(params) {
            break
        }
        key := params[i]
        value := params[i+1]
        placeholder := ":" + key
        result = replaceFirst(result, placeholder, url.PathEscape(value))
    }
    return result
}

// replaceFirst avoids importing the strings package for a single use of strings.Replace(s, old, new, 1).
func replaceFirst(s, old, new string) string {
    idx := 0
    for i := 0; i <= len(s)-len(old); i++ {
        if s[i:i+len(old)] == old {
            idx = i
            return s[:idx] + new + s[idx+len(old):]
        }
    }
    return s
}

func UserIndexURL() string {
    return UserIndex
}

func UserShowURL(username string) string {
    return RouteURL(UserShow, userSlug, username)
}

func UserNewURL() string {
    return UserNew
}

func UserCreateURL() string {
    return UserCreate
}

func UserEditURL(username string) string {
    return RouteURL(UserEdit, userSlug, username)
}

func UserUpdateURL(username string) string {
    return RouteURL(UserUpdate, userSlug, username)
}

func UserDeleteURL(username string) string {
    return RouteURL(UserDelete, userSlug, username)
}
```

**Characteristics:**

- **Slug constant** - DRY principle for parameter name
- **Route constants** - RESTful HYPERMEDIA pattern
- **Form routes** - `/new` and `/edit` for HTML forms
- **URL encoding** - Automatic via `url.PathEscape`
- **Type safety** - Typed helper functions
- **HYPERMEDIA routes** - No `/api/` prefix

## Route Constants Pattern

Route constants provide compile-time safety and prevent typos:

**Benefits:**

1. **Compile-time errors** - Typos caught by compiler
2. **Refactoring safety** - IDE renames all usages
3. **Single source of truth** - Route defined once
4. **Documentation** - Constants self-document available routes

**Pattern:**

```go
const (
    UserIndex  = "/users"                    // List all users
    UserShow   = "/users/:users"             // Show specific user
    UserNew    = "/users/new"                // New user form
    UserCreate = "/users"                    // Create user (POST)
    UserEdit   = "/users/:users/edit"        // Edit user form
    UserUpdate = "/users/:users"             // Update user (PUT/PATCH)
    UserDelete = "/users/:users"             // Delete user (DELETE)
)
```

## Slug Constants

Slug constants keep parameter names consistent:

```go
const userSlug = "users"

const (
    UserShow = "/" + userSlug + "/:" + userSlug  // "/users/:users"
    UserEdit = "/" + userSlug + "/:" + userSlug + "/edit"  // "/users/:users/edit"
)
```

**Why not `:id`?**

HYPERMEDIA routes use readable identifiers in URLs:

- `/users/:users` → `/users/johndoe` ✓ (readable, SEO-friendly)
- `/users/:id` → `/users/12345` ✗ (opaque, not SEO-friendly)

## RouteURL Helper Pattern

Each domain includes a `RouteURL` helper function for parameter substitution:

```go
func RouteURL(route string, params ...string) string {
    if len(params) == 0 {
        return route
    }

    result := route
    for i := 0; i < len(params); i += 2 {
        if i+1 >= len(params) {
            break
        }
        key := params[i]
        value := params[i+1]
        placeholder := ":" + key
        result = replaceFirst(result, placeholder, url.PathEscape(value))
    }
    return result
}
```

**Features:**

- **URL encoding** - Automatic via `url.PathEscape`
- **Multiple parameters** - Supports any number of params
- **Injection prevention** - Spaces become `%20`, `@` becomes `%40`, etc.

**Usage:**

```go
// Simple case
url := RouteURL(UserIndex)  // "/users"

// With parameters
url := RouteURL(UserShow, "users", "johndoe")  // "/users/johndoe"

// Special characters are encoded
url := RouteURL(UserShow, "users", "user@example.com")  // "/users/user%40example.com"
```

## Typed Helper Functions

Domain files provide type-safe helper functions:

```go
func UserShowURL(username string) string {
    return RouteURL(UserShow, userSlug, username)
}

func UserEditURL(username string) string {
    return RouteURL(UserEdit, userSlug, username)
}
```

**Benefits:**

1. **Type safety** - Compiler enforces correct parameter types
2. **IDE autocomplete** - Discoverability of available routes
3. **Refactoring** - Rename parameters safely
4. **Less error-prone** - No manual parameter name typos

**Usage in handlers:**

```go
func (h *UserHandler) HandleShow(w http.ResponseWriter, r *http.Request) {
    username := chi.URLParam(r, "users")
    user, err := h.service.GetByUsername(r.Context(), username)
    if err != nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    // Use helper to generate URLs in templates
    editURL := users.UserEditURL(user.Username)
    deleteURL := users.UserDeleteURL(user.Username)

    // Render templ template with URLs
    component := views.UserProfile(user, editURL, deleteURL)
    component.Render(r.Context(), w)
}
```

## Route Registration

Routes are registered in `internal/http/routes.go`:

```go
package http

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "yourapp/internal/http/handlers"
    "yourapp/internal/http/routes/health"
    "yourapp/internal/http/routes/users"
    httpmiddleware "yourapp/internal/http/middleware"
)

func (s *Server) routes() {
    s.router.Use(middleware.RequestID)
    s.router.Use(middleware.RealIP)
    s.router.Use(httpmiddleware.Logging(s.logger))
    s.router.Use(middleware.Recoverer)

    // Health check (no auth required)
    s.router.Get(health.APIHealth, s.handleHealthCheck())

    // User routes (HYPERMEDIA)
    userHandler := handlers.NewUserHandler(s.userService, s.logger)
    s.router.Get(users.UserIndex, userHandler.HandleIndex)
    s.router.Get(users.UserShow, userHandler.HandleShow)
    s.router.Get(users.UserNew, userHandler.HandleNew)
    s.router.Post(users.UserCreate, userHandler.HandleCreate)
    s.router.Get(users.UserEdit, userHandler.HandleEdit)
    s.router.Post(users.UserUpdate, userHandler.HandleUpdate)
    s.router.Post(users.UserDelete, userHandler.HandleDelete)
}
```

**Pattern:**

1. Import domain route packages (`routes/health`, `routes/users`)
2. Use route constants from imported packages
3. Register with appropriate HTTP methods
4. HYPERMEDIA routes use GET for forms, POST for mutations

## Testing Routes

Test route constants and helpers to ensure correctness:

**`internal/http/routes/users_test.go`:**

```go
package routes

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestUserRoutes(t *testing.T) {
    t.Run("route constants have correct values", func(t *testing.T) {
        assert.Equal(t, "/users", UserIndex)
        assert.Equal(t, "/users/:users", UserShow)
        assert.Equal(t, "/users/new", UserNew)
        assert.Equal(t, "/users", UserCreate)
        assert.Equal(t, "/users/:users/edit", UserEdit)
        assert.Equal(t, "/users/:users", UserUpdate)
        assert.Equal(t, "/users/:users", UserDelete)
    })

    t.Run("routes follow HYPERMEDIA pattern", func(t *testing.T) {
        // HYPERMEDIA routes use readable slugs like :users (not :id)
        assert.Contains(t, UserShow, ":users")
        assert.Contains(t, UserEdit, ":users")
        assert.Contains(t, UserUpdate, ":users")
        assert.Contains(t, UserDelete, ":users")
    })

    t.Run("routes do not contain /api/ prefix", func(t *testing.T) {
        // HYPERMEDIA routes serve HTML, not JSON APIs
        assert.NotContains(t, UserIndex, "/api/")
        assert.NotContains(t, UserShow, "/api/")
        assert.NotContains(t, UserNew, "/api/")
        assert.NotContains(t, UserCreate, "/api/")
        assert.NotContains(t, UserEdit, "/api/")
        assert.NotContains(t, UserUpdate, "/api/")
        assert.NotContains(t, UserDelete, "/api/")
    })

    t.Run("routes include form routes for HYPERMEDIA", func(t *testing.T) {
        // HYPERMEDIA patterns include /new and /edit for HTML forms
        assert.Contains(t, UserNew, "/new")
        assert.Contains(t, UserEdit, "/edit")
    })

    t.Run("userSlug constant has correct value", func(t *testing.T) {
        assert.Equal(t, "users", userSlug)
    })
}

func TestRouteURL(t *testing.T) {
    tests := []struct {
        name     string
        route    string
        params   []string
        expected string
    }{
        {
            name:     "no parameters",
            route:    "/users",
            params:   nil,
            expected: "/users",
        },
        {
            name:     "single parameter",
            route:    "/users/:users",
            params:   []string{"users", "john"},
            expected: "/users/john",
        },
        {
            name:     "URL encodes spaces",
            route:    "/users/:users",
            params:   []string{"users", "john doe"},
            expected: "/users/john%20doe",
        },
        {
            name:     "URL encodes special characters",
            route:    "/users/:users",
            params:   []string{"users", "user@example.com"},
            expected: "/users/user%40example.com",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := RouteURL(tt.route, tt.params...)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestUserHelpers(t *testing.T) {
    t.Run("UserIndexURL returns correct path", func(t *testing.T) {
        assert.Equal(t, "/users", UserIndexURL())
    })

    t.Run("UserShowURL with normal username", func(t *testing.T) {
        assert.Equal(t, "/users/john", UserShowURL("john"))
    })

    t.Run("UserShowURL with special characters", func(t *testing.T) {
        assert.Equal(t, "/users/user%40example.com", UserShowURL("user@example.com"))
    })

    t.Run("UserEditURL with normal username", func(t *testing.T) {
        assert.Equal(t, "/users/john/edit", UserEditURL("john"))
    })
}
```

## API Routes (Rare Exceptions)

JSON APIs are the exception in Tracks applications. Use the `--api` flag when generating resources that need JSON endpoints:

```bash
tracks generate resource webhooks --api
```

**Generated API routes:**

```go
const (
    APIWebhooks       = "/api/webhooks"
    APIWebhookShow    = "/api/webhooks/:id"
    APIWebhookCreate  = "/api/webhooks"
    APIWebhookUpdate  = "/api/webhooks/:id"
    APIWebhookDelete  = "/api/webhooks/:id"
)
```

**Characteristics:**

- Prefixed with `/api/` for clarity
- Use `:id` instead of readable slugs (IDs are fine for APIs)
- No form routes (`/new`, `/edit` not needed for JSON)
- Return JSON responses
- Typically used for:
  - Health/metrics endpoints
  - Webhook receivers
  - Third-party integrations
  - Mobile app backends

## Best Practices

### 1. Default to HYPERMEDIA

Always start with HTML-serving routes. Only use `--api` if you truly need JSON responses.

### 2. Use Domain-Based Files

Create separate route files for each domain:

```text
routes/
├── users.go      # User domain
├── posts.go      # Post domain
├── comments.go   # Comment domain
└── health.go     # Health checks
```

### 3. Use Slug Constants

Keep parameter names consistent with slug constants:

```go
const postSlug = "posts"

const (
    PostShow = "/" + postSlug + "/:" + postSlug
)
```

### 4. Always Use Typed Helpers

Never manually construct URLs:

```go
// Good
url := users.UserShowURL(username)

// Bad
url := "/users/" + username  // No URL encoding! Injection risk!
```

### 5. Test Your Routes

Write tests for route constants and helpers to catch typos and ensure URL encoding works.

### 6. Include Form Routes

HYPERMEDIA applications need form routes:

```go
const (
    UserNew    = "/users/new"        // GET: Show form
    UserCreate = "/users"            // POST: Process form
    UserEdit   = "/users/:users/edit"  // GET: Show edit form
    UserUpdate = "/users/:users"     // POST: Process update
)
```

### 7. Use Readable Slugs

HYPERMEDIA routes benefit from readable URLs:

- `/users/johndoe` ✓
- `/posts/getting-started-with-go` ✓
- `/users/123` ✗ (opaque, not SEO-friendly)

## Next Steps

- [Architecture Overview](./architecture-overview.md) - High-level system design
- [Layer Guide](./layer-guide.md) - Detailed explanation of each layer
- [Patterns](./patterns.md) - Common implementation patterns
- [Testing Guide](./testing.md) - Testing strategies
