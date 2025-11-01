# Security Architecture

**[← Back to Summary](./0_summary.md)**

## Overview

Security in Tracks is built-in from the ground up, not bolted on later. Every generated application includes comprehensive security measures by default, including secure headers, Content Security Policy (CSP), CSRF protection, rate limiting, and secure session management.

## Goals

- Secure by default with zero-config security headers
- Protection against common web vulnerabilities (XSS, CSRF, SQL injection)
- Content Security Policy with automatic nonce generation
- Session security with secure cookie defaults
- Rate limiting to prevent abuse and DDoS

## User Stories

- As a developer, I want security headers automatically applied so I don't forget them
- As a developer, I want CSRF protection without managing tokens manually
- As a site owner, I want protection against XSS attacks via CSP
- As a user, I want my session to be secure and not hijackable
- As an admin, I want rate limiting to prevent brute force attacks

## CSP Nonce Generation

⚠️ **CRITICAL: Middleware Order**
The nonce middleware MUST be registered before SecureHeaders middleware so that the CSP can include the nonce value. This is essential for inline script security.

```go
// internal/http/middleware/nonce.go
package middleware

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "net/http"

    "github.com/a-h/templ"
)

func WithNonce() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            nonce := generateNonce()
            ctx := templ.WithNonce(r.Context(), nonce)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func generateNonce() string {
    b := make([]byte, 16)
    rand.Read(b)
    return base64.StdEncoding.EncodeToString(b)
}
```

## Security Headers

```go
// internal/http/middleware/security.go
package middleware

import (
    "net/http"
    "strings"
    "github.com/a-h/templ"
)

func SecureHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            nonce := templ.GetNonce(r.Context())

            // Strict Transport Security (HSTS)
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

            // Prevent clickjacking
            w.Header().Set("X-Frame-Options", "DENY")

            // Prevent MIME type sniffing
            w.Header().Set("X-Content-Type-Options", "nosniff")

            // Control referrer information
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

            // Restrict browser features
            w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

            // Content Security Policy with nonce
            csp := strings.Join([]string{
                "default-src 'self';",
                "script-src 'self' 'strict-dynamic' 'nonce-" + nonce + "';",
                "style-src 'self' 'unsafe-inline';",  // Required for Alpine.js x-show/x-if
                "img-src 'self' data:;",
                "font-src 'self';",
                "connect-src 'self';",
                "frame-ancestors 'none';",
            }, " ")
            w.Header().Set("Content-Security-Policy", csp)

            next.ServeHTTP(w, r)
        })
    }
}
```

## Session Security

Sessions are configured with secure defaults that prevent hijacking and ensure privacy.

```go
// internal/config/session.go
package config

import (
    "net/http"
    "os"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/alexedwards/scs/v2/stores/cookiestore"
    "github.com/alexedwards/scs/v2/stores/redisstore"
    "github.com/gomodule/redigo/redis"
)

func NewSessionManager(cfg SessionConfig) *scs.SessionManager {
    sessionManager := scs.New()

    // Secure session configuration
    sessionManager.Lifetime = 24 * time.Hour
    sessionManager.IdleTimeout = 20 * time.Minute
    sessionManager.Cookie.Name = "tracks_session"
    sessionManager.Cookie.HttpOnly = true         // Prevent JavaScript access
    sessionManager.Cookie.Secure = true           // HTTPS only
    sessionManager.Cookie.SameSite = http.SameSiteLaxMode // OAuth-friendly, CSRF protection
    sessionManager.Cookie.Persist = true          // Survive browser restart

    // Use encrypted cookie store by default
    sessionKey := []byte(os.Getenv("SESSION_ENC_KEY")) // 32 or 64 bytes
    if len(sessionKey) < 32 {
        panic("SESSION_ENC_KEY must be at least 32 bytes")
    }
    sessionManager.Store = cookiestore.New(sessionKey)

    // Optional Redis for complex session data
    if cfg.Store == "redis" && cfg.RedisURL != "" {
        pool := &redis.Pool{
            Dial: func() (redis.Conn, error) {
                return redis.DialURL(cfg.RedisURL)
            },
        }
        sessionManager.Store = redisstore.New(pool)
    }

    return sessionManager
}
```

## Rate Limiting

Rate limiting is applied at multiple levels to prevent abuse and brute force attacks.

```go
// internal/pkg/ratelimit/limiter.go
package ratelimit

import (
    "fmt"
    "time"
    "github.com/go-chi/httprate"
)

// Rate limiter with configurable limits
type RateLimiter struct {
    store  RateLimitStore
    config RateLimitConfig
}

type RateLimitConfig struct {
    OTPPerHour        int `yaml:"otp_per_hour" default:"5"`
    LoginPerMinute    int `yaml:"login_per_minute" default:"5"`
    RegisterPerHour   int `yaml:"register_per_hour" default:"10"`
    APIRequestsPerMin int `yaml:"api_per_minute" default:"100"`
}

// Middleware for general API rate limiting
func APIRateLimit() func(http.Handler) http.Handler {
    // 100 requests per minute per IP
    return httprate.LimitByIP(100, time.Minute)
}

// Middleware for auth endpoints
func AuthRateLimit() func(http.Handler) http.Handler {
    // 5 attempts per minute per IP
    return httprate.Limit(
        5,
        time.Minute,
        httprate.WithKeyFunc(func(r *http.Request) (string, error) {
            // Rate limit by IP + email combo for auth
            ip := httprate.KeyByIP(r)
            email := r.FormValue("email")
            return fmt.Sprintf("%s:%s", ip, email), nil
        }),
    )
}

// Service-level rate limiting for OTP
func (s *RateLimiter) AllowOTP(email string) bool {
    key := fmt.Sprintf("otp:%s", email)
    count := s.store.Increment(key, time.Hour)
    return count <= s.config.OTPPerHour
}

// Service-level rate limiting for login
func (s *RateLimiter) AllowLogin(email string) bool {
    key := fmt.Sprintf("login:%s", email)
    count := s.store.Increment(key, time.Minute)
    return count <= s.config.LoginPerMinute
}
```

## CSRF Protection

CSRF protection is built into the session management system and form handling.

```go
// internal/http/middleware/csrf.go
package middleware

import (
    "net/http"
    "github.com/gorilla/csrf"
)

func CSRFProtection(authKey []byte) func(http.Handler) http.Handler {
    return csrf.Protect(
        authKey,
        csrf.Secure(true),                    // Require HTTPS
        csrf.HttpOnly(true),                  // Prevent JS access
        csrf.SameSite(csrf.SameSiteLaxMode), // OAuth-friendly
        csrf.Path("/"),
        csrf.FieldName("csrf_token"),
        csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)),
    )
}

func csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusForbidden)
    w.Write([]byte("CSRF token validation failed"))
}
```

Template integration:

```go
// In templates, automatically include CSRF token
templ LoginForm(ctx context.Context) {
    <form method="post" action="/login">
        @CSRFToken(ctx)
        <input type="email" name="email" required/>
        <button type="submit">Login</button>
    </form>
}

// Helper component
templ CSRFToken(ctx context.Context) {
    <input type="hidden" name="csrf_token" value={ csrf.Token(ctx) }/>
}
```

## File Upload Security

File uploads are validated using magic numbers to prevent malicious file types.

```go
// internal/http/handlers/upload_handler.go
package handlers

import (
    "io"
    "net/http"
    "github.com/h2non/filetype"
)

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
    // Limit upload size
    r.ParseMultipartForm(10 << 20) // 10MB

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error getting file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // ✅ FIXED: Validate file type using magic numbers
    // Read enough bytes for complete magic number detection
    head := make([]byte, 512)  // Changed from 261 to 512 for better detection
    file.Read(head)

    // Check file type
    kind, _ := filetype.Match(head)
    if !filetype.IsImage(head) {
        http.Error(w, "Only images allowed", http.StatusBadRequest)
        return
    }

    // Additional validation
    if kind.MIME.Value != "" && !isAllowedMIME(kind.MIME.Value) {
        http.Error(w, "File type not allowed", http.StatusBadRequest)
        return
    }

    // Check for malicious content
    if containsMaliciousPatterns(head) {
        http.Error(w, "File contains suspicious content", http.StatusBadRequest)
        return
    }

    // Reset file pointer
    file.Seek(0, io.SeekStart)

    // Process upload...
}

func isAllowedMIME(mime string) bool {
    allowed := []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
    }
    for _, a := range allowed {
        if a == mime {
            return true
        }
    }
    return false
}

func containsMaliciousPatterns(data []byte) bool {
    // Check for common attack vectors
    patterns := [][]byte{
        []byte("<?php"),
        []byte("<script"),
        []byte("javascript:"),
        []byte("<iframe"),
    }
    for _, pattern := range patterns {
        if bytes.Contains(data, pattern) {
            return true
        }
    }
    return false
}
```

## SQL Injection Prevention

SQL injection is prevented through the use of SQLC, which generates type-safe Go code from SQL queries. All queries are parameterized at compile time.

```sql
-- queries/users.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, email, username) VALUES (?, ?, ?) RETURNING *;
```

Generated Go code is always safe:

```go
// This is generated code - always uses parameterized queries
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
    row := q.db.QueryRowContext(ctx, getUserByEmail, email)
    // ... safe scanning code ...
}
```

## XSS Prevention

XSS is prevented through:

1. Template auto-escaping (templ)
2. Content Security Policy
3. Input sanitization

```go
// internal/pkg/sanitizer/sanitizer.go
package sanitizer

import (
    "github.com/microcosm-cc/bluemonday"
)

type Sanitizer struct {
    policy *bluemonday.Policy
}

func NewSanitizer() *Sanitizer {
    // Strict policy - no HTML allowed
    strict := bluemonday.StrictPolicy()

    // UGC policy - allows basic formatting
    ugc := bluemonday.UGCPolicy()

    return &Sanitizer{
        policy: ugc,
    }
}

func (s *Sanitizer) SanitizeHTML(input string) string {
    return s.policy.Sanitize(input)
}

func (s *Sanitizer) SanitizePlainText(input string) string {
    return bluemonday.StrictPolicy().Sanitize(input)
}
```

## Security Checklist

### Development

- [ ] Set strong `SESSION_ENC_KEY` (minimum 32 bytes)
- [ ] Configure CSRF auth key
- [ ] Enable HTTPS in development with mkcert
- [ ] Test CSP violations in browser console
- [ ] Verify rate limiting works

### Production

- [ ] Force HTTPS with redirect
- [ ] Set secure cookie flags
- [ ] Enable HSTS preloading
- [ ] Configure rate limits appropriately
- [ ] Set up fail2ban for repeated violations
- [ ] Enable audit logging
- [ ] Configure WAF rules
- [ ] Set up security monitoring

## Testing Security

```go
// internal/http/handlers/security_test.go
func TestSecurityHeaders(t *testing.T) {
    handler := NewServer()

    req := httptest.NewRequest("GET", "/", nil)
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    // Check security headers
    assert.NotEmpty(t, rr.Header().Get("X-Frame-Options"))
    assert.NotEmpty(t, rr.Header().Get("Content-Security-Policy"))
    assert.NotEmpty(t, rr.Header().Get("X-Content-Type-Options"))
    assert.Contains(t, rr.Header().Get("Content-Security-Policy"), "nonce-")
}

func TestRateLimiting(t *testing.T) {
    limiter := NewRateLimiter(config)

    // Should allow first attempts
    for i := 0; i < 5; i++ {
        assert.True(t, limiter.AllowLogin("test@example.com"))
    }

    // Should block after limit
    assert.False(t, limiter.AllowLogin("test@example.com"))
}
```

## Best Practices

1. **Never disable security features for convenience** - Find the right way, not the easy way
2. **Test security measures regularly** - Automated security tests in CI
3. **Keep dependencies updated** - Security patches are critical
4. **Monitor for CSP violations** - They indicate potential XSS attempts
5. **Log security events** - Failed logins, rate limit hits, CSRF failures
6. **Use secure defaults** - Make the secure path the easy path
7. **Defense in depth** - Multiple layers of security

## Next Steps

- Continue to [Templates & Assets →](./7_templates_assets.md)
- Back to [← Web Layer](./5_web_layer.md)
- Return to [Summary](./0_summary.md)
