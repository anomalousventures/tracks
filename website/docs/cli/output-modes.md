---
sidebar_position: 3
---

# Output Modes

Tracks CLI supports multiple output modes to adapt to different environments and use cases.

## Mode Types

### Console Mode (Default)

Human-readable, colored terminal output.

**Features:**

- Colored text with semantic meaning
- Pretty tables with aligned columns
- Progress bars for long operations
- Optimized for human readability

**When Used:**

- Running in a TTY terminal
- Interactive development
- Default mode

**Example:**

```bash
$ tracks version
Tracks v0.1.0
Commit: abc123def
Built: 2025-10-24T08:00:00Z
```

**Colors:**

- **Title** - Bold purple
- **Success** - Green
- **Error** - Red
- **Warning** - Orange
- **Muted** - Gray

### JSON Mode

Machine-readable structured output for scripts and automation.

**Features:**

- Valid JSON output
- Pretty-printed with 2-space indentation
- Structured data for parsing
- No progress bars (not suitable for streaming)

**When Used:**

- Scripting and automation
- CI/CD pipelines
- Parsing output programmatically
- Using `--json` flag

**Example:**

```bash
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
```

**Parsing:**

```bash
# Extract version with jq
tracks --json version | jq -r '.title'

# Extract commit
tracks --json version | jq -r '.sections[0].body' | grep 'Commit:' | cut -d' ' -f2
```

### TUI Mode (Phase 4)

Interactive full-screen terminal interface.

**Planned Features:**

- Full-screen application
- Keyboard navigation
- Forms and prompts
- File browsers
- Real-time updates
- Mouse support

**When Used:**

- Interactive code generation
- Configuration wizards
- File selection
- Complex workflows

**Status:** Coming in Phase 4

## Mode Detection

The CLI automatically chooses the best mode based on your environment.

### Detection Priority

1. **`--json` flag** → JSON mode (highest priority)
2. **`--interactive` flag** → TUI mode (Phase 4)
3. **CI environment** → Console mode (no colors)
4. **Non-TTY** (piped/redirected) → Console mode
5. **TTY terminal** → Console mode with colors

### Examples

```bash
# Auto-detect (TTY terminal) → Console mode
tracks version

# Force JSON mode
tracks --json version

# CI environment (auto-detected) → Console mode, no colors
CI=true tracks version

# Piped output (auto-detected) → Console mode
tracks version | grep "Commit"

# Force interactive TUI (Phase 4)
tracks --interactive
```

## Color Control

### Disabling Colors

Colors are automatically disabled in:

- CI environments (CI env var set)
- Non-TTY output (pipes, redirects)
- When `NO_COLOR` is set

**Manual control:**

```bash
# Disable with flag
tracks --no-color version

# Disable with env var (standard)
NO_COLOR=1 tracks version

# Disable with Tracks env var
TRACKS_NO_COLOR=true tracks version
```

### Checking Color Support

The CLI respects:

- `NO_COLOR` - Standard environment variable
- `CLICOLOR` - Standard color control
- TTY detection via `isatty`
- CI environment detection

## Output Structure

### Console Mode

**Format:**

```text
<Title>                    # Bold, colored
<blank line>
<Section Title>            # If present
<Section Body>             # Plain text
<blank line>
<Table Headers>            # Bold, colored
<Table Separator>          # Dashes
<Table Rows>               # Aligned columns
<blank line>
<Progress Bar>             # If applicable
```

**Example:**

```text
Installation Complete

Successfully installed dependencies

Package         Version    Status
────────────────────────────────
cobra           1.10.1     active
viper           1.20.1     active

Downloading [████████████] 100%
```

### JSON Mode

**Schema:**

```json
{
  "title": "string (optional)",
  "sections": [
    {
      "title": "string (optional)",
      "body": "string"
    }
  ],
  "tables": [
    {
      "headers": ["string", ...],
      "rows": [
        ["string", ...],
        ...
      ]
    }
  ]
}
```

**Notes:**

- Top-level fields are optional
- Empty arrays/strings may be omitted
- Progress is not included (JSON is static)

## Environment Variables

Control output mode via environment variables:

```bash
# Force JSON mode
export TRACKS_JSON=true

# Disable colors
export NO_COLOR=1
# or
export TRACKS_NO_COLOR=true

# Force interactive (Phase 4)
export TRACKS_INTERACTIVE=true

# Set log level
export TRACKS_LOG_LEVEL=debug  # debug, info, warn, error, off
```

**Priority:** Flags > Env Vars > Auto-detection

## Use Cases

### Scripting

Use JSON mode for reliable parsing:

```bash
#!/bin/bash

# Get version info
VERSION_JSON=$(tracks --json version)

# Extract fields
TITLE=$(echo "$VERSION_JSON" | jq -r '.title')
BODY=$(echo "$VERSION_JSON" | jq -r '.sections[0].body')

echo "Version: $TITLE"
echo "Details: $BODY"
```

### CI/CD Integration

Console mode works great in CI:

```yaml
# GitHub Actions
- name: Check version
  run: |
    tracks version          # Console output, no colors
    tracks --json version   # Structured for parsing
```

### Development

Console mode is optimized for development:

```bash
# Colored, readable output
tracks version

# With verbose details
tracks -v version

# Quiet (only errors)
tracks -q version
```

### Automation

JSON mode for tools and scripts:

```python
import subprocess
import json

# Run CLI, capture JSON
result = subprocess.run(
    ['tracks', '--json', 'version'],
    capture_output=True,
    text=True
)

# Parse output
data = json.loads(result.stdout)
version = data['title']
commit = data['sections'][0]['body']

print(f"Tracks {version}")
```

## Examples

### Console Mode Examples

**Success output:**

```bash
$ tracks new myapp
Creating new Tracks application: myapp

✓ Project structure created
✓ Dependencies installed
✓ Git repository initialized

Next steps:
  cd myapp
  tracks dev
```

**Table output:**

```bash
$ tracks generate list
Available Generators

Name            Description           Status
───────────────────────────────────────────
resource        Full CRUD resource    stable
model           Database model        stable
controller      HTTP controller       beta
```

**Progress bar:**

```bash
$ tracks install deps
Installing dependencies [████████████████████] 100%
```

### JSON Mode Examples

**Version:**

```json
{
  "title": "Tracks v0.1.0",
  "sections": [
    {
      "title": "",
      "body": "Commit: abc123\nBuilt: 2025-10-24"
    }
  ]
}
```

**List with table:**

```json
{
  "title": "Available Generators",
  "tables": [
    {
      "headers": ["Name", "Description", "Status"],
      "rows": [
        ["resource", "Full CRUD resource", "stable"],
        ["model", "Database model", "stable"]
      ]
    }
  ]
}
```

## Tips

### Best Practices

1. **Use JSON for scripting** - Reliable, structured output
2. **Use console for interactive** - Better user experience
3. **Disable colors in CI** - Automatic, no action needed
4. **Check exit codes** - Don't rely on output parsing
5. **Respect NO_COLOR** - Standard environment variable

### Common Mistakes

**Don't parse console output:**

```bash
# BAD - brittle, breaks with format changes
version=$(tracks version | grep "Tracks" | cut -d' ' -f2)

# GOOD - use JSON mode
version=$(tracks --json version | jq -r '.title' | cut -d' ' -f2)
```

**Don't force colors in CI:**

```bash
# BAD - ANSI codes pollute logs
tracks version  # In CI, colors auto-disabled

# GOOD - explicit if needed
tracks --no-color version
```

**Don't ignore errors:**

```bash
# BAD - no error checking
output=$(tracks version)

# GOOD - check exit code
if ! output=$(tracks version); then
  echo "Command failed!"
  exit 1
fi
```

## Related Documentation

- [CLI Overview](./overview.mdx) - Getting started
- [Commands Reference](./commands.md) - Complete command list
- [Contributing](https://github.com/anomalousventures/tracks/blob/main/CONTRIBUTING.md) - Development guide
