# Middleware Guide

Learn about the production-ready middleware stack in Tracks-generated applications.

## Overview

Middleware in Go web applications are functions that wrap HTTP handlers, executing code before and/or after the handler runs. Tracks generates applications with a carefully ordered middleware stack for security, observability, and performance.

## Middleware Stack

Tracks applications include the following middleware, applied in this order:

| Middleware | Purpose | Configurable |
|------------|---------|--------------|
| RequestID | Assigns unique ID to each request for tracing | No |
| RealIP | Extracts real client IP from proxy headers | No |
| Compress | Gzip response compression | No |
| Logger | Structured request logging | No |
| Recoverer | Catches panics, logs stack traces | No |
| Timeout | Cancels long-running requests | Yes |
| Throttle | Limits concurrent requests | Yes |
| CSP | Generates nonces for Content-Security-Policy | No |
| SecurityHeaders | Sets security headers (X-Frame-Options, etc.) | No |
| CORS | Handles cross-origin requests | Yes |

## Middleware Ordering

The order of middleware matters for both security and correctness:

1. **RequestID first** - Ensures all subsequent logging includes the request ID
2. **RealIP before logging** - Accurate client IP in logs
3. **Compress early** - Wraps response writer before content is written
4. **Logger/Recoverer** - Catch panics before they propagate
5. **Timeout before Throttle** - Timeout applies to total request time
6. **Security headers last** - Applied after request processing

Changing the order can break functionality or create security issues.

## Configuration Reference

Configure middleware via environment variables in `.env`:

### Timeout Middleware

Cancels requests that take too long, preventing slow clients from consuming server resources.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_MIDDLEWARE_TIMEOUT_REQUEST` | duration | `60s` | Maximum request duration |

**When to adjust:**

- Increase for endpoints with long-running operations (file uploads, report generation)
- Decrease in high-traffic environments to fail fast
- Set to `0` to disable (not recommended in production)

**Behavior:** When timeout is reached, the request context is cancelled. Well-behaved handlers should check `ctx.Done()` and stop processing.

### Throttle Middleware

Limits concurrent requests to prevent server overload. Uses Chi's `Throttle` middleware.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_MIDDLEWARE_THROTTLE_LIMIT` | int | `100` | Maximum concurrent requests |
| `APP_MIDDLEWARE_THROTTLE_BACKLOG_LIMIT` | int | `50` | Requests queued when at limit |
| `APP_MIDDLEWARE_THROTTLE_BACKLOG_TIMEOUT` | duration | `30s` | How long queued requests wait |

**When to adjust:**

- Increase limit for servers with more resources
- Decrease limit if requests are resource-intensive
- Set backlog to `0` to reject immediately when at limit (returns 503)

**Behavior:**

1. Requests under limit: processed immediately
2. Requests at limit with backlog space: queued
3. Requests at limit with full backlog: 503 Service Unavailable
4. Queued requests timing out: 503 Service Unavailable

### CORS Middleware

Handles Cross-Origin Resource Sharing for requests from different domains. **Disabled by default.**

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_MIDDLEWARE_CORS_ENABLED` | bool | `false` | Enable CORS handling |
| `APP_MIDDLEWARE_CORS_ALLOWED_ORIGINS` | string[] | `[]` | Allowed origins (comma-separated) |
| `APP_MIDDLEWARE_CORS_ALLOWED_METHODS` | string[] | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `APP_MIDDLEWARE_CORS_ALLOWED_HEADERS` | string[] | `Accept,Authorization,Content-Type,X-CSRF-Token` | Allowed request headers |
| `APP_MIDDLEWARE_CORS_EXPOSED_HEADERS` | string[] | `[]` | Headers exposed to JavaScript |
| `APP_MIDDLEWARE_CORS_ALLOW_CREDENTIALS` | bool | `false` | Allow cookies/auth headers |
| `APP_MIDDLEWARE_CORS_MAX_AGE` | int | `300` | Preflight cache duration (seconds) |

**When to enable:**

- SPA on different domain than API
- Mobile app calling your API
- Third-party integrations

**When NOT to enable:**

- Same-origin requests (SPA served from same domain)
- Server-rendered HTML with HTMX (same-origin by default)

**Security considerations:**

- Never use `*` with `AllowCredentials=true`
- Specify exact origins, not patterns
- Be restrictive with allowed headers

**Example configuration:**

```bash
APP_MIDDLEWARE_CORS_ENABLED=true
APP_MIDDLEWARE_CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
APP_MIDDLEWARE_CORS_ALLOW_CREDENTIALS=true
```

## Security Headers

The `SecurityHeaders` middleware sets the following headers:

### X-Content-Type-Options

```text
X-Content-Type-Options: nosniff
```

Prevents browsers from MIME-sniffing responses, reducing XSS risk.

### X-Frame-Options

```text
X-Frame-Options: DENY
```

Prevents clickjacking by blocking the page from being embedded in frames.

### Referrer-Policy

```text
Referrer-Policy: strict-origin-when-cross-origin
```

Controls how much referrer information is sent with requests:

- Same-origin: full URL
- Cross-origin: origin only (no path)
- HTTPS to HTTP: no referrer

### Permissions-Policy

```text
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

Disables browser features that could be exploited. Modify if your app needs these features.

### Content-Security-Policy

```text
Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-xxx'; ...
```

Restricts resource loading to prevent XSS. The CSP middleware generates a unique nonce for each request, allowing inline scripts that include the nonce.

**CSP directives:**

| Directive | Value | Purpose |
|-----------|-------|---------|
| `default-src` | `'self'` | Fallback for unspecified directives |
| `script-src` | `'self' 'nonce-xxx'` | Scripts from same origin + nonced inline |
| `style-src` | `'self' 'unsafe-inline'` | Styles from same origin + inline |
| `img-src` | `'self' data:` | Images from same origin + data URIs |
| `font-src` | `'self'` | Fonts from same origin |
| `connect-src` | `'self'` | XHR/fetch to same origin |
| `frame-ancestors` | `'none'` | Cannot be framed |
| `base-uri` | `'self'` | Restricts `<base>` tag |
| `form-action` | `'self'` | Form submissions to same origin |

## Using CSP Nonces

The CSP middleware generates a nonce and stores it in the request context. Access it in templ components:

```templ
package components

import "github.com/a-h/templ"

templ InlineScript() {
    <script nonce={ templ.GetNonce(ctx) }>
        console.log("This script is allowed by CSP");
    </script>
}
```

All HTMX scripts included by Tracks use the nonce automatically.

## Adding Custom Middleware

Add middleware in `internal/http/routes.go`:

```go
func (s *Server) routes() {
    // ... existing middleware ...

    // Add custom middleware before routes
    s.router.Use(myCustomMiddleware)

    // ... routes ...
}

func myCustomMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        start := time.Now()

        next.ServeHTTP(w, r)

        // After handler
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    })
}
```

### Route-Specific Middleware

Apply middleware to specific routes using `r.Group()`:

```go
s.router.Route("/admin", func(r chi.Router) {
    r.Use(adminAuthMiddleware)
    r.Get("/", adminHandler.Dashboard)
})
```

## Disabling Middleware

To disable configurable middleware, set their values to disable:

- **Timeout:** Set `APP_MIDDLEWARE_TIMEOUT_REQUEST=0`
- **Throttle:** Set `APP_MIDDLEWARE_THROTTLE_LIMIT=0`
- **CORS:** Set `APP_MIDDLEWARE_CORS_ENABLED=false` (default)

Non-configurable middleware (RequestID, Compress, Logger, etc.) must be removed from `routes.go` if not needed.

**Warning:** Disabling security middleware (Recoverer, SecurityHeaders) in production is not recommended.

## Testing Middleware

Verify headers with curl:

```bash
# Check security headers
curl -I http://localhost:8080/

# Expected output includes:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# Content-Security-Policy: default-src 'self'; ...
```

Test CORS preflight:

```bash
curl -X OPTIONS -H "Origin: https://app.example.com" \
     -H "Access-Control-Request-Method: POST" \
     -I http://localhost:8080/api/endpoint
```

## Related Topics

- [Architecture Overview](/docs/guides/architecture-overview) - Overall application structure
- [Layer Guide](/docs/guides/layer-guide) - HTTP layer details
- [Caching Guide](/docs/guides/caching) - Asset caching middleware
