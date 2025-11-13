---
sidebar_position: 1
---

# Installation

This guide covers all the ways to install the Tracks CLI on your system.

## Prerequisites

Before installing Tracks, ensure you have:

- **Go 1.25 or later** - Required for building from source or using `go install`
- **Git** - For cloning the repository (source installation only)
- **CGO toolchain** (optional) - Required only if you plan to generate projects using LibSQL or SQLite databases:
  - **Linux**: `gcc`, `musl-dev` (Alpine), or `build-essential` (Debian/Ubuntu)
  - **macOS**: Xcode Command Line Tools (`xcode-select --install`)
  - **Windows**: MinGW-w64 or TDM-GCC

:::tip
If you only plan to use PostgreSQL databases, you don't need CGO tooling. PostgreSQL drivers are pure Go.
:::

## Installation Methods

Choose the installation method that best fits your workflow.

### Package Managers

The easiest way to install and keep Tracks up to date.

#### macOS - Homebrew

```bash
brew install anomalousventures/tap/tracks
```

#### Windows - Scoop

```bash
scoop bucket add anomalousventures https://github.com/anomalousventures/scoop-bucket
scoop install tracks
```

### Direct Binary Download

Download pre-built binaries from [GitHub Releases](https://github.com/anomalousventures/tracks/releases/latest):

- **Linux**: `tracks_linux_amd64.tar.gz`, `tracks_linux_arm64.tar.gz`
- **macOS**: `tracks_darwin_amd64.tar.gz`, `tracks_darwin_arm64.tar.gz` (Apple Silicon)
- **Windows**: `tracks_windows_amd64.zip`, `tracks_windows_arm64.zip`

#### Linux/macOS Example

```bash
# Download and extract (replace <version> and platform as needed)
curl -L https://github.com/anomalousventures/tracks/releases/download/<version>/tracks_linux_amd64.tar.gz | tar xz

# Move to PATH location
sudo mv tracks /usr/local/bin/

# Verify installation
tracks version
```

#### Windows Example (PowerShell)

```powershell
# Download and extract (replace <version> as needed)
Invoke-WebRequest -Uri "https://github.com/anomalousventures/tracks/releases/download/<version>/tracks_windows_amd64.zip" -OutFile "tracks.zip"
Expand-Archive -Path "tracks.zip" -DestinationPath "."

# Move to PATH location or add current directory to PATH
Move-Item "tracks.exe" "$env:USERPROFILE\bin\"
```

### Go Install

Install using Go's built-in package manager:

```bash
# Install specific version
go install github.com/anomalousventures/tracks/cmd/tracks@<version>

# Install latest release
go install github.com/anomalousventures/tracks/cmd/tracks@latest
```

:::note
Using `go install` requires Go 1.25+ and adds the binary to `$GOPATH/bin`. Make sure this directory is in your `$PATH`.
:::

### Docker

Run Tracks without installing it locally:

```bash
# Run tracks command via Docker (replace <version> as needed)
docker run --rm ghcr.io/anomalousventures/tracks:<version> version

# Pull specific version
docker pull ghcr.io/anomalousventures/tracks:<version>

# Use latest stable release
docker pull ghcr.io/anomalousventures/tracks:latest
```

To generate a project using Docker:

```bash
# Generate project in current directory
docker run --rm -v "$(pwd):/workspace" -w /workspace \
  ghcr.io/anomalousventures/tracks:latest new myapp
```

### From Source (Development Builds)

Install from the latest source code for development or testing unreleased features:

```bash
# Clone repository
git clone https://github.com/anomalousventures/tracks.git
cd tracks

# Build CLI
make build

# Binary available at ./bin/tracks
./bin/tracks version
```

For development work, you can also use:

```bash
# Install to $GOPATH/bin
make install

# Or run directly without installing
make dev version
```

## Verification

After installation, verify Tracks is working:

```bash
tracks version
```

You should see output like:

```text
Tracks CLI vX.X.X
Commit: abc1234
Built:  2025-01-15T10:30:00Z
Go:     go1.25.0
```

Test creating a new project:

```bash
# Show available options
tracks new --help

# Generate a test project (dry-run)
tracks new testapp --db=sqlite3
```

## Troubleshooting

### Command Not Found

If you get `command not found: tracks`, ensure the installation directory is in your `$PATH`:

```bash
# Check current PATH
echo $PATH

# For go install, add $GOPATH/bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# For manual installation, add /usr/local/bin
export PATH="$PATH:/usr/local/bin"
```

Make this permanent by adding the `export` line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.).

### Permission Denied (Linux/macOS)

If you get "permission denied" when running the binary:

```bash
# Make binary executable
chmod +x /path/to/tracks
```

### CGO Build Errors

If you see errors about missing C compiler when generating projects with LibSQL/SQLite:

**Linux (Debian/Ubuntu):**

```bash
sudo apt-get update
sudo apt-get install build-essential
```

**Linux (Alpine):**

```bash
apk add gcc musl-dev
```

**macOS:**

```bash
xcode-select --install
```

**Windows:**

Install MinGW-w64 via [MSYS2](https://www.msys2.org/) or use [TDM-GCC](https://jmeubank.github.io/tdm-gcc/).

:::tip
If you don't need LibSQL/SQLite, use PostgreSQL which doesn't require CGO:

```bash
tracks new myapp --db=postgres
```

:::

## Next Steps

Now that Tracks is installed, you're ready to create your first project:

- **[Quickstart Tutorial](./quickstart.md)** - Build your first Tracks application in 15 minutes
- **[CLI Commands](../cli/commands.md)** - Complete command reference
- **[tracks new Command](../cli/new.md)** - Detailed project generation guide

## Need Help?

- **GitHub Issues**: [Report bugs or request features](https://github.com/anomalousventures/tracks/issues)
- **Discussions**: [Ask questions and share ideas](https://github.com/anomalousventures/tracks/discussions)
- **Contributing**: [Development setup guide](https://github.com/anomalousventures/tracks/blob/main/CONTRIBUTING.md)
