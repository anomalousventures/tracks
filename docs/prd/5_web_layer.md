# Web Layer

**[← Back to Summary](./0_summary.md)**

## Overview

The web layer in Tracks uses **Chi** as the router (a framework choice) to handle HTTP requests with composable middleware chains. The design emphasizes hypermedia-driven patterns with HTMX, semantic URLs, and clear separation between GET (forms) and POST (actions) handlers.

## Goals

- Fast, composable routing with Chi
- Comprehensive middleware stack for security and observability
- Automatic compression and caching for assets
- Request ID generation for tracing
- Graceful shutdown with in-flight request handling
- Compile-time safe route references

## User Stories

- As a developer, I want composable middleware chains
- As a developer, I want automatic request logging and metrics
- As a user, I want fast page loads with compressed assets
- As a DevOps engineer, I want request IDs for debugging
- As a developer, I want graceful shutdown during deployments
- As a developer, I want compile-time errors when I reference non-existent routes
- As a developer, I want separate handlers for GET and POST on the same path
- As a user, I want readable URLs like `/u/johndoe` not `/users/uuid`
- As a developer, I want generated helper functions for route paths

## Hypermedia Route Patterns

Routes are designed around user actions and page transitions, not CRUD operations. This aligns with the hypermedia-driven architecture where the server controls application state through HTML responses.

```go
// internal/http/routes/routes.go
package routes

const (
    // Pages (nouns - GET only)
    Home          = "/"
    About         = "/about"
    Dashboard     = "/dashboard"

    // Authentication flows (GET shows form, POST processes)
    Login         = "/login"
    Register      = "/register"
    VerifyOTP     = "/verify-otp"
    Logout        = "/logout"          // POST only

    // User profiles (slugs not IDs)
    UserProfile   = "/u/{username}"    // Public profile
    EditProfile   = "/settings/profile" // Current user only

    // Content with slugs
    PostView      = "/p/{slug}"        // View post
    PostNew       = "/write"           // New post form
    PostEdit      = "/p/{slug}/edit"   // Edit form

    // Actions (POST only)
    CreatePost    = "/create-post"
    UpdatePost    = "/update-post"
    DeletePost    = "/delete-post"
    FollowUser    = "/follow-user"
    UnfollowUser  = "/unfollow-user"

    // Search/Discovery
    Search        = "/search"          // GET with ?q=
    Explore       = "/explore"         // GET with filters

    // Static
    Assets        = "/assets/*"
    Robots        = "/robots.txt"
    Sitemap       = "/sitemap.xml"
)

// Generated route helpers (compile-time safe)
func UserProfilePath(username string) string {
    return "/u/" + username
}

func PostPath(slug string) string {
    return "/p/" + slug
}

func PostEditPath(slug string) string {
    return "/p/" + slug + "/edit"
}

// For use in templ templates
func UserProfileURL(username string) templ.SafeURL {
    return templ.URL("/u/" + username)
}

func PostURL(slug string) templ.SafeURL {
    return templ.URL("/p/" + slug)
}
```

## Router Setup

The server uses a struct-based pattern with dependency injection and builder methods for testability.

### Server Structure (internal/http/server.go)

```go
package http

import (
    "context"
    "net/http"
    "github.com/go-chi/chi/v5"
    "myapp/internal/interfaces"
)

type Server struct {
    router chi.Router
    config *Config

    // Injected services
    healthService interfaces.HealthService
    postService   interfaces.PostService
}

func NewServer(cfg *Config) *Server {
    return &Server{
        router: chi.NewRouter(),
        config: cfg,
    }
}

// Dependency injection methods
func (s *Server) WithHealthService(svc interfaces.HealthService) *Server {
    s.healthService = svc
    return s
}

func (s *Server) WithPostService(svc interfaces.PostService) *Server {
    s.postService = svc
    return s
}

// Called after all dependencies injected
func (s *Server) RegisterRoutes() *Server {
    s.routes()
    return s
}

func (s *Server) ListenAndServe() error {
    srv := &http.Server{
        Addr:         s.config.Port,
        Handler:      s.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed:", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return srv.Shutdown(ctx)
}
```

### Routes Registration (internal/http/routes.go)

⚠️ **CRITICAL: Middleware Order** - The order of middleware registration matters!

All routes in a single file with **markers for incremental generation**:

```go
package http

import (
    "myapp/internal/routes"
)

func (s *Server) routes() {
    // Core middleware (order matters!)
    s.router.Use(middleware.RequestID)
    s.router.Use(middleware.RealIP)
    s.router.Use(middleware.Logger)
    s.router.Use(middleware.Recoverer)

    // OpenTelemetry
    s.router.Use(otelchi.Middleware("tracks-app",
        otelchi.WithChiRoutes(s.router),
        otelchi.WithRequestMethodInSpanName(true),
    ))

    // Rate limiting
    s.router.Use(httprate.LimitByIP(100, time.Minute))

    // ⚠️ CRITICAL: Security middleware order!
    s.router.Use(middleware.WithNonce())      // Must precede SecureHeaders
    s.router.Use(middleware.SecureHeaders())  // CSP, HSTS, referrer-policy

    // Application middleware
    s.router.Use(middleware.Session())
    s.router.Use(middleware.I18n(i18n.Bundle))
    s.router.Use(middleware.CacheHeaders())
    s.router.Use(middleware.Compress(5))

    // API routes (JSON only)
    // TRACKS:API_ROUTES:BEGIN
    s.router.Get(routes.APIHealth, s.handleHealthCheck())
    // TRACKS:API_ROUTES:END

    // Public web routes (HTML)
    // TRACKS:WEB_ROUTES:BEGIN
    s.router.Get(routes.Home, s.handleHome())
    s.router.Get(routes.About, s.handleAbout())
    // TRACKS:WEB_ROUTES:END

    // Protected routes (requires auth)
    s.router.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth)

        // TRACKS:PROTECTED_ROUTES:BEGIN
        r.Get(routes.Dashboard, s.handleDashboard())
        // TRACKS:PROTECTED_ROUTES:END
    })

    // SEO routes (plaintext/XML, not under /api)
    s.router.Get("/robots.txt", s.handleRobots())
    s.router.Get("/sitemap.xml", s.handleSitemap())
}
```

### Main Wiring (cmd/server/main.go)

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // TRACKS:DB:BEGIN
    database, err := db.New(cfg.DatabaseURL)
    if err != nil {
        return fmt.Errorf("connect db: %w", err)
    }
    defer database.Close()
    // TRACKS:DB:END

    // TRACKS:REPOSITORIES:BEGIN
    postRepo := posts.NewRepository(database)
    // TRACKS:REPOSITORIES:END

    // TRACKS:SERVICES:BEGIN
    healthService := health.NewService()
    postService := posts.NewService(postRepo)
    // TRACKS:SERVICES:END

    srv := http.NewServer(cfg).
        WithHealthService(healthService).
        WithPostService(postService).
        RegisterRoutes()

    return srv.ListenAndServe()
}
```

### API Design Policy

- **`/api/*` endpoints:** Return **JSON only** (no HTML). Keep this surface minimal and stable.
- **Top-level exceptions:** `/robots.txt`, `/sitemap.xml`, `/LLMs.txt` (plaintext/XML for SEO, not under /api)
- **Everything else:** Returns HTML for hypermedia-driven interactions

## Middleware Stack

### Core Middleware

```go
// internal/http/middleware/core.go

// RequestID generates a unique ID for each request
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        requestID := r.Header.Get("X-Request-Id")
        if requestID == "" {
            requestID = nanoid.New()
        }

        ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)
        w.Header().Set("X-Request-Id", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Logger logs all HTTP requests with structured logging
func Logger(next http.Handler) http.Handler {
    return httplog.Handler(zerolog.New(os.Stdout))(next)
}
```

### Cache Headers Middleware

```go
// internal/http/middleware/cache.go
func CacheHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            path := r.URL.Path

            // Set cache headers based on path patterns
            if strings.HasPrefix(path, "/dashboard") || strings.HasPrefix(path, "/settings") {
                // Private, no-cache for user-specific content
                w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
                w.Header().Set("Pragma", "no-cache")
                next.ServeHTTP(w, r)
                return
            }

            if strings.HasPrefix(path, "/assets/") {
                // Immutable - hashed filenames
                w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
                w.Header().Set("CDN-Cache-Control", "max-age=31536000")
                next.ServeHTTP(w, r)
                return
            }

            // Default: short cache for public content
            w.Header().Set("Cache-Control", "public, max-age=60")
            next.ServeHTTP(w, r)
        })
    }
}
```

## Handler Patterns

### Hypermedia-Driven Handlers

Handlers follow a consistent pattern:

1. **GET handlers** display forms and pages
2. **POST handlers** process actions and redirect
3. **Never expose internal IDs** - use slugs and usernames
4. **Return HTML by default** - JSON only for `/api/*` endpoints

```go
// internal/http/handlers/auth_handler.go

// Separate handlers for GET (form) and POST (action)
func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
    // Check if already logged in
    if userID := h.sessions.GetString(r.Context(), "user_id"); userID != "" {
        http.Redirect(w, r, routes.Dashboard, http.StatusSeeOther)
        return
    }

    views.LoginForm(r.Context()).Render(r.Context(), w)
}

func (h *AuthHandler) ProcessLogin(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")

    // Rate limit login attempts
    if !h.rateLimiter.AllowLogin(email) {
        w.WriteHeader(http.StatusTooManyRequests)
        views.LoginForm(r.Context(),
            WithError("Too many attempts. Please try again later.")).
            Render(r.Context(), w)
        return
    }

    // Send OTP
    if err := h.auth.SendOTP(r.Context(), email); err != nil {
        w.WriteHeader(http.StatusUnprocessableEntity)
        views.LoginForm(r.Context(),
            WithError("Failed to send code")).
            Render(r.Context(), w)
        return
    }

    // Store email in session for OTP verification
    h.sessions.Put(r.Context(), "otp_email", email)

    // Redirect to OTP verification
    http.Redirect(w, r, routes.VerifyOTP, http.StatusSeeOther)
}
```

### Content Handlers

```go
// internal/http/handlers/post_handler.go

// Content handlers work with slugs, not IDs
func (h *PostHandler) Show(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, routes.PostSlugParam)  // Use exported constant, no magic strings

    post, err := h.postService.GetBySlug(r.Context(), slug)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Never expose internal IDs
    views.Post(PostView{
        Slug:        post.Slug,
        Title:       post.Title,
        Content:     post.Content,
        AuthorName:  post.Author.Username,
        PublishedAt: post.PublishedAt,
    }).Render(r.Context(), w)
}

func (h *PostHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, routes.PostSlugParam)  // Use exported constant, no magic strings
    userID := r.Context().Value("user_id").(string)

    post, err := h.postService.GetBySlug(r.Context(), slug)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Check ownership or permission
    if post.AuthorID != userID && !h.canEdit(r.Context(), userID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    views.PostEditForm(PostEditView{
        Slug:    post.Slug,
        Title:   post.Title,
        Content: post.Content,
    }).Render(r.Context(), w)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)

    // Parse form
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    slug := r.FormValue("slug")
    title := r.FormValue("title")
    content := r.FormValue("content")

    // Update post
    post, err := h.postService.Update(r.Context(), slug, UpdatePostDTO{
        Title:   title,
        Content: content,
        UserID:  userID,
    })
    if err != nil {
        // Re-render form with error
        w.WriteHeader(http.StatusUnprocessableEntity)
        views.PostEditForm(PostEditView{
            Slug:    slug,
            Title:   title,
            Content: content,
            Error:   "Failed to update post",
        }).Render(r.Context(), w)
        return
    }

    // Redirect to updated post
    http.Redirect(w, r, routes.PostPath(post.Slug), http.StatusSeeOther)
}
```

## Graceful Shutdown

```go
// cmd/server/main.go
func main() {
    // ... setup ...

    srv := &http.Server{
        Addr:         cfg.Port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed to start:", err)
        }
    }()

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server shutdown complete")
}
```

## Best Practices

1. **Middleware Order Matters** - Security middleware must be in the correct order
2. **Separate GET and POST Handlers** - Different concerns, different handlers
3. **Use Slugs, Not IDs** - Human-readable URLs improve UX and SEO
4. **Redirect After POST** - Prevent duplicate submissions (PRG pattern)
5. **Rate Limit Sensitive Endpoints** - Login, registration, password reset
6. **Cache Static Assets Aggressively** - Use content hashing for cache busting
7. **Add Request IDs Early** - Essential for distributed tracing
8. **Handle Errors Gracefully** - Re-render forms with errors, don't lose user input

## Testing

```go
// internal/http/handlers/post_handler_test.go
func TestPostHandler_Show(t *testing.T) {
    // Setup
    handler := NewPostHandler(mockService)

    // Create request
    req := httptest.NewRequest("GET", "/p/my-first-post", nil)
    req = req.WithContext(chi.NewRouteContext())
    chi.URLParam(req, routes.PostSlugParam, "my-first-post")  // Use constant in test setup too

    // Record response
    rr := httptest.NewRecorder()

    // Execute
    handler.Show(rr, req)

    // Assert
    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "My First Post")
}
```

## Next Steps

- Continue to [Security →](./6_security.md)
- Back to [← Authorization & RBAC](./4_authorization_rbac.md)
- Return to [Summary](./0_summary.md)
