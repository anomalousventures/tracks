---
sidebar_position: 1
---

# Commands Overview

Quick reference for all Tracks CLI commands.

## Available Commands

### [tracks new](new.mdx)

Create a new Tracks application with production-ready structure and database support.

### [tracks db](db.md)

Manage database migrations. Subcommands:

- `tracks db migrate` - Apply pending migrations
- `tracks db rollback` - Roll back last migration
- `tracks db status` - Show migration status
- `tracks db reset` - Reset database (rollback all, reapply)

### [tracks version](version.md)

Display version, commit, and build information.

### [tracks help](help.md)

Show help information for any command.

## Global Flags

All commands support these flags:

- `--json` - Output in JSON format
- `--no-color` - Disable colored output
- `--verbose`, `-v` - Verbose output
- `--quiet`, `-q` - Suppress non-error output
- `--help` - Show help for command

See [Output Modes](output-modes.md) for details on JSON and formatting options.

## See Also

- [CLI Overview](overview.mdx) - Getting started with Tracks CLI
- [Output Modes](output-modes.md) - JSON and console output formatting
