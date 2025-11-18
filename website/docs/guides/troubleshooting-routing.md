# Troubleshooting Routing

This guide covers common routing issues in Tracks applications and how to debug them. All Tracks applications use the [Chi router](https://github.com/go-chi/chi), which provides a lightweight, idiomatic router for building Go HTTP services.

## Route Not Found (404)

When you get a 404 error, the route isn't registered or the pattern doesn't match.

### Debug with chi.Walk()

The `chi.Walk()` function is your best friend for debugging routing issues. It prints all registered routes:

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()

    // Register your routes...

    // Print all registered routes for debugging
    chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
        fmt.Printf("[%s] %s\n", method, route)
        return nil
    })

    http.ListenAndServe(":8080", r)
}
```

**Common causes:**

1. **Route not registered** - Check `internal/http/routes.go` to ensure the route is registered
2. **Typo in route constant** - Verify the constant value matches the expected pattern
3. **Middleware blocking request** - A middleware earlier in the chain may be returning early
4. **Route order matters** - More specific routes must be registered before catch-all patterns

### Route Registration Checklist

```go
// 1. Define constant in internal/http/routes/users.go
const UserShow = "/users/:username"

// 2. Register in internal/http/routes.go
s.router.Get(routes.UserShow, userHandler.HandleShow)

// 3. Verify handler exists
func (h *UserHandler) HandleShow(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### Testing Routes with curl

```bash
# Test route exists
curl -i http://localhost:8080/users/johndoe

# Expected: 200 OK or appropriate response
# If 404: Route not registered or pattern mismatch
```

## Middleware Not Executing

Middleware must be registered in the correct order and must call `next.ServeHTTP()` to continue the chain.

### Middleware Registration Order

Middleware executes in the order it's registered:

```go
func (s *Server) routes() {
    // Global middleware (runs for ALL routes)
    s.router.Use(middleware.RequestID)      // 1. Runs first
    s.router.Use(middleware.RealIP)         // 2. Then this
    s.router.Use(httpmiddleware.Logging(s.logger))  // 3. Then logging
    s.router.Use(middleware.Recoverer)      // 4. Finally recoverer

    // Routes registered after middleware
    s.router.Get("/users", handler.HandleIndex)
}
```

**Common mistakes:**

1. **Registering routes before middleware** - Routes registered before `.Use()` won't have that middleware
2. **Forgetting to call next.ServeHTTP()** - Breaks the middleware chain
3. **Middleware panic without recovery** - Use `middleware.Recoverer` to catch panics

### Correct Middleware Pattern

```go
func MyMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Do something before handler

            // CRITICAL: Call next handler in chain
            next.ServeHTTP(w, r)

            // Do something after handler (optional)
        })
    }
}
```

### Debugging Middleware Execution

Add logging to verify middleware is running:

```go
func MyMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Debug().
                Str("path", r.URL.Path).
                Msg("MyMiddleware executing")

            next.ServeHTTP(w, r)

            logger.Debug().
                Str("path", r.URL.Path).
                Msg("MyMiddleware complete")
        })
    }
}
```

## Route Parameters Not Found

Chi uses URL parameters with `:param` syntax. If `chi.URLParam()` returns an empty string, the parameter name doesn't match.

### Correct Parameter Extraction

```go
// Route definition in internal/http/routes/users.go
const (
    UserSlugParam = "username"
    UserShow = "/users/:username"  // Must use :username
)

// Handler in internal/http/handlers/user_handler.go
func (h *UserHandler) HandleShow(w http.ResponseWriter, r *http.Request) {
    // Extract parameter - use the constant to avoid typos
    username := chi.URLParam(r, routes.UserSlugParam)

    if username == "" {
        http.Error(w, "Username parameter missing", http.StatusBadRequest)
        return
    }

    // Use username...
}
```

**Common mistakes:**

1. **Typo in parameter name** - `chi.URLParam(r, "user")` when route uses `:username`
2. **Missing colon in route** - `/users/username` instead of `/users/:username`
3. **Wrong route pattern** - Using `chi.URLParam(r, "id")` on a route with `:username`

### Always Use Route Constants

Never use magic strings for parameter names:

```go
// Good - uses constant
username := chi.URLParam(r, routes.UserSlugParam)

// Bad - magic string (typo-prone)
username := chi.URLParam(r, "username")
```

## Route Conflicts

Chi matches routes in the order they're registered. More specific routes must come before catch-all patterns.

### Route Specificity

```go
func (s *Server) routes() {
    // Specific routes FIRST
    s.router.Get("/users/new", handler.HandleNew)       // Must be first
    s.router.Get("/users/:username", handler.HandleShow)   // Then parameterized

    // This won't work if reversed - /:username would catch /new
}
```

**Rule:** Static segments beat parameters, so register static routes first.

### Debugging Route Conflicts

Use `chi.Walk()` to see the order routes are registered:

```go
chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
    fmt.Printf("[%s] %s\n", method, route)
    return nil
})

// Output shows registration order:
// [GET] /users/new
// [GET] /users/:username
// ✓ Correct order
```

## CORS and Preflight Issues

CORS requires proper middleware configuration and understanding of preflight requests.

### CORS Middleware Setup

```go
import (
    "github.com/go-chi/cors"
)

func (s *Server) routes() {
    // CORS middleware must be registered early
    s.router.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Routes...
}
```

### Preflight Requests (OPTIONS)

Browsers send OPTIONS requests before POST/PUT/DELETE. Chi handles this automatically if you register the route with POST/PUT/DELETE methods.

**Common mistake:**

```go
// Bad - only handles POST, OPTIONS requests will 404
s.router.Post("/users", handler.HandleCreate)

// Good - Chi automatically handles OPTIONS for registered methods
s.router.Post("/users", handler.HandleCreate)  // OPTIONS automatically allowed
```

### Debugging CORS with curl

```bash
# Test preflight request
curl -i -X OPTIONS http://localhost:8080/users \
  -H "Origin: https://app.example.com" \
  -H "Access-Control-Request-Method: POST"

# Expected headers in response:
# Access-Control-Allow-Origin: https://app.example.com
# Access-Control-Allow-Methods: POST, OPTIONS
# Access-Control-Allow-Headers: Content-Type, ...
```

## Server Won't Shutdown Gracefully

Graceful shutdown requires proper context handling and timeout configuration.

### Correct Graceful Shutdown Pattern

```go
func (s *Server) Start(ctx context.Context) error {
    srv := &http.Server{
        Addr:    s.addr,
        Handler: s.router,
    }

    // Channel to signal server errors
    serverErrors := make(chan error, 1)

    // Start server in goroutine
    go func() {
        s.logger.Info().Str("addr", s.addr).Msg("Starting server")
        serverErrors <- srv.ListenAndServe()
    }()

    // Wait for interrupt signal or server error
    select {
    case err := <-serverErrors:
        return fmt.Errorf("server error: %w", err)

    case <-ctx.Done():
        s.logger.Info().Msg("Shutdown signal received")

        // Give outstanding requests time to complete
        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        if err := srv.Shutdown(ctx); err != nil {
            s.logger.Error().Err(err).Msg("Graceful shutdown failed")
            return srv.Close()  // Force close
        }

        s.logger.Info().Msg("Server stopped gracefully")
        return nil
    }
}
```

### Common Shutdown Issues

1. **Blocked goroutines** - Background tasks not respecting context cancellation
2. **Short timeout** - Increase shutdown timeout if requests take longer
3. **Database connection not closed** - Ensure DB cleanup in shutdown

### Debugging Shutdown Hangs

Add logging to identify what's blocking:

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

s.logger.Info().Msg("Starting graceful shutdown")

if err := srv.Shutdown(ctx); err != nil {
    s.logger.Error().
        Err(err).
        Msg("Shutdown timeout - some requests didn't complete")
    return srv.Close()
}

s.logger.Info().Msg("All requests completed, server stopped")
```

## Common Chi Router Gotchas

### 1. Trailing Slashes

Chi treats `/users` and `/users/` as different routes by default.

```go
// These are DIFFERENT routes
s.router.Get("/users", handler.HandleIndex)   // /users
s.router.Get("/users/", handler.HandleIndex)  // /users/ (different!)

// Solution: Use middleware to handle trailing slashes
import "github.com/go-chi/chi/v5/middleware"

s.router.Use(middleware.StripSlashes)  // Strips trailing slashes
```

### 2. Method Mismatch

Chi routes are method-specific. GET request to a POST route will 404.

```go
// Route only accepts POST
s.router.Post("/users", handler.HandleCreate)

// GET /users will return 404, not 405 Method Not Allowed
```

### 3. Middleware Scope

Middleware can be global, group-scoped, or route-specific:

```go
// Global - applies to all routes
s.router.Use(middleware.Logger)

// Group-scoped - applies to routes in this group
s.router.Group(func(r chi.Router) {
    r.Use(AuthMiddleware)
    r.Get("/admin", handler.HandleAdmin)
})

// Route-specific - only this route
s.router.With(RateLimitMiddleware).Get("/api/heavy", handler.HandleHeavy)
```

## Debugging Checklist

When debugging routing issues, check in this order:

1. **Is the route registered?**
   - Use `chi.Walk()` to list all routes
   - Check `internal/http/routes.go` for registration

2. **Does the pattern match?**
   - Verify route constant value
   - Check for typos in parameter names
   - Test with curl: `curl -i http://localhost:8080/your/path`

3. **Is middleware blocking?**
   - Add logging to each middleware
   - Verify `next.ServeHTTP()` is called
   - Check middleware order

4. **Are parameters extracted correctly?**
   - Use route constants for parameter names
   - Check for empty string returns from `chi.URLParam()`
   - Verify `:param` syntax in route pattern

5. **Is the method correct?**
   - GET vs POST vs PUT vs DELETE
   - Check browser dev tools for actual method sent

6. **Check response headers**
   - Use curl with `-i` flag to see headers
   - Verify Content-Type is set correctly
   - Check for CORS headers if cross-origin

## Useful Tools

### curl Examples

```bash
# View full request/response including headers
curl -i http://localhost:8080/users

# Follow redirects
curl -L http://localhost:8080/users/new

# Send POST with form data
curl -X POST http://localhost:8080/users \
  -d "username=john&email=john@example.com"

# Send JSON
curl -X POST http://localhost:8080/api/webhooks \
  -H "Content-Type: application/json" \
  -d '{"event":"user.created"}'

# Test CORS preflight
curl -i -X OPTIONS http://localhost:8080/users \
  -H "Origin: https://app.example.com" \
  -H "Access-Control-Request-Method: POST"
```

### Chi Debugging Functions

```go
// Print all registered routes
chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
    fmt.Printf("[%s] %s (middlewares: %d)\n", method, route, len(middlewares))
    return nil
})

// Get current route pattern in handler
routePattern := chi.RouteContext(r.Context()).RoutePattern()
fmt.Printf("Matched route: %s\n", routePattern)
```

## Next Steps

- [Routing Guide](./routing-guide.md) - Comprehensive routing patterns and best practices
- [Architecture Overview](./architecture-overview.md) - High-level system design
- [Patterns](./patterns.md) - Common implementation patterns
- [Testing Guide](./testing.md) - Testing strategies for routes and handlers
