---
sidebar_position: 1
---

# Commands Overview

Quick reference for all Tracks CLI commands.

## Available Commands

### [`tracks new`](./new.md)

Create a new Tracks application with production-ready structure.

```bash
tracks new myapp
tracks new myapp --db postgres --module github.com/me/myapp
```

### [`tracks version`](./version.md)

Display version, commit, and build information.

```bash
tracks version
tracks --json version
```

### [`tracks help`](./help.md)

Show help information for commands.

```bash
tracks help
tracks help new
```

## Global Flags

All commands support these flags:

- `--json` - Output in JSON format
- `--no-color` - Disable colored output
- `--verbose`, `-v` - Verbose output
- `--quiet`, `-q` - Suppress non-error output
- `--help` - Show help for command

See [Output Modes](./output-modes.md) for details on JSON and formatting options.
