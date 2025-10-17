# Tracks Monorepo Setup

This document describes the monorepo structure and tooling setup for the Tracks project.

## Repository Structure

```text
tracks/
├── cmd/                       # Go CLI commands
│   ├── tracks/               # Main CLI tool
│   └── tracks-mcp/           # MCP server
│
├── internal/                 # Private Go packages
│   ├── generator/
│   ├── tui/
│   └── mcp/
│
├── pkg/                      # Public Go packages
│
├── examples/                 # Example applications
│   ├── blog-example/
│   └── saas-example/
│
├── website/                  # Docusaurus documentation site
│   ├── docs/                # Documentation (versioned)
│   ├── blog/                # Blog posts
│   ├── src/                 # React components & pages
│   ├── static/              # Static assets
│   ├── docusaurus.config.js
│   └── package.json
│
├── .github/workflows/        # CI/CD
│   ├── ci.yml               # Tests & lints
│   ├── release.yml          # GoReleaser
│   └── deploy-website.yml   # Docusaurus deployment
│
├── .chglog/                  # git-chglog configuration
│
├── docs/                     # Project documentation
│   ├── prd/                 # Product requirements
│   └── RELEASING.md         # Release process
│
├── go.mod                    # Go modules
├── package.json              # Root workspace config
├── pnpm-workspace.yaml       # pnpm workspace
├── .goreleaser.yml           # Release automation
├── Dockerfile                # Docker image
├── Makefile                  # Development commands
└── README.md                 # Project readme
```

## Tooling Stack

### Go Release Automation

- **GoReleaser** - Builds binaries, Docker images, Homebrew formulas
- **git-chglog** - Generates changelogs from conventional commits
- Builds for: Linux, macOS, Windows (amd64 & arm64)
- Docker images published to GitHub Container Registry
- Homebrew formula auto-generated

### Documentation

- **Docusaurus 3** - Documentation + marketing site
- **MDX** - Markdown with React components
- **Algolia DocSearch** - Search (configured but disabled until applied)
- Deployed to GitHub Pages automatically on push to main

### Development

- **pnpm** - Package manager with workspace support
- **Conventional Commits** - Commit message format
- **golangci-lint** - Go linting
- **markdownlint-cli2** - Markdown linting

## Quick Start

### Prerequisites

```bash
# Go 1.23+
go version

# Node.js 24+ and pnpm
node --version
pnpm --version

# Install pnpm if needed
npm install -g pnpm@9
```

### Initial Setup

```bash
# Install dependencies
pnpm install

# Download Go dependencies
go mod download

# View all available commands
make help
```

### Development Workflow

```bash
# Start Docusaurus dev server
make website-dev

# Run Go tests
make test

# Run all linters
make lint

# Build binaries
make build-all
```

## Makefile Commands

### Testing

```bash
make test                # Run unit tests
make test-coverage       # Run tests with coverage report
make test-integration    # Run integration tests
make test-all            # Run all tests
```

### Building

```bash
make build               # Build tracks CLI
make build-mcp           # Build tracks-mcp server
make build-all           # Build all binaries
make install             # Install tracks to $GOPATH/bin
```

### Website

```bash
make website-dev         # Start dev server (hot reload)
make website-build       # Build for production
make website-serve       # Serve built site locally
```

### Linting

```bash
make lint-md             # Lint markdown files
make lint-md-fix         # Auto-fix markdown issues
make lint-go             # Run golangci-lint
make lint                # Run all linters
```

### Release

```bash
make changelog           # Generate CHANGELOG.md
make release-dry-run     # Test release locally
make release VERSION=v0.1.0  # Create and push release tag
```

### Utility

```bash
make deps                # Download and tidy Go dependencies
make clean               # Remove build artifacts
make help                # Show all commands
```

## Release Process

### 1. Make Changes

- Follow Conventional Commits format
- Write tests
- Update documentation

### 2. Generate Changelog

```bash
make changelog
git add CHANGELOG.md
git commit -m "chore: update changelog"
```

### 3. Create Release

```bash
# This will:
# - Update CHANGELOG.md
# - Create and push git tag
# - Trigger GoReleaser via GitHub Actions
make release VERSION=v0.1.0
```

### 4. GitHub Actions

The release workflow will:

- Build binaries for all platforms
- Create Docker images (amd64 & arm64)
- Publish to GitHub Releases
- Update Homebrew formula
- Generate release notes

## Website Deployment

### Automatic Deployment

Pushes to `main` that affect `website/**` or `docs/prd/**` trigger automatic deployment to GitHub Pages.

### Manual Deployment

```bash
# Build and deploy
make website-deploy

# Or use pnpm directly
pnpm run website:deploy
```

### Local Testing

```bash
# Development server with hot reload
make website-dev

# Build and serve production version
make website-build
make website-serve
```

## Conventional Commits

We follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

### Examples

```bash
git commit -m "feat(cli): add interactive resource generator"
git commit -m "fix(auth): handle expired tokens correctly"
git commit -m "docs: update installation instructions"
```

## CI/CD Workflows

### CI Workflow (`ci.yml`)

Runs on: Push & PR to `main`/`develop`

- Go tests with coverage
- golangci-lint
- Markdown linting

### Release Workflow (`release.yml`)

Runs on: Tag push (`v*`)

- Builds cross-platform binaries
- Creates Docker images
- Publishes GitHub Release
- Updates package managers

### Website Deployment (`deploy-website.yml`)

Runs on: Push to `main` affecting website files

- Builds Docusaurus site
- Deploys to GitHub Pages

## Next Steps

1. **Initialize Go Modules**

   ```bash
   go mod init github.com/anomalousventures/tracks
   ```

2. **Create Initial Code**
   - Implement cmd/tracks/main.go
   - Set up basic CLI structure

3. **Test Release Process**

   ```bash
   make release-dry-run
   ```

4. **Configure Algolia Search**
   - Apply for DocSearch
   - Update website/docusaurus.config.js

5. **Set up GitHub Pages**
   - Enable in repository settings
   - Configure custom domain (optional)

## Troubleshooting

### pnpm install fails

```bash
# Clear cache and reinstall
pnpm store prune
rm -rf node_modules
pnpm install
```

### GoReleaser fails

```bash
# Test locally
make release-dry-run

# Check configuration
goreleaser check
```

### Website build fails

```bash
# Clear Docusaurus cache
rm -rf website/.docusaurus website/build
make website-build
```

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Docusaurus Documentation](https://docusaurus.io/)
- [git-chglog Documentation](https://github.com/git-chglog/git-chglog)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [pnpm Workspaces](https://pnpm.io/workspaces)

## Questions?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines or open a [GitHub Discussion](https://github.com/anomalousventures/tracks/discussions).
