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

⚠️ **CRITICAL: Middleware Order**
The order of middleware registration matters! The nonce middleware MUST be registered before SecureHeaders so that the CSP can include the nonce value.

```go
// internal/http/server.go
func NewServer(cfg config.Config, services *app.Services) *Server {
    r := chi.NewRouter()

    // Core middleware (order matters!)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // OpenTelemetry
    r.Use(otelchi.Middleware("tracks-app",
        otelchi.WithChiRoutes(r),
        otelchi.WithRequestMethodInSpanName(true),
    ))

    // Rate limiting
    r.Use(httprate.LimitByIP(100, time.Minute))

    // ⚠️ CRITICAL: Security middleware order matters!
    r.Use(middleware.WithNonce())      // Must precede SecureHeaders so CSP can include nonce
    r.Use(middleware.SecureHeaders())  // CSP, HSTS, referrer-policy, etc.

    // Application middleware
    r.Use(middleware.Session())
    r.Use(middleware.I18n(i18n.Bundle))
    r.Use(middleware.CacheHeaders())   // Caching policy
    r.Use(middleware.Compress(5))

    // Observability
    r.Use(middleware.Metrics())
    r.Use(middleware.Tracing())

    // Static assets (immutable with hash)
    r.Handle(routes.Assets, http.StripPrefix("/assets", assets.Handler()))

    // Public routes with verb-specific handlers
    r.Get(routes.Home, homeHandler.Index)
    r.Get(routes.About, homeHandler.About)

    // Authentication - different handlers for GET/POST
    r.Get(routes.Login, authHandler.ShowLoginForm)
    r.Post(routes.Login, authHandler.ProcessLogin)

    r.Get(routes.Register, authHandler.ShowRegisterForm)
    r.Post(routes.Register, authHandler.ProcessRegister)

    r.Get(routes.VerifyOTP, authHandler.ShowOTPForm)
    r.Post(routes.VerifyOTP, authHandler.ProcessOTP)

    // Public profiles (with username, not UUID)
    r.Get(routes.UserProfile, profileHandler.ShowPublic)

    // Public content (with slug, not UUID)
    r.Get(routes.PostView, postHandler.Show)

    // Search/discovery
    r.Get(routes.Search, searchHandler.Search)
    r.Get(routes.Explore, exploreHandler.Browse)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth)

        // Dashboard
        r.Get(routes.Dashboard, dashboardHandler.Show)

        // Profile management (current user)
        r.Get(routes.EditProfile, profileHandler.EditOwn)
        r.Post(routes.EditProfile, profileHandler.UpdateOwn)

        // Content creation
        r.Get(routes.PostNew, postHandler.ShowCreateForm)
        r.Post(routes.CreatePost, postHandler.Create)

        // Content editing (ownership checked in handler)
        r.Get(routes.PostEdit, postHandler.ShowEditForm)
        r.Post(routes.UpdatePost, postHandler.Update)
        r.Post(routes.DeletePost, postHandler.Delete)

        // Social actions (POST only)
        r.Post(routes.FollowUser, socialHandler.Follow)
        r.Post(routes.UnfollowUser, socialHandler.Unfollow)

        // Logout
        r.Post(routes.Logout, authHandler.Logout)
    })

    // API routes (JSON responses only)
    r.Route("/api", func(api chi.Router) {
        api.Use(middleware.ContentTypeJSON)
        api.Get("/health", healthHandler.Check)
        api.Handle("/metrics", promhttp.Handler())
    })

    // SEO routes (plaintext/XML, not under /api)
    r.Get("/robots.txt", seoHandler.Robots)
    r.Get("/sitemap.xml", seoHandler.Sitemap)
    r.Get("/LLMs.txt", seoHandler.LLMsTxt) // Optional AI-friendly resource

    return &Server{
        Router:   r,
        Config:   cfg,
        Services: services,
    }
}
```

### API Design Policy

- **`/api/*` endpoints:** Return **JSON only** (no HTML). Keep this surface minimal and stable.
- **Top-level exceptions:** `/robots.txt`, `/sitemap.xml`, `/LLMs.txt` (plaintext/XML for SEO, not under /api)
- **Everything else:** Returns HTML for hypermedia-driven interactions

## Middleware Stack

### Core Middleware

```go
// internal/middleware/core.go

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
// internal/handlers/auth_handler.go

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
// internal/handlers/post_handler.go

// Content handlers work with slugs, not IDs
func (h *PostHandler) Show(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")

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
    slug := chi.URLParam(r, "slug")
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
// internal/handlers/post_handler_test.go
func TestPostHandler_Show(t *testing.T) {
    // Setup
    handler := NewPostHandler(mockService)

    // Create request
    req := httptest.NewRequest("GET", "/p/my-first-post", nil)
    req = req.WithContext(chi.NewRouteContext())
    chi.URLParam(req, "slug", "my-first-post")

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
