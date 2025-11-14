# Templates & Assets

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks uses **templ** for type-safe HTML generation and **hashfs** for content-addressed asset serving. This combination provides compile-time template validation, automatic cache busting, and optimal performance with zero runtime overhead.

## Template System

### Goals

- Type-safe HTML generation with compile-time validation
- Component reusability and composability
- Seamless integration with Go's type system
- Zero runtime overhead
- Asset fingerprinting for cache busting
- Production-ready UI components out of the box
- Component ownership and customization freedom

### User Stories

- As a developer, I want compile-time errors for template mistakes so I catch bugs before runtime
- As a developer, I want IDE autocomplete for template functions and variables
- As a developer, I want to compose UI from reusable components
- As a frontend developer, I want to use familiar HTML syntax with Go expressions
- As a developer, I want professional UI components without building from scratch
- As a developer, I want to customize UI components to match my design

## UI Component Library (Templ-UI)

**Decision:** [ADR-009: Templ-UI for UI Components](../adr/009-templui-for-ui-components.md)

Every Tracks project includes **templ-ui** as a core dependency, providing 40+ production-ready UI components styled with TailwindCSS. Components are copied into your project during generation, giving you full ownership and customization freedom.

### Installation & Setup

During `tracks new` project generation:

```bash
# Tracks automatically runs:
templui init --dir internal/http/views/components
templui add "*"  # Install full starter set
```

Configuration is stored in `templui.yaml`:

```yaml
# templui.yaml
version: "1.0"
module: "github.com/myuser/myapp"
components_dir: "internal/http/views/components"
ui_dir: "internal/http/views/components/ui"
```

### Component Categories

**Forms** (input, textarea, button, select, checkbox, radio, label)

```go
// Using templ-ui form components
@ui.Button(ui.ButtonProps{
    Variant: "primary",
    Size: "md",
    Type: "submit",
}) {
    Log in
}

@ui.Input(ui.InputProps{
    Type: "email",
    Placeholder: "you@example.com",
    Required: true,
})
```

**Layout** (card, modal, dialog, sidebar, tabs, accordion, sheet)

```go
// Card component
@ui.Card(ui.CardProps{Class: "max-w-md"}) {
    @ui.CardHeader() {
        <h2>Welcome</h2>
    }
    @ui.CardContent() {
        <p>This is a templ-ui card component.</p>
    }
    @ui.CardFooter() {
        @ui.Button(ui.ButtonProps{Variant: "outline"}) {
            Learn more
        }
    }
}
```

**Feedback** (alert, toast, progress, spinner, skeleton)

```go
// Alert component
@ui.Alert(ui.AlertProps{Variant: "success"}) {
    @ui.AlertTitle() { Success! }
    @ui.AlertDescription() {
        Your changes have been saved.
    }
}
```

### Adding More Components

Users can add additional components as needed:

```bash
# Via tracks CLI (recommended)
tracks ui add calendar
tracks ui add data-table

# List available components
tracks ui list
```

### Customization Workflow

Components are copied into `internal/http/views/components/`, giving users full control:

```go
// internal/http/views/components/ui/button.templ
// You own this code - customize freely!
package ui

type ButtonProps struct {
    Variant  string // "primary" | "secondary" | "outline"
    Size     string // "sm" | "md" | "lg"
    Type     string // "button" | "submit"
    Disabled bool
    Class    string // Additional Tailwind classes
}

templ Button(props ButtonProps) {
    <button
        type={ props.Type }
        disabled?={ props.Disabled }
        class={
            templ.CSSClasses({
                "btn": true,
                "btn-primary": props.Variant == "primary",
                "btn-secondary": props.Variant == "secondary",
                "btn-outline": props.Variant == "outline",
                "btn-sm": props.Size == "sm",
                "btn-md": props.Size == "md" || props.Size == "",
                "btn-lg": props.Size == "lg",
                props.Class: props.Class != "",
            })
        }
    >
        { children... }
    </button>
}
```

Users can modify styling, behavior, or structure without breaking upstream compatibility since they own the code.

### Dark Mode Support

All templ-ui components include dark mode support via Tailwind's `dark:` classes:

```css
/* Automatically included in components */
.btn-primary {
    @apply bg-blue-600 text-white hover:bg-blue-700;
    @apply dark:bg-blue-500 dark:hover:bg-blue-600;
}
```

Theme switching handled by Alpine.js component (see CSS & JavaScript section).

### Asset Helper Components

```go
// internal/http/views/components/assets.templ
package components

import (
    "myapp/internal/assets"
    "github.com/a-h/templ"
)

// CSS loads a stylesheet with cache-busted hash name
templ CSS(path string) {
    <link
        rel="stylesheet"
        href={ "/assets/" + assets.HashName(path) }
        hx-preserve="true"
    />
}

// JS loads JavaScript with CSP nonce and hash name
templ JS(path string, nonce string) {
    <script
        src={ "/assets/" + assets.HashName(path) }
        nonce={ nonce }
        defer
        hx-preserve="true"
    ></script>
}

// Favicon with multiple sizes
templ Favicon() {
    <link rel="icon" type="image/x-icon" href={ "/assets/" + assets.HashName("favicon.ico") }/>
    <link rel="icon" type="image/png" sizes="16x16" href={ "/assets/" + assets.HashName("favicon-16x16.png") }/>
    <link rel="icon" type="image/png" sizes="32x32" href={ "/assets/" + assets.HashName("favicon-32x32.png") }/>
    <link rel="apple-touch-icon" sizes="180x180" href={ "/assets/" + assets.HashName("apple-touch-icon.png") }/>
}

// Meta tags for SEO
templ Meta(title, description string) {
    <title>{ title } - Tracks</title>
    <meta name="description" content={ description }/>
    <meta property="og:title" content={ title }/>
    <meta property="og:description" content={ description }/>
}
```

### Layout Pattern

```go
// internal/http/views/layouts/base.templ
package layouts

import (
    "myapp/internal/http/views/components"
    "github.com/a-h/templ"
)

templ Base(title, description string, nonce string) {
    <!DOCTYPE html>
    <html lang="en" data-theme="auto">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            @components.Meta(title, description)
            @components.Favicon()
            @components.CSS("app.css")
            @components.HTMXConfig(nonce)
            @components.JS("htmx.min.js", nonce)
            @components.JS("alpine.min.js", nonce)
        </head>
        <body hx-boost="true">
            <div id="toast-container" aria-live="polite"></div>
            { children... }
        </body>
    </html>
}

// HTMX configuration component
templ HTMXConfig(nonce string) {
    <meta name="htmx-config" content={`
        {
            "selfRequestsOnly": true,
            "inlineScriptNonce": "` + nonce + `",
            "useTemplateFragments": true,
            "scrollBehavior": "smooth"
        }
    `}/>
}
```

### Page Templates

```go
// internal/http/views/pages/home.templ
package pages

import (
    "myapp/internal/http/views/layouts"
    "myapp/internal/http/views/components"
)

templ HomePage(ctx context.Context) {
    @layouts.Base("Home", "Welcome to Tracks", templ.GetNonce(ctx)) {
        <header>
            @components.Nav(ctx)
        </header>

        <main class="container mx-auto px-4 py-8">
            <h1 class="text-4xl font-bold mb-4">Welcome to Tracks</h1>

            if user := GetUser(ctx); user != nil {
                <p>Hello, { user.Name }!</p>
                @components.UserDashboard(user)
            } else {
                <p>Please <a href="/login" class="link">log in</a> to continue.</p>
            }
        </main>

        <footer>
            @components.Footer()
        </footer>
    }
}
```

## Asset Management

### Goals

- Modern image formats (WebP, AVIF) for smaller file sizes
- Responsive images with proper srcset/sizes
- Content-addressed assets for cache busting
- Zero-runtime overhead with embedded assets
- Automatic CSS/JS bundling and minification

### User Stories

- As a user, I want fast-loading images in modern formats
- As a developer, I want automatic image optimization
- As a developer, I want cache-busted assets without manual versioning
- As a user, I want responsive images that look good on all devices
- As a developer, I want simple asset management

### hashfs Integration

```go
// internal/assets/embed.go
package assets

import (
    "embed"
    "net/http"
    "github.com/benbjohnson/hashfs"
)

//go:embed dist/**
var dist embed.FS

var FS *hashfs.FS

func Init() error {
    fs, err := hashfs.NewFS(dist)
    if err != nil {
        return err
    }
    FS = fs
    return nil
}

func Handler() http.Handler {
    return http.FileServerFS(FS)
}

func HashName(path string) string {
    return FS.HashName(path)
}
```

### Image Optimization Pipeline

```bash
# Single image with defaults
tracks image:prep web/images/hero.jpg
# Generates: hero-320w.webp, hero-640w.webp, hero-1024w.webp, hero-1920w.webp
#            hero-320w.avif, hero-640w.avif, ...
#            hero-placeholder.webp (blur, 20px wide)
#            hero_image.templ component

# Custom sizes and formats
tracks image:prep web/images/product.png \
  --sizes=400,800,1200 \
  --formats=webp,jpg \
  --quality=85 \
  --placeholder \
  --component
```

### Generated Image Component

```go
// internal/http/views/components/hero_image.templ (generated)
package components

templ HeroImage(alt string) {
    <picture>
        <source
            type="image/avif"
            srcset="/assets/hero-320w.avif 320w,
                    /assets/hero-640w.avif 640w,
                    /assets/hero-1024w.avif 1024w,
                    /assets/hero-1920w.avif 1920w"
            sizes="(max-width: 640px) 100vw,
                   (max-width: 1024px) 80vw,
                   1920px"
        />
        <source
            type="image/webp"
            srcset="/assets/hero-320w.webp 320w,
                    /assets/hero-640w.webp 640w,
                    /assets/hero-1024w.webp 1024w,
                    /assets/hero-1920w.webp 1920w"
            sizes="(max-width: 640px) 100vw,
                   (max-width: 1024px) 80vw,
                   1920px"
        />
        <img
            src="/assets/hero-1024w.jpg"
            alt={ alt }
            loading="lazy"
            decoding="async"
            style="background-image: url('/assets/hero-placeholder.webp')"
            class="lazyload"
        />
    </picture>
}
```

## Rich Text Editor

### Goals

- Modern rich text editing with Lexical
- Store both JSON state and rendered HTML
- Secure HTML sanitization
- Preserve formatting for re-editing

### User Stories

- As a content creator, I want a modern rich text editor
- As a developer, I want to store Lexical's JSON state for editing
- As a developer, I want pre-rendered HTML for fast display
- As a security engineer, I want all user HTML sanitized
- As a user, I want my formatting preserved when editing

### Lexical Storage

```go
// internal/pkg/editor/types.go
package editor

type EditorContent struct {
    ID          int64           `db:"id"`
    LexicalJSON json.RawMessage `db:"lexical_json"`
    HTMLCache   string          `db:"html_cache"`
    UpdatedAt   time.Time       `db:"updated_at"`
}

// Sanitization policy
func NewLexicalSanitizer() *bluemonday.Policy {
    policy := bluemonday.NewPolicy()

    // Text formatting
    policy.AllowElements("p", "br", "strong", "b", "em", "i",
                         "u", "s", "del", "mark")

    // Headings
    policy.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")

    // Lists
    policy.AllowLists()

    // Blockquotes and code
    policy.AllowElements("blockquote", "code", "pre")
    policy.AllowAttrs("class").OnElements("code", "pre")

    // Links with security
    policy.AllowStandardURLs()
    policy.AllowAttrs("href", "title").OnElements("a")
    policy.RequireNoFollowOnLinks(true)
    policy.RequireNoFollowOnFullyQualifiedLinks(true)

    // Images from CDN only
    policy.AllowAttrs("src", "alt", "title", "width", "height").
        Matching(regexp.MustCompile(
            `^https://cdn\.example\.com/.*$`)).
        OnElements("img")

    return policy
}

// Save and sanitize content
func (s *EditorService) SaveContent(ctx context.Context,
                                    id int64,
                                    lexicalJSON json.RawMessage,
                                    htmlContent string) error {
    // Sanitize HTML
    sanitized := s.sanitizer.Sanitize(htmlContent)

    // Store both JSON and sanitized HTML
    return s.repo.UpdateContent(ctx, id, lexicalJSON, sanitized)
}
```

## Internationalization

### Goals

- Seamless multi-language support with minimal boilerplate
- Context-based locale detection and propagation
- Type-safe translation keys with compile-time validation
- Pluralization support for all languages
- Easy integration with templ templates

### User Stories

- As a developer, I want translations to work automatically based on Accept-Language headers
- As a developer, I want compile-time errors for missing translation keys
- As a content editor, I want to manage translations in simple YAML files
- As a user, I want the site to display in my preferred language
- As a developer, I want pluralization to work correctly for all languages

### i18n Architecture

Tracks uses **ctxi18n** for context-aware translations with excellent templ integration.

### i18n Middleware

```go
// internal/http/middleware/i18n.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/invopop/ctxi18n"
)

func I18n(bundle *ctxi18n.Bundle) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check for explicit locale in query/cookie
            locale := r.URL.Query().Get("locale")
            if locale == "" {
                if cookie, err := r.Cookie("locale"); err == nil {
                    locale = cookie.Value
                }
            }

            // Fall back to Accept-Language header
            if locale == "" {
                locale = parseAcceptLanguage(r.Header.Get("Accept-Language"))
            }

            // Set locale in context
            ctx := ctxi18n.WithLocale(r.Context(), locale)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func parseAcceptLanguage(header string) string {
    if header == "" {
        return "en"
    }

    // Parse "en-US,en;q=0.9,es;q=0.8" format
    parts := strings.Split(header, ",")
    if len(parts) > 0 {
        lang := strings.Split(parts[0], ";")[0]
        lang = strings.TrimSpace(lang)
        lang = strings.Split(lang, "-")[0] // Convert en-US to en
        return lang
    }

    return "en"
}
```

### Translation Files

```yaml
# internal/i18n/translations/en.yaml
common:
  app_name: "Tracks"
  welcome: "Welcome"
  login: "Log in"
  logout: "Log out"
  register: "Register"

auth:
  email_label: "Email address"
  email_placeholder: "you@example.com"
  otp_sent: "We've sent a code to {email}"
  otp_label: "Enter your verification code"

errors:
  not_found: "Page not found"
  unauthorized: "You must log in to continue"
  server_error: "Something went wrong"

posts:
  new: "New post"
  edit: "Edit post"
  delete: "Delete post"
  confirm_delete: "Are you sure you want to delete this post?"

  # Pluralization
  count:
    zero: "No posts"
    one: "1 post"
    other: "{count} posts"
```

### Using Translations in Templates

```go
// internal/http/views/pages/login.templ
package pages

import (
    "github.com/invopop/ctxi18n"
    "github.com/a-h/templ"
)

templ LoginPage(ctx context.Context) {
    <form method="post" action="/login">
        <label for="email">
            { ctxi18n.T(ctx, "auth.email_label") }
        </label>
        <input
            type="email"
            id="email"
            name="email"
            placeholder={ ctxi18n.T(ctx, "auth.email_placeholder") }
            required
        />
        <button type="submit">
            { ctxi18n.T(ctx, "common.login") }
        </button>
    </form>
}

// With pluralization
templ PostCount(ctx context.Context, count int) {
    <p>
        { ctxi18n.T(ctx, "posts.count", ctxi18n.M{"count": count}) }
    </p>
}
```

## CSS & JavaScript

### TailwindCSS Setup

```css
/* web/styles/app.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
    :root {
        --color-primary: theme('colors.blue.600');
        --color-secondary: theme('colors.gray.600');
    }

    [data-theme="dark"] {
        --color-primary: theme('colors.blue.400');
        --color-secondary: theme('colors.gray.400');
    }
}

@layer components {
    .btn {
        @apply px-4 py-2 rounded-lg font-medium transition-colors;
    }

    .btn-primary {
        @apply btn bg-blue-600 text-white hover:bg-blue-700;
    }

    .link {
        @apply text-blue-600 hover:text-blue-800 underline;
    }
}
```

### Alpine.js Components

```javascript
// web/scripts/app.js
import Alpine from 'alpinejs'
import morph from '@alpinejs/morph'
import persist from '@alpinejs/persist'

// Register plugins
Alpine.plugin(morph)
Alpine.plugin(persist)

// Theme switcher component
Alpine.data('theme', () => ({
    current: Alpine.$persist('auto').as('theme'),

    toggle() {
        const themes = ['light', 'dark', 'auto']
        const index = themes.indexOf(this.current)
        this.current = themes[(index + 1) % themes.length]
        this.apply()
    },

    apply() {
        if (this.current === 'auto') {
            const dark = window.matchMedia('(prefers-color-scheme: dark)').matches
            document.documentElement.dataset.theme = dark ? 'dark' : 'light'
        } else {
            document.documentElement.dataset.theme = this.current
        }
    },

    init() {
        this.apply()
        window.matchMedia('(prefers-color-scheme: dark)')
              .addEventListener('change', () => this.apply())
    }
}))

// Start Alpine
Alpine.start()
```

## Build Process

### Code Generation

Template generation is part of the standard `make generate` workflow:

```makefile
# Makefile
.PHONY: generate
generate:
	go tool sqlc generate
	go tool templ generate
	go tool mockery
```

Templ generates Go code from `.templ` files, similar to how SQLC generates from `.sql` files.

### Asset Building

```makefile
.PHONY: assets
assets: css js images

.PHONY: css
css:
	npx tailwindcss -i ./web/styles/app.css -o ./internal/assets/dist/app.css --minify

.PHONY: js
js:
	esbuild web/scripts/app.js \
		--bundle \
		--minify \
		--sourcemap \
		--target=es2020 \
		--outfile=internal/assets/dist/app.js

.PHONY: images
images:
	tracks image:prep web/images/*.{jpg,png} --all
```

## Best Practices

1. **Use templ-ui components for common UI patterns** - Don't rebuild what exists
2. **Customize components when needed** - You own the code, modify freely
3. **Use components for reusability** - Don't repeat HTML structures
4. **Always sanitize user HTML** - Use bluemonday policies
5. **Optimize images on build** - Don't serve unoptimized images
6. **Use content-addressed URLs** - Enables aggressive caching
7. **Keep translations organized** - Use nested keys for clarity
8. **Test template rendering** - Templates can have runtime errors
9. **Use Alpine for interactivity** - Keep JavaScript minimal
10. **Update components intentionally** - Run `tracks ui add <component>` to update from upstream

## Testing

```go
// internal/http/views/pages/home_test.go
func TestHomePage(t *testing.T) {
    ctx := context.Background()
    ctx = templ.WithNonce(ctx, "test-nonce")

    user := &User{Name: "Alice"}
    ctx = context.WithValue(ctx, "user", user)

    var buf bytes.Buffer
    err := HomePage(ctx).Render(ctx, &buf)
    assert.NoError(t, err)

    html := buf.String()
    assert.Contains(t, html, "Welcome to Tracks")
    assert.Contains(t, html, "Hello, Alice!")
    assert.Contains(t, html, "nonce=\"test-nonce\"")
}
```

## Next Steps

- Continue to [External Services →](./8_external_services.md)
- Back to [← Security](./6_security.md)
- Return to [Summary](./0_summary.md)
