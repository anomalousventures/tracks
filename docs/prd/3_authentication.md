# Authentication System

**[← Back to Summary](./0_summary.md) | [← Database Layer](./2_database_layer.md) | [Authorization & RBAC →](./4_authorization_rbac.md)**

## Overview

Tracks provides a secure, flexible authentication system that prioritizes passwordless authentication while supporting OAuth2 and optional password-based auth. Sessions are managed securely with encrypted cookies or Redis backing.

## Goals

- Support multiple authentication methods (OTP, OAuth2, passwords)
- Secure session management with database or Redis backing
- Dynamic OAuth provider registration based on configuration
- Rate limiting on all authentication endpoints
- Passwordless by default for better security
- Magic link support as an alternative to OTP

## User Stories

- As a user, I want to log in with just my email (passwordless)
- As a user, I want to authenticate with my Google/GitHub account
- As a developer, I want OAuth providers configured via environment variables
- As a security engineer, I want rate limiting on login attempts
- As a user, I want secure sessions that persist across restarts
- As a user, I want the option to use a password if required by policy

## Authentication Methods

### 1. OTP (One-Time Password) - Default

The primary authentication method uses email or SMS-based OTP codes.

```go
// internal/services/auth.go
package services

import (
    "context"
    "crypto/rand"
    "encoding/binary"
    "errors"
    "fmt"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/markbates/goth"
    "myapp/internal/pkg/identifier"
)

type AuthService struct {
    sessions      *scs.SessionManager
    emailAdapter  EmailAdapter
    smsAdapter    SMSAdapter
    otpStore      OTPStore
    userRepo      UserRepository
    permService   *PermissionService
    gothProviders []goth.Provider
    rateLimiter   *RateLimiter
}

// OTP generation with crypto/rand (6 digits)
func (s *AuthService) GenerateOTP() (string, error) {
    b := make([]byte, 4)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("generating random bytes: %w", err)
    }
    // Generate number between 100000-999999
    num := binary.BigEndian.Uint32(b) % 900000 + 100000
    return fmt.Sprintf("%06d", num), nil
}

// Send OTP with rate limiting
func (s *AuthService) SendOTP(ctx context.Context, email string) error {
    // Rate limit OTP generation (5 per hour per email)
    if !s.rateLimiter.AllowOTP(email) {
        return ErrRateLimited
    }

    code, err := s.GenerateOTP()
    if err != nil {
        return err
    }

    // Store with 10-minute expiry
    if err := s.otpStore.Set(ctx, email, code, 10*time.Minute); err != nil {
        return fmt.Errorf("storing OTP: %w", err)
    }

    // Send via email
    return s.emailAdapter.SendOTP(ctx, email, code)
}

// OTP verification with attempt limiting
func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (*User, error) {
    key := fmt.Sprintf("otp:%s", email)

    // Check attempts (max 3)
    attempts, _ := s.otpStore.GetAttempts(ctx, key)
    if attempts >= 3 {
        s.otpStore.Delete(ctx, key)
        return nil, ErrTooManyAttempts
    }

    storedCode, err := s.otpStore.Get(ctx, key)
    if err != nil {
        return nil, ErrOTPExpired
    }

    if storedCode != code {
        s.otpStore.IncrementAttempts(ctx, key)
        return nil, ErrInvalidOTP
    }

    // Success - delete OTP
    s.otpStore.Delete(ctx, key)

    // Get or create user
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            // Auto-create user on first login
            user = &User{
                ID:    identifier.NewID(),  // UUIDv7
                Email: email,
                Username: generateUsername(email),
            }
            if err := s.userRepo.Create(ctx, user); err != nil {
                return nil, fmt.Errorf("creating user: %w", err)
            }

            // Assign default role
            if err := s.permService.GrantRole(ctx, user.ID, "user", "system", "new user"); err != nil {
                // Log but don't fail login
                log.Error("failed to grant default role", "error", err)
            }
        } else {
            return nil, err
        }
    }

    return user, nil
}
```

### 2. Magic Links

Alternative to OTP using secure, time-limited links.

```go
// Magic link generation
func (s *AuthService) GenerateMagicLink(ctx context.Context, email string) error {
    if !s.rateLimiter.AllowMagicLink(email) {
        return ErrRateLimited
    }

    token := generateSecureToken(32) // 32 bytes of randomness

    // Store with 15-minute expiry
    if err := s.otpStore.Set(ctx, fmt.Sprintf("magic:%s", token), email, 15*time.Minute); err != nil {
        return err
    }

    link := fmt.Sprintf("%s/auth/magic?token=%s", s.config.BaseURL, token)
    return s.emailAdapter.SendMagicLink(ctx, email, link)
}

func (s *AuthService) VerifyMagicLink(ctx context.Context, token string) (*User, error) {
    key := fmt.Sprintf("magic:%s", token)

    email, err := s.otpStore.Get(ctx, key)
    if err != nil {
        return nil, ErrInvalidToken
    }

    // Delete token (one-time use)
    s.otpStore.Delete(ctx, key)

    // Get or create user (same as OTP flow)
    return s.getOrCreateUser(ctx, email)
}

func generateSecureToken(length int) string {
    b := make([]byte, length)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}
```

### 3. OAuth2 Providers

Dynamic OAuth provider registration based on configuration.

```go
// internal/config/oauth.go
package config

import (
    "github.com/markbates/goth"
    "github.com/markbates/goth/providers/github"
    "github.com/markbates/goth/providers/google"
    "github.com/markbates/goth/providers/microsoft"
)

type OAuthConfig struct {
    GitHub    ProviderConfig
    Google    ProviderConfig
    Microsoft ProviderConfig
    // Add more providers as needed
}

type ProviderConfig struct {
    ClientID     string `mapstructure:"client_id"`
    ClientSecret string `mapstructure:"client_secret"`
    CallbackURL  string `mapstructure:"callback_url"`
    Enabled      bool   `mapstructure:"enabled"`
}

func SetupOAuthProviders(cfg OAuthConfig) []goth.Provider {
    var providers []goth.Provider

    // Only register providers with valid credentials
    if cfg.GitHub.Enabled && cfg.GitHub.ClientID != "" {
        providers = append(providers,
            github.New(
                cfg.GitHub.ClientID,
                cfg.GitHub.ClientSecret,
                cfg.GitHub.CallbackURL,
                "user:email",
            ),
        )
    }

    if cfg.Google.Enabled && cfg.Google.ClientID != "" {
        providers = append(providers,
            google.New(
                cfg.Google.ClientID,
                cfg.Google.ClientSecret,
                cfg.Google.CallbackURL,
                "email", "profile",
            ),
        )
    }

    if cfg.Microsoft.Enabled && cfg.Microsoft.ClientID != "" {
        providers = append(providers,
            microsoft.New(
                cfg.Microsoft.ClientID,
                cfg.Microsoft.ClientSecret,
                cfg.Microsoft.CallbackURL,
                "User.Read",
            ),
        )
    }

    if len(providers) > 0 {
        goth.UseProviders(providers...)
    }

    return providers
}

// Dynamic route registration
func RegisterOAuthRoutes(r chi.Router, providers []goth.Provider) {
    for _, p := range providers {
        provider := p.Name()
        r.Get(fmt.Sprintf("/auth/%s", provider), BeginOAuthHandler)
        r.Get(fmt.Sprintf("/auth/%s/callback", provider), OAuthCallbackHandler)
    }
}
```

### 4. Password Authentication (Optional)

While passwordless is preferred, passwords can be enabled for compliance.

```go
// internal/services/auth_password.go
package services

import (
    "golang.org/x/crypto/bcrypt"
)

// Only used when passwords are enabled
func (s *AuthService) AuthenticateWithPassword(ctx context.Context, email, password string) (*User, error) {
    if !s.config.PasswordsEnabled {
        return nil, ErrPasswordsDisabled
    }

    if !s.rateLimiter.AllowPasswordAttempt(email) {
        return nil, ErrRateLimited
    }

    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        // Perform dummy hash to prevent timing attacks
        bcrypt.CompareHashAndPassword([]byte("$2a$10$dummy"), []byte(password))
        return nil, ErrInvalidCredentials
    }

    if user.PasswordHash == "" {
        return nil, ErrPasswordNotSet
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        s.rateLimiter.RecordFailedAttempt(email)
        return nil, ErrInvalidCredentials
    }

    return user, nil
}

func (s *AuthService) SetPassword(ctx context.Context, userID, password string) error {
    // Validate password strength
    if err := validatePasswordStrength(password); err != nil {
        return err
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("hashing password: %w", err)
    }

    return s.userRepo.UpdatePasswordHash(ctx, userID, string(hash))
}
```

## Rate Limiting

Comprehensive rate limiting for all authentication endpoints.

```go
// internal/services/rate_limiter.go
package services

import (
    "context"
    "fmt"
    "time"
)

type RateLimiter struct {
    store  RateLimitStore
    config RateLimitConfig
}

type RateLimitConfig struct {
    OTPPerHour         int           `yaml:"otp_per_hour" default:"5"`
    LoginPerMinute     int           `yaml:"login_per_minute" default:"5"`
    RegisterPerHour    int           `yaml:"register_per_hour" default:"10"`
    MagicLinkPerHour   int           `yaml:"magic_link_per_hour" default:"3"`
    PasswordPerMinute  int           `yaml:"password_per_minute" default:"3"`
    LockoutDuration    time.Duration `yaml:"lockout_duration" default:"15m"`
}

type RateLimitStore interface {
    Increment(ctx context.Context, key string, window time.Duration) (int, error)
    Get(ctx context.Context, key string) (int, error)
    SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error
}

func (r *RateLimiter) AllowOTP(identifier string) bool {
    key := fmt.Sprintf("otp_rate:%s", identifier)
    count, _ := r.store.Increment(context.Background(), key, time.Hour)
    return count <= r.config.OTPPerHour
}

func (r *RateLimiter) AllowMagicLink(identifier string) bool {
    key := fmt.Sprintf("magic_rate:%s", identifier)
    count, _ := r.store.Increment(context.Background(), key, time.Hour)
    return count <= r.config.MagicLinkPerHour
}

func (r *RateLimiter) AllowPasswordAttempt(identifier string) bool {
    // Check if account is locked
    lockKey := fmt.Sprintf("locked:%s", identifier)
    if locked, _ := r.store.Get(context.Background(), lockKey); locked > 0 {
        return false
    }

    key := fmt.Sprintf("password_rate:%s", identifier)
    count, _ := r.store.Increment(context.Background(), key, time.Minute)
    return count <= r.config.PasswordPerMinute
}

func (r *RateLimiter) RecordFailedAttempt(identifier string) {
    key := fmt.Sprintf("failed:%s", identifier)
    count, _ := r.store.Increment(context.Background(), key, 15*time.Minute)

    // Lock account after 5 failed attempts
    if count >= 5 {
        lockKey := fmt.Sprintf("locked:%s", identifier)
        r.store.SetWithExpiry(context.Background(), lockKey, 1, r.config.LockoutDuration)
    }
}
```

## Session Management

Secure session management with multiple storage backends.

```go
// internal/middleware/session.go
package middleware

import (
    "context"
    "net/http"
    "os"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/alexedwards/scs/v2/memstore"
    "github.com/alexedwards/scs/redisstore"
    "github.com/gomodule/redigo/redis"
    "myapp/internal/pkg/identifier"
)

type SessionConfig struct {
    Store        string        `mapstructure:"store"`         // cookie, memory, redis
    Lifetime     time.Duration `mapstructure:"lifetime"`      // Default: 24h
    IdleTimeout  time.Duration `mapstructure:"idle_timeout"`  // Default: 20m
    CookieName   string        `mapstructure:"cookie_name"`   // Default: tracks_session
    CookieDomain string        `mapstructure:"cookie_domain"`
    RedisURL     string        `mapstructure:"redis_url"`
    Secure       bool          `mapstructure:"secure"`        // HTTPS only
}

func SetupSessions(cfg SessionConfig) *scs.SessionManager {
    sessionManager := scs.New()

    // Configure session settings
    sessionManager.Lifetime = cfg.Lifetime
    if sessionManager.Lifetime == 0 {
        sessionManager.Lifetime = 24 * time.Hour
    }

    sessionManager.IdleTimeout = cfg.IdleTimeout
    if sessionManager.IdleTimeout == 0 {
        sessionManager.IdleTimeout = 20 * time.Minute
    }

    sessionManager.Cookie.Name = cfg.CookieName
    if sessionManager.Cookie.Name == "" {
        sessionManager.Cookie.Name = "tracks_session"
    }

    sessionManager.Cookie.Domain = cfg.CookieDomain
    sessionManager.Cookie.HttpOnly = true
    sessionManager.Cookie.Secure = cfg.Secure
    sessionManager.Cookie.SameSite = http.SameSiteLaxMode // OAuth-friendly
    sessionManager.Cookie.Persist = true

    // Setup store based on configuration
    switch cfg.Store {
    case "redis":
        pool := &redis.Pool{
            MaxIdle:     10,
            MaxActive:   100,
            Wait:        true,
            IdleTimeout: 240 * time.Second,
            Dial: func() (redis.Conn, error) {
                return redis.DialURL(cfg.RedisURL)
            },
            TestOnBorrow: func(c redis.Conn, t time.Time) error {
                _, err := c.Do("PING")
                return err
            },
        }
        sessionManager.Store = redisstore.New(pool)

    case "memory":
        sessionManager.Store = memstore.New()

    default: // cookie (encrypted)
        sessionKey := []byte(os.Getenv("SESSION_KEY"))
        if len(sessionKey) < 32 {
            // Generate a random key if not provided (development only)
            sessionKey = make([]byte, 32)
            if _, err := rand.Read(sessionKey); err != nil {
                panic("failed to generate session key")
            }
        }
        sessionManager.Store = scs.NewCookieStore(sessionKey)
    }

    return sessionManager
}

// RequireAuth middleware
func RequireAuth(sessions *scs.SessionManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := sessions.GetString(r.Context(), "user_id")
            if userID == "" {
                sessions.Put(r.Context(), "flash_error", "Please log in to continue")
                sessions.Put(r.Context(), "redirect_after_login", r.URL.Path)
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            // Validate UUID format (UUIDv7)
            if err := identifier.ValidateID(userID); err != nil {
                sessions.Remove(r.Context(), "user_id")
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            // Refresh session activity
            sessions.RenewToken(r.Context())

            // Add user ID to context
            ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// OptionalAuth middleware - adds user to context if authenticated
func OptionalAuth(sessions *scs.SessionManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := sessions.GetString(r.Context(), "user_id")

            ctx := r.Context()
            if userID != "" && identifier.ValidateID(userID) == nil {
                ctx = context.WithValue(ctx, ContextKeyUserID, userID)
            }

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Authentication Handlers

HTTP handlers for authentication flows.

```go
// internal/handlers/auth_handler.go
package handlers

type AuthHandler struct {
    authService *services.AuthService
    sessions    *scs.SessionManager
}

// OTP login flow
func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
    // Check if already logged in
    if h.sessions.GetString(r.Context(), "user_id") != "" {
        http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
        return
    }

    data := LoginPageData{
        CSRFToken: csrf.Token(r),
        Providers: h.authService.GetEnabledProviders(),
    }

    views.LoginPage(data).Render(r.Context(), w)
}

func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")

    if err := validateEmail(email); err != nil {
        h.renderError(w, r, "Invalid email address")
        return
    }

    if err := h.authService.SendOTP(r.Context(), email); err != nil {
        if errors.Is(err, services.ErrRateLimited) {
            h.renderError(w, r, "Too many attempts. Please try again later.")
            return
        }
        h.renderError(w, r, "Failed to send code. Please try again.")
        return
    }

    // Store email in session for verification step
    h.sessions.Put(r.Context(), "otp_email", email)

    // Show OTP verification form
    views.OTPForm(email).Render(r.Context(), w)
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
    email := h.sessions.GetString(r.Context(), "otp_email")
    code := r.FormValue("code")

    if email == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    user, err := h.authService.VerifyOTP(r.Context(), email, code)
    if err != nil {
        h.renderError(w, r, "Invalid or expired code")
        return
    }

    // Create session
    h.sessions.Put(r.Context(), "user_id", user.ID)
    h.sessions.Put(r.Context(), "user_email", user.Email)
    h.sessions.Remove(r.Context(), "otp_email")

    // Check for redirect
    redirectURL := h.sessions.PopString(r.Context(), "redirect_after_login")
    if redirectURL == "" {
        redirectURL = "/dashboard"
    }

    http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    h.sessions.Destroy(r.Context())
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

## Security Considerations

### Session Security

1. **Secure cookies**: Always use HttpOnly, Secure, and SameSite flags
2. **Session rotation**: Regenerate session ID on privilege changes
3. **Idle timeout**: Automatically expire inactive sessions
4. **Encryption**: All session data encrypted at rest

### Authentication Security

1. **Rate limiting**: Prevent brute force attacks
2. **Account lockout**: Temporary lockout after failed attempts
3. **Timing attack prevention**: Consistent response times
4. **Token security**: Cryptographically secure random tokens

### CSRF Protection

```go
// Use gorilla/csrf or similar
func SetupCSRF(sessionManager *scs.SessionManager) func(http.Handler) http.Handler {
    return csrf.Protect(
        []byte(os.Getenv("CSRF_KEY")),
        csrf.Secure(true),
        csrf.HttpOnly(true),
        csrf.SameSite(csrf.SameSiteLaxMode),
        csrf.Path("/"),
    )
}
```

## Error Definitions

```go
// internal/services/auth_errors.go
package services

import "errors"

var (
    ErrRateLimited       = errors.New("rate limit exceeded")
    ErrOTPExpired        = errors.New("OTP has expired")
    ErrInvalidOTP        = errors.New("invalid OTP")
    ErrTooManyAttempts   = errors.New("too many failed attempts")
    ErrInvalidToken      = errors.New("invalid or expired token")
    ErrPasswordsDisabled = errors.New("password authentication is disabled")
    ErrPasswordNotSet    = errors.New("password not set for this account")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrAccountLocked     = errors.New("account temporarily locked")
)
```

## Configuration

```yaml
# Example authentication configuration
auth:
  methods:
    otp:
      enabled: true
      length: 6
      expiry: 10m
    magic_link:
      enabled: true
      expiry: 15m
    password:
      enabled: false  # Disabled by default
      min_length: 8
      require_uppercase: true
      require_number: true
    oauth:
      github:
        enabled: true
        client_id: ${GITHUB_CLIENT_ID}
        client_secret: ${GITHUB_CLIENT_SECRET}
        callback_url: ${BASE_URL}/auth/github/callback
      google:
        enabled: true
        client_id: ${GOOGLE_CLIENT_ID}
        client_secret: ${GOOGLE_CLIENT_SECRET}
        callback_url: ${BASE_URL}/auth/google/callback

  session:
    store: cookie  # cookie, memory, redis
    lifetime: 24h
    idle_timeout: 20m
    cookie_name: tracks_session
    secure: true  # HTTPS only

  rate_limits:
    otp_per_hour: 5
    login_per_minute: 5
    register_per_hour: 10
    magic_link_per_hour: 3
    password_per_minute: 3
    lockout_duration: 15m
```

## Next Steps

- Continue to [Authorization & RBAC →](./4_authorization_rbac.md)
- Back to [Database Layer](./2_database_layer.md)
- Back to [Summary](./0_summary.md)
