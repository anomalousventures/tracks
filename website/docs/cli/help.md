# tracks help

Show help information for commands.

## Usage

```bash
tracks help [command]
```

## Description

Displays help text for Tracks commands with usage, flags, and examples.

## Examples

### General help

```bash
$ tracks help
A Rails-like web framework for Go

Usage:
  tracks [command]

Available Commands:
  new         Create a new Tracks application
  version     Show version information
  help        Help about any command

Flags:
  --json              Output in JSON format
  --no-color          Disable colored output
  -v, --verbose       Verbose output
  -q, --quiet         Quiet mode
  -h, --help          Help for tracks

Use "tracks [command] --help" for more information about a command.
```

### Command-specific help

```bash
$ tracks help new
Create a new Tracks application

Usage:
  tracks new <project-name> [flags]

Flags:
  --db string       Database driver (go-libsql, sqlite3, postgres)
  --module string   Go module path
  --no-git          Skip git initialization

Global Flags:
  --json        Output in JSON format
  --no-color    Disable colored output
```

## Shorthand

```bash
# --help flag shows same output
tracks --help
tracks new --help
```

## Flags

Supports all [global flags](./commands.md#global-flags).

## See Also

- [Commands Reference](commands.md) - All available commands
- [Output Modes](output-modes.md) - JSON and console output
- [tracks new](new.mdx) - Create a new Tracks application
- [tracks version](version.md) - Display version information
