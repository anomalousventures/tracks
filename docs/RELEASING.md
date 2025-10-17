# Release Process

This document outlines the versioning strategy, release process, and tagging conventions for Tracks.

## Versioning Strategy

Tracks follows [Semantic Versioning 2.0.0](https://semver.org/):

```text
MAJOR.MINOR.PATCH

Example: 1.2.3
- 1 = Major version
- 2 = Minor version
- 3 = Patch version
```

### Version Increment Rules

#### MAJOR (Breaking Changes)

Increment when making incompatible API changes:

- Removing or renaming public APIs
- Changing function signatures
- Changing generated code structure
- Removing CLI commands or flags
- Changing configuration file format

**Example**: `1.5.2 → 2.0.0`

#### MINOR (New Features)

Increment when adding functionality in a backward-compatible manner:

- New CLI commands
- New code generators
- New configuration options
- New middleware or utilities
- Performance improvements

**Example**: `1.5.2 → 1.6.0`

#### PATCH (Bug Fixes)

Increment when making backward-compatible bug fixes:

- Bug fixes
- Documentation updates
- Dependency updates (security)
- Minor refactoring

**Example**: `1.5.2 → 1.5.3`

### Pre-release Versions

During development, use pre-release identifiers:

```text
0.1.0-alpha.1    # Early development
0.1.0-beta.1     # Feature complete, testing
0.1.0-rc.1       # Release candidate
```

## Current Version Status

Tracks is currently in **pre-1.0 development** (v0.x.x):

- **API may change** without notice
- **Breaking changes** can occur in minor versions
- **Use in production** at your own risk
- **Feedback welcome** to shape the API

### Path to 1.0

Version 1.0.0 will be released when:

1. ✅ Core features are stable
2. ✅ Documentation is complete
3. ✅ Test coverage is >80%
4. ✅ Production usage by early adopters
5. ✅ API is finalized and documented
6. ✅ Migration guides are available

## Git Workflow

### Branch Strategy

```text
main              # Stable releases only
├── develop       # Integration branch
├── feature/*     # New features
├── fix/*         # Bug fixes
└── release/*     # Release preparation
```

### Branch Naming

```bash
# Features
feature/add-oauth-providers
feature/improve-tui-performance

# Bug fixes
fix/handle-nil-pointer-in-auth
fix/migration-rollback-error

# Releases
release/v1.2.0
release/v2.0.0-beta.1
```

## Release Process

### 1. Preparation

```bash
# Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/v0.5.0

# Update version in relevant files
# - package.json
# - internal/version/version.go
# - README.md (if referenced)

# Update CHANGELOG.md
# Add release notes under new version heading
```

### 2. Version File Updates

```go
// internal/version/version.go
package version

const (
    Version = "0.5.0"
    Commit  = "set by build"
    Date    = "set by build"
)
```

```json
// package.json
{
  "version": "0.5.0"
}
```

### 3. Generate Changelog

```bash
# Install changelog generator if not present
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

# Generate changelog for new version
git-chglog --next-tag v0.5.0 -o CHANGELOG.md

# Or manually update CHANGELOG.md with:
# - New features
# - Bug fixes
# - Breaking changes
# - Deprecations
```

### 4. Testing

```bash
# Run full test suite
make test-all

# Run linters
make lint

# Test build
make build

# Test installation
go install ./cmd/tracks

# Manual smoke tests
tracks version
tracks new test-app
cd test-app
tracks generate resource post title:string
tracks dev
```

### 5. Create Release PR

```bash
# Commit version bump
git add .
git commit -m "chore: bump version to v0.5.0"

# Push release branch
git push origin release/v0.5.0

# Create PR: release/v0.5.0 → main
# Title: "Release v0.5.0"
# Body: Include changelog excerpt
```

### 6. Merge and Tag

After PR approval:

```bash
# Merge to main
git checkout main
git pull origin main

# Create annotated tag
git tag -a v0.5.0 -m "Release v0.5.0

New Features:
- Feature 1
- Feature 2

Bug Fixes:
- Fix 1
- Fix 2

Breaking Changes:
- Change 1
"

# Push tag
git push origin v0.5.0

# Merge main back to develop
git checkout develop
git merge main
git push origin develop
```

### 7. GitHub Release

1. Go to GitHub Releases
2. Click "Draft a new release"
3. Select the tag (v0.5.0)
4. Title: "Tracks v0.5.0"
5. Description: Copy from CHANGELOG.md
6. Add binary attachments (if applicable)
7. Check "Pre-release" if < v1.0.0
8. Publish release

### 8. Announcement

- Update project website
- Post in GitHub Discussions
- Tweet/social media
- Update documentation site

## Tag Conventions

### Format

```text
v{MAJOR}.{MINOR}.{PATCH}[-{PRERELEASE}]

Examples:
v0.1.0
v0.5.0-alpha.1
v0.9.0-beta.2
v0.9.0-rc.1
v1.0.0
v1.2.3
v2.0.0
```

### Tag Types

#### Release Tags

```bash
# Stable release
git tag -a v1.0.0 -m "Release v1.0.0"

# Pre-release
git tag -a v1.0.0-beta.1 -m "Beta release v1.0.0-beta.1"
```

#### Lightweight vs Annotated

**Always use annotated tags** for releases:

```bash
# GOOD: Annotated (includes metadata)
git tag -a v1.0.0 -m "Release v1.0.0"

# BAD: Lightweight (just a pointer)
git tag v1.0.0
```

Annotated tags include:

- Tagger name and email
- Date
- Message
- GPG signature (if configured)

### Tag Signing

For security, sign release tags:

```bash
# Setup GPG key
gpg --gen-key

# Configure Git
git config --global user.signingkey YOUR_KEY_ID
git config --global tag.gpgSign true

# Create signed tag
git tag -s v1.0.0 -m "Release v1.0.0"

# Verify signature
git tag -v v1.0.0
```

## Changelog Format

### Structure

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Feature in development

## [1.0.0] - 2025-03-15

### Added
- OAuth provider support
- PostgreSQL driver option
- Health check endpoints

### Changed
- Improved TUI performance
- Updated dependencies

### Fixed
- Migration rollback issue
- Session timeout bug

### Deprecated
- Old configuration format (use new YAML format)

### Removed
- Legacy authentication method

### Security
- Fixed SQL injection vulnerability in query builder

## [0.9.0] - 2025-02-01
...
```

### Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Vulnerability fixes

## Automation

### GitHub Actions Workflow

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run tests
        run: make test

      - name: Build binaries
        run: make build-all

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          body_path: RELEASE_NOTES.md
          draft: false
          prerelease: ${{ contains(github.ref, 'alpha') || contains(github.ref, 'beta') || contains(github.ref, 'rc') }}
```

## Release Checklist

Before releasing, ensure:

- [ ] All tests pass
- [ ] Linters pass
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version numbers are bumped
- [ ] Migration guide exists (if breaking changes)
- [ ] Security review completed (for security fixes)
- [ ] Performance regression tests pass
- [ ] Example applications work
- [ ] Installation instructions are current

## Emergency Releases

For critical security fixes:

1. Create hotfix branch from main
2. Apply minimal fix
3. Test thoroughly
4. Create patch release (increment PATCH)
5. Tag and release ASAP
6. Backport to develop

```bash
git checkout main
git checkout -b hotfix/security-fix
# Apply fix
git commit -m "fix: critical security vulnerability"
git checkout main
git merge hotfix/security-fix
git tag -a v1.0.1 -m "Security fix"
git push origin main v1.0.1
git checkout develop
git merge main
```

## Version Support

### Support Policy

- **Latest major version**: Full support
- **Previous major version**: Security fixes only (6 months)
- **Older versions**: No support

Example:

- v2.x.x - Full support
- v1.x.x - Security fixes until 6 months after v2.0.0
- v0.x.x - No support after v1.0.0

### Deprecation Policy

Features marked as deprecated:

1. Announced in CHANGELOG
2. Warning messages added to code
3. Documentation updated with migration path
4. Removed in next major version

Minimum deprecation period: 3 months or 1 major version

## Communication

### Release Announcements

Publish announcements on:

1. GitHub Releases
2. GitHub Discussions
3. Project blog/website
4. Twitter/social media
5. Go community forums (if major release)

### Migration Guides

For breaking changes, provide:

- Clear explanation of changes
- Code examples (before/after)
- Automated migration tools (if possible)
- Timeline for deprecated feature removal

## Questions?

For questions about the release process:

- Open a GitHub Discussion
- Email: releases@anomalousventures.com
- See CONTRIBUTING.md for general contribution guidelines

---

**Last Updated**: January 2025
