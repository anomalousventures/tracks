# tracks version

Display version, commit, and build information.

## Usage

```bash
tracks version [flags]
```

## Description

Shows Tracks version, git commit hash, and build timestamp.

## Examples

### Console output

```bash
$ tracks version
Tracks v0.1.0
Commit: abc123def
Built: 2025-10-24T08:00:00Z
```

### JSON output

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

### Quiet output

```bash
$ tracks -q version
v0.1.0
```

### Verbose output

```bash
$ tracks -v version
Tracks CLI v0.1.0
Git Commit: abc123def456
Build Date: 2025-10-24T08:00:00Z
Go Version: go1.25.3
Platform: linux/amd64
```

## Flags

Supports all [global flags](./commands.md#global-flags).

## Scripting

Extract version for scripts:

```bash
# Get just the version
version=$(tracks -q version)

# Parse JSON output
version_json=$(tracks --json version)
version=$(echo "$version_json" | jq -r '.title | split(" ")[1]')
```

## Shorthand

```bash
# --version flag shows same output
tracks --version
```

## See Also

- [Commands Reference](commands.md) - All available commands
- [Output Modes](output-modes.md) - JSON and console output
- [tracks help](help.md) - Get help on any command
