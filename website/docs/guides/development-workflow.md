# Development Workflow Guide

Learn how to use the development tools in Tracks-generated applications for an efficient coding experience.

## Overview

Tracks-generated projects include a complete development workflow powered by [Air](https://github.com/air-verse/air) for live reload. When you run `make dev`, the server automatically rebuilds and restarts whenever you change code.

## Quick Start

```bash
# Start the development server
make dev
```

This command:

1. Starts Air's file watcher
2. Watches for changes to source files
3. Runs `make generate assets` before each build
4. Compiles and restarts the server

> **Note:** Air is installed as a Go tool dependency. The `go tool air` command is
> automatically available after running `go mod download`. No global installation needed.

## What Gets Watched

Air monitors these file types for changes:

| Extension | Purpose |
|-----------|---------|
| `.go` | Go source files |
| `.templ` | templ template files |
| `.css` | Stylesheets (Tailwind) |
| `.js` | JavaScript files |
| `.tpl`, `.tmpl`, `.html` | Other template files |

When any of these files change, Air triggers a rebuild.

## What Gets Excluded

To prevent infinite rebuild loops and unnecessary rebuilds, these are excluded:

| Pattern | Reason |
|---------|--------|
| `*_templ.go` | Generated templ output |
| `internal/assets/dist` | Compiled assets |
| `internal/db/generated` | Generated SQLC code |
| `tests/mocks` | Generated mock files |
| `*_test.go` | Test files |
| `tmp` | Air's working directory |
| `node_modules` | npm dependencies |
| `vendor` | Go vendor directory |

## Build Pipeline

Before each Go compilation, Air runs `make generate assets`:

```text
File change detected
        │
        ▼
┌─────────────────────┐
│   make generate     │
├─────────────────────┤
│ • templ generate    │
│ • mockery           │
│ • sqlc generate     │
└─────────────────────┘
        │
        ▼
┌─────────────────────┐
│   make assets       │
├─────────────────────┤
│ • make css          │
│ • make js           │
└─────────────────────┘
        │
        ▼
┌─────────────────────┐
│     go build        │
└─────────────────────┘
        │
        ▼
   Server restart
```

## Asset Pipeline Details

### CSS (Tailwind)

```bash
make css
```

Compiles `internal/assets/web/css/app.css` using Tailwind CSS v4:

- Input: `internal/assets/web/css/app.css`
- Output: `internal/assets/dist/css/app.css`
- Features: Minification, tree-shaking

### JavaScript (esbuild)

```bash
make js
```

Bundles JavaScript using esbuild:

- Input: `internal/assets/web/js/*.js`
- Output: `internal/assets/dist/js`
- Features: Bundling, minification

## Directory Structure

```text
internal/assets/
├── web/                  # Source files (you edit these)
│   ├── css/
│   │   └── app.css       # Tailwind entry point
│   ├── js/
│   │   ├── app.js        # Main JS entry
│   │   └── lib/
│   │       └── htmx.js   # HTMX import
│   └── images/
│       └── ...
│
└── dist/                 # Built files (generated, gitignored)
    ├── css/
    │   └── app.css       # Compiled CSS
    └── js/
        └── app.js        # Bundled JS
```

## Air Configuration

The `.air.toml` file controls the development workflow:

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  delay = 1000
  include_ext = ["go", "tpl", "tmpl", "html", "css", "js", "templ"]
  exclude_dir = ["internal/assets/dist", "internal/db/generated", "tests/mocks", "tmp", "vendor", "node_modules"]
  exclude_regex = ["_test.go", "_templ.go"]
  pre_cmd = ["make generate assets"]
```

Key settings:

- **delay**: Wait 1 second after file changes before rebuilding (debouncing)
- **include_ext**: File types to watch
- **exclude_dir**: Directories to ignore
- **exclude_regex**: File patterns to ignore
- **pre_cmd**: Commands to run before Go build

## Common Tasks

### Manually Rebuild Assets

```bash
# Rebuild both CSS and JS
make assets

# Rebuild CSS only
make css

# Rebuild JS only
make js
```

### Regenerate Code

```bash
# Regenerate all (templ + mocks + SQLC)
make generate

# Regenerate templ templates only
make templ

# Regenerate mocks only
make mocks

# Regenerate SQL code only
make sqlc
```

### Check What Air Will Do

Before starting, you can verify your Air configuration:

```bash
cat .air.toml
```

## Troubleshooting

### Rebuild Loop

**Symptom:** Server keeps rebuilding indefinitely.

**Cause:** Air is watching generated output files.

**Solution:** Ensure `internal/assets/dist` is in `exclude_dir` and `_templ.go` is in `exclude_regex`.

### Assets Not Updating

**Symptom:** CSS/JS changes don't appear in browser.

**Cause:** Browser cache or missing asset rebuild.

**Solution:**

1. Hard refresh (Ctrl+Shift+R or Cmd+Shift+R)
2. Verify `make assets` runs on each rebuild
3. Check that `css` and `js` are in `include_ext`

### Templ Changes Not Reflected

**Symptom:** HTML template changes don't appear.

**Cause:** templ not regenerating.

**Solution:**

1. Verify `templ` is in `include_ext`
2. Check that `make generate` runs via `pre_cmd`
3. Ensure `_templ.go` is excluded to prevent double builds

### Slow Rebuilds

**Symptom:** Rebuilds take too long.

**Cause:** Unnecessary regeneration or large dependencies.

**Solution:**

1. Increase `delay` to batch rapid changes
2. Ensure `node_modules` is excluded
3. Use `exclude_unchanged = true` if helpful

### Port Already in Use

**Symptom:** `bind: address already in use`

**Cause:** Previous server instance still running.

**Solution:**

```bash
# Find and kill the process
lsof -i :8080 | grep LISTEN
kill <PID>

# Or use a different port
APP_SERVER_PORT=8081 make dev
```

## Environment Variables

Configure the development server with environment variables:

```bash
# Server port (default: 8080)
APP_SERVER_PORT=3000 make dev

# Enable debug logging
APP_LOG_LEVEL=debug make dev

# Database URL
APP_DATABASE_URL="postgres://..." make dev
```

See `.env.example` for all available options.

## Related Topics

- [Architecture Overview](/docs/guides/architecture-overview) - Application structure
- [Asset Caching](/docs/guides/caching) - How assets are served in production
- [Testing Guide](/docs/guides/testing) - Running tests
