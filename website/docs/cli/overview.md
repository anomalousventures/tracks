---
sidebar_position: 1
---

# CLI Overview

The Tracks CLI is a command-line tool for creating and managing Go web applications. It generates production-ready code following Go best practices.

## Installation

### From Source (Current)

```bash
# Clone repository
git clone https://github.com/anomalousventures/tracks.git
cd tracks

# Build CLI
make build

# Binary available at ./bin/tracks
./bin/tracks version
```

### From Release (v0.1.0+)

#### Package Managers

```bash
# macOS (Homebrew)
brew install anomalousventures/tap/tracks

# Windows (Scoop)
scoop bucket add anomalousventures https://github.com/anomalousventures/scoop-bucket
scoop install tracks

# Go install
go install github.com/anomalousventures/tracks/cmd/tracks@v0.1.0
```

#### Docker

```bash
# Run tracks via Docker
docker run --rm ghcr.io/anomalousventures/tracks:0.1.0 version

# Pull specific version
docker pull ghcr.io/anomalousventures/tracks:0.1.0

# Use latest
docker pull ghcr.io/anomalousventures/tracks:latest
```

#### Direct Download

Download pre-built binaries from [GitHub Releases](https://github.com/anomalousventures/tracks/releases/latest):

- **Linux**: `tracks_linux_amd64.tar.gz`, `tracks_linux_arm64.tar.gz`
- **macOS**: `tracks_darwin_amd64.tar.gz`, `tracks_darwin_arm64.tar.gz`
- **Windows**: `tracks_windows_amd64.zip`, `tracks_windows_arm64.zip`

```bash
# Example: Linux AMD64
curl -L https://github.com/anomalousventures/tracks/releases/download/v0.1.0/tracks_linux_amd64.tar.gz | tar xz
sudo mv tracks /usr/local/bin/
```

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
