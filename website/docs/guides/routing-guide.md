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
import "yourapp/internal/http/routes"

func (s *Server) routes() {
    s.router.Get(routes.APIHealth, s.handleHealthCheck())
}
```

## Complex Domains (HYPERMEDIA Routes)

Most domains serve HTML and have parameterized routes. The user domain demonstrates the full pattern.

**Note:** The examples below are taken directly from Tracks' production templates (`users.go.tmpl`, `users_test.go.tmpl`), ensuring they represent actual generated code.

**`internal/http/routes/users.go`:**

```go
package routes

import (
    "net/url"
)

// UserSlugParam is exported so handlers can extract parameters without magic strings.
// usersPath remains unexported as it's an internal routing detail.
const (
    usersPath     = "users"
    UserSlugParam = "username"
)

// HYPERMEDIA-First Pattern (Default for all generated resources):
// Routes serve HTML via templ and include form routes (/new, /edit) for RESTful HTML.
//
// API Alternative: Use --api flag to generate JSON routes with /api prefix and no forms.
//
// Type-Safe URL Generation:
// Use typed helpers (UserShowURL, UserEditURL, etc.) instead of manual string concatenation.
// Provides compile-time safety and automatic URL encoding to prevent injection attacks.
const (
    UserIndex  = "/" + usersPath
    UserShow   = "/" + usersPath + "/:" + UserSlugParam
    UserNew    = "/" + usersPath + "/new"
    UserCreate = "/" + usersPath
    UserEdit   = "/" + usersPath + "/:" + UserSlugParam + "/edit"
    UserUpdate = "/" + usersPath + "/:" + UserSlugParam
    UserDelete = "/" + usersPath + "/:" + UserSlugParam
)

// RouteURL is a low-level helper. Use typed functions (UserShowURL, etc.) for better type safety.
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

// replaceFirst avoids importing strings package to keep route files lightweight.
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
    return RouteURL(UserShow, UserSlugParam, username)
}

func UserNewURL() string {
    return UserNew
}

func UserCreateURL() string {
    return UserCreate
}

func UserEditURL(username string) string {
    return RouteURL(UserEdit, UserSlugParam, username)
}

func UserUpdateURL(username string) string {
    return RouteURL(UserUpdate, UserSlugParam, username)
}

func UserDeleteURL(username string) string {
    return RouteURL(UserDelete, UserSlugParam, username)
}
```

**Characteristics:**

- **Path and parameter constants** - Separate constants for base path and parameter name (no magic strings)
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
    UserShow   = "/users/:username"             // Show specific user
    UserNew    = "/users/new"                // New user form
    UserCreate = "/users"                    // Create user (POST)
    UserEdit   = "/users/:username/edit"        // Edit user form
    UserUpdate = "/users/:username"             // Update user (PUT/PATCH)
    UserDelete = "/users/:username"             // Delete user (DELETE)
)
```

## Path and Parameter Constants

Separate constants for paths and parameters prevent magic strings and keep the code maintainable:

```go
const (
    usersPath     = "users"
    UserSlugParam = "username"
)

const (
    UserShow = "/" + usersPath + "/:" + UserSlugParam  // "/users/:username"
    UserEdit = "/" + usersPath + "/:" + UserSlugParam + "/edit"  // "/users/:username/edit"
)
```

**Why separate constants?**

- **Path constant** (`usersPath`) - The base path segment in the URL (unexported, internal detail)
- **Parameter constant** (`UserSlugParam`) - The parameter name used in route patterns, RouteURL calls, and handlers (exported so handlers can use `chi.URLParam(r, routes.UserSlugParam)`)
- **No magic strings** - Both are referenced by constant, not hardcoded strings
- **Consistent naming** - All resources follow `{Resource}SlugParam` pattern (e.g., `UserSlugParam`, `PostSlugParam`)

**Why not `:id`?**

HYPERMEDIA routes use readable identifiers (slugs) in URLs:

- `/users/:username` → `/users/johndoe` ✓ (readable, SEO-friendly)
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

// With parameters - use constant, not magic string
url := RouteURL(UserShow, routes.UserSlugParam, "johndoe")  // "/users/johndoe"

// Special characters are encoded
url := RouteURL(UserShow, routes.UserSlugParam, "user@example.com")  // "/users/user%40example.com"
```

## Typed Helper Functions

Domain files provide type-safe helper functions:

```go
func UserShowURL(username string) string {
    return RouteURL(UserShow, UserSlugParam, username)
}

func UserEditURL(username string) string {
    return RouteURL(UserEdit, UserSlugParam, username)
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
    username := chi.URLParam(r, routes.UserSlugParam)
    user, err := h.service.GetByUsername(r.Context(), username)
    if err != nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    // Use helper to generate URLs in templates
    editURL := routes.UserEditURL(user.Username)
    deleteURL := routes.UserDeleteURL(user.Username)

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
    "yourapp/internal/http/routes"
    httpmiddleware "yourapp/internal/http/middleware"
)

func (s *Server) routes() {
    s.router.Use(middleware.RequestID)
    s.router.Use(middleware.RealIP)
    s.router.Use(httpmiddleware.Logging(s.logger))
    s.router.Use(middleware.Recoverer)

    // Health check (no auth required)
    s.router.Get(routes.APIHealth, s.handleHealthCheck())

    // User routes (HYPERMEDIA)
    userHandler := handlers.NewUserHandler(s.userService, s.logger)
    s.router.Get(routes.UserIndex, userHandler.HandleIndex)
    s.router.Get(routes.UserShow, userHandler.HandleShow)
    s.router.Get(routes.UserNew, userHandler.HandleNew)
    s.router.Post(routes.UserCreate, userHandler.HandleCreate)
    s.router.Get(routes.UserEdit, userHandler.HandleEdit)
    s.router.Post(routes.UserUpdate, userHandler.HandleUpdate)
    s.router.Post(routes.UserDelete, userHandler.HandleDelete)
}
```

**Pattern:**

1. Import the routes package (`"yourapp/internal/http/routes"`)
2. Use route constants from the package (e.g., `routes.APIHealth`, `routes.UserIndex`, `routes.UserShow`)
3. Register with appropriate HTTP methods
4. HYPERMEDIA routes use GET for forms, POST for mutations

**Note:** All domain route files (health.go, users.go, etc.) are in the same `routes` package. They're organized into separate files for maintainability, but share the same package namespace.

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
        assert.Equal(t, "/users/:username", UserShow)
        assert.Equal(t, "/users/new", UserNew)
        assert.Equal(t, "/users", UserCreate)
        assert.Equal(t, "/users/:username/edit", UserEdit)
        assert.Equal(t, "/users/:username", UserUpdate)
        assert.Equal(t, "/users/:username", UserDelete)
    })

    t.Run("routes follow HYPERMEDIA pattern", func(t *testing.T) {
        // HYPERMEDIA routes use readable slugs like :username (not :id)
        assert.Contains(t, UserShow, ":username")
        assert.Contains(t, UserEdit, ":username")
        assert.Contains(t, UserUpdate, ":username")
        assert.Contains(t, UserDelete, ":username")
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

    t.Run("parameter constants have correct values", func(t *testing.T) {
        assert.Equal(t, "users", usersPath)
        assert.Equal(t, "username", UserSlugParam)
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
            route:    "/users/:username",
            params:   []string{"username", "john"},
            expected: "/users/john",
        },
        {
            name:     "multiple parameters",
            route:    "/users/:username/posts/:posts",
            params:   []string{"username", "john", "posts", "hello-world"},
            expected: "/users/john/posts/hello-world",
        },
        {
            name:     "URL encodes spaces",
            route:    "/users/:username",
            params:   []string{"username", "john doe"},
            expected: "/users/john%20doe",
        },
        {
            name:     "URL encodes special characters",
            route:    "/users/:username",
            params:   []string{"username", "user@example.com"},
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

### 3. Use Path and Parameter Constants

Keep paths and parameter names consistent with separate constants:

```go
const (
    postsPath     = "posts"
    postSlugParam = "post_slug"
)

const (
    PostShow = "/" + postsPath + "/:" + postSlugParam
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
    UserEdit   = "/users/:username/edit"  // GET: Show edit form
    UserUpdate = "/users/:username"     // POST: Process update
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
