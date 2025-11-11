---
sidebar_position: 1
---

# CLI Overview

The Tracks CLI is a command-line tool for creating and managing Go web applications. It generates production-ready code following Go best practices.

:::tip Installation
See the **[Installation Guide](../getting-started/installation.md)** for setup instructions, including package managers, Docker, and direct downloads.
:::

## Key Features

- **Multiple Output Modes** - Console, JSON, and TUI (Phase 4). See [Output Modes](./output-modes.md)
- **Smart Mode Detection** - Automatically adapts to CI, TTY, and piped environments
- **Cross-Platform** - Works on Linux, macOS, and Windows
- **Environment Variable Support** - Configure via `TRACKS_*` variables

See [Commands Reference](./commands.md) for available commands and flags.

## Getting Help

```bash
# General help
tracks --help

# Command-specific help
tracks version --help
tracks new --help

# List all commands
tracks help
```

## What's Next?

- [Commands Reference](./commands.md) - Complete command list
- [Output Modes](./output-modes.md) - Detailed output format guide
- [Contributing Guide](https://github.com/anomalousventures/tracks/blob/main/CONTRIBUTING.md) - Development guide
