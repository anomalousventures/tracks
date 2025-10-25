---
sidebar_position: 2
---

# Commands Reference

Complete reference for all Tracks CLI commands.

## `tracks`

Root command - displays help or interactive TUI (Phase 4).

**Usage:**

```bash
tracks [flags]
```

**Behavior:**

- Without arguments: Shows message about TUI mode (Phase 4)
- With `--help`: Shows full help information
- With `--version`: Shows version (same as `tracks version`)

**Examples:**

```bash
# Show help
tracks --help

# JSON output
tracks --json
```

## `tracks version`

Display version information.

**Usage:**

```bash
tracks version [flags]
```

**Output:**

Shows tracks version, git commit, and build date.

**Examples:**

```bash
# Console output
$ tracks version
Tracks v0.1.0
Commit: abc123def
Built: 2025-10-24T08:00:00Z

# JSON output
$ tracks --json version
{
  "title": "Tracks v0.1.0",
  "sections": [
    {
      "title": "",
      "body": "Commit: abc123def\nBuilt: 2025-10-24T08:00:00Z"
    }
  ]
}

# Quiet output (just version)
$ tracks -q version
v0.1.0

# Verbose output (more details)
$ tracks -v version
Tracks CLI v0.1.0
Git Commit: abc123def456
Build Date: 2025-10-24T08:00:00Z
Go Version: go1.25.3
Platform: linux/amd64
```

## `tracks new`

Create a new Tracks application.

**Status:** Placeholder implementation (full implementation coming soon)

**Usage:**

```bash
tracks new <project-name> [flags]
```

**Arguments:**

- `<project-name>` - Name of the project to create (required)

**Examples:**

```bash
# Create new application
tracks new myapp

# With verbose output
tracks -v new myapp

# JSON output
tracks --json new myapp
```

**Future Flags** (planned):

- `--db <driver>` - Database driver (postgres, sqlite, libsql)
- `--module <path>` - Custom Go module path
- `--no-git` - Skip git initialization
- `--auth <method>` - Authentication method (session, jwt)

## `tracks help`

Show help information for any command.

**Usage:**

```bash
tracks help [command]
```

**Examples:**

```bash
# General help
tracks help

# Command-specific help
tracks help version
tracks help new
```

## Global Flags

These flags work with all commands:

### `--json`

Output in JSON format for scripting and automation.

```bash
tracks --json version
tracks version --json  # Can come before or after command
```

### `--no-color`

Disable colored output (also respects `NO_COLOR` env var).

```bash
tracks --no-color version
```

### `--interactive`

Force interactive TUI mode even in non-TTY environments (Phase 4).

```bash
tracks --interactive
```

### `-v, --verbose`

Enable verbose output with additional details.

```bash
tracks -v version
tracks --verbose new myapp
```

### `-q, --quiet`

Suppress non-error output (only show errors).

```bash
tracks -q version
```

**Note:** `--verbose` and `--quiet` are mutually exclusive.

### `--help`

Show help information for any command.

```bash
tracks --help
tracks version --help
```

### `--version`

Show version information (shorthand for `tracks version`).

```bash
tracks --version
```

## Environment Variables

All flags can be set via environment variables with `TRACKS_` prefix:

```bash
# Force JSON mode
export TRACKS_JSON=true
tracks version

# Disable colors
export TRACKS_NO_COLOR=true
tracks version

# Set log level
export TRACKS_LOG_LEVEL=debug
tracks version

# Standard NO_COLOR (no prefix)
export NO_COLOR=1
tracks version
```

Environment variables are overridden by command-line flags.

## Exit Codes

The CLI uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of command (wrong arguments, invalid flags)

## Examples

### Scripting with JSON

```bash
#!/bin/bash

# Get version info as JSON
version_json=$(tracks --json version)

# Extract version with jq
version=$(echo "$version_json" | jq -r '.title | split(" ")[1]')

echo "Running tracks $version"
```

### CI/CD Integration

```yaml
# .github/workflows/deploy.yml
- name: Check Tracks version
  run: |
    tracks version
    tracks --json version > tracks-version.json
```

### Quiet Mode for Scripts

```bash
# Only show errors
if tracks -q new myapp; then
  echo "Success!"
else
  echo "Failed!"
  exit 1
fi
```

## Coming Soon

Future commands in development:

- `tracks generate` - Generate code (models, controllers, etc.)
- `tracks dev` - Start development server
- `tracks db migrate` - Run database migrations
- `tracks db seed` - Seed database with data
- `tracks deploy` - Deploy application

See the [Roadmap](https://github.com/anomalousventures/tracks/blob/main/docs/roadmap/README.md) for details.

## Related Documentation

- [CLI Overview](./overview.md) - Getting started
- [Output Modes](./output-modes.md) - Output format details
- [Contributing](https://github.com/anomalousventures/tracks/blob/main/CONTRIBUTING.md) - Development guide
