# Release Process

This document provides comprehensive guidance for creating releases of Tracks. It documents lessons learned from v0.1.0 and establishes best practices for future releases.

## Quick Reference

**For experienced maintainers:**
See [.github/RELEASE_CHECKLIST.md](../.github/RELEASE_CHECKLIST.md) for the quick checklist.

**For first-time releases or troubleshooting:**
Read this document thoroughly.

## Prerequisites

### Required Tools

- **Go 1.25+** - For building and testing
- **git** - For tagging and version control
- **gh CLI** - For GitHub API interactions
- **make** - For running automation targets

### Required Secrets

The following GitHub repository secrets must be configured:

1. **`TRACKS_RELEASER_TOKEN`** (required)
   - Personal Access Token (PAT) with `repo` scope
   - Must have write access to:
     - `anomalousventures/tracks` (main repo)
     - `anomalousventures/homebrew-tap`
     - `anomalousventures/scoop-bucket`
   - Used by GoReleaser to publish to Homebrew and Scoop
   - **Note:** The default `GITHUB_TOKEN` cannot write to separate repositories

2. **`GITHUB_TOKEN`** (automatic)
   - Automatically provided by GitHub Actions
   - Used for Docker registry (ghcr.io) and main repo operations

### Access Requirements

- Write access to `anomalousventures/tracks` repository
- Permissions to create tags and releases
- Permissions to merge PRs (if using PR workflow)

## Release Workflow

### Phase 1: Preparation

#### 1. Verify Clean State

```bash
git checkout main
git pull
git status  # Should show clean working tree
```

#### 2. Run Full Test Suite

```bash
make lint    # All linters must pass
make test    # All tests must pass
```

**Critical:** Do not proceed if linting or tests fail.

#### 3. Update CHANGELOG.md

The CHANGELOG.md file must be updated BEFORE creating the release tag.

**Why:** GoReleaser includes CHANGELOG.md in release archives (.tar.gz, .zip). If the file doesn't exist or is outdated, the release will fail or ship with incorrect changelog.

**How:**

1. Generate changelog from commits:
   ```bash
   go tool git-chglog -o CHANGELOG.md
   ```

2. Edit CHANGELOG.md to add release summary:
   - Add a human-friendly summary at the top of the version section
   - Include key highlights (3-5 bullet points)
   - Explain what's new for users downloading archives
   - See v0.1.0 in CHANGELOG.md for example format

3. Fix any markdown linting issues:
   ```bash
   make lint-md
   ```

4. Create PR for CHANGELOG:
   ```bash
   git checkout -b chore/changelog-v0.x.0
   git add CHANGELOG.md
   git commit -m "chore: add CHANGELOG.md for v0.x.0 release"
   git push -u origin chore/changelog-v0.x.0
   gh pr create --title "chore: add CHANGELOG.md for v0.x.0 release" --body "..."
   ```

5. **Wait for PR approval and merge**

6. Pull latest main:
   ```bash
   git checkout main
   git pull
   ```

**Lessons Learned (v0.1.0):**
- Attempt #1 failed: CHANGELOG.md was missing entirely
- GoReleaser config line 69 includes CHANGELOG.md in archives
- Always create CHANGELOG via PR BEFORE tagging

### Phase 2: Create Release

#### 1. Create and Push Tag

Tags trigger the GitHub Actions release workflow automatically.

```bash
# Ensure you're on latest main with CHANGELOG
git checkout main
git pull

# Create annotated tag
git tag -a v0.x.0 -m "Release v0.x.0"

# Push tag to trigger release workflow
git push origin v0.x.0
```

**Tag Format:**
- Use semantic versioning: `v{MAJOR}.{MINOR}.{PATCH}`
- Always prefix with `v`
- Use annotated tags (with `-a` flag)

#### 2. Monitor Release Workflow

The tag push triggers `.github/workflows/release.yml`:

```bash
# Watch workflow progress
gh run list --workflow=release.yml --limit 1
gh run watch <RUN_ID>
```

**Expected Steps:**
1. Checkout code
2. Setup Go
3. Run tests
4. Setup QEMU (for ARM64 emulation)
5. Setup Docker Buildx (for multi-arch builds)
6. Docker Login (ghcr.io)
7. Run GoReleaser:
   - Build binaries (12 total: 6 platforms × 2 tools)
   - Build Docker images (AMD64 + ARM64)
   - Create packages (.deb, .rpm, .apk)
   - Create archives (.tar.gz, .zip)
   - Generate changelog
   - Publish Docker images
   - Create draft GitHub release
   - Publish to Homebrew tap
   - Publish to Scoop bucket

**Workflow Duration:** ~2-3 minutes for successful build

#### 3. Verify Workflow Success

If workflow fails, see [Troubleshooting](#troubleshooting) section.

If successful:
```bash
# Verify draft release created
gh release list --limit 1

# Verify distribution channels updated
gh api repos/anomalousventures/homebrew-tap/commits --jq '.[0].commit.message'
gh api repos/anomalousventures/scoop-bucket/commits --jq '.[0].commit.message'
```

### Phase 3: Review and Publish

#### 1. Review Draft Release

```bash
# View release details
gh release view v0.x.0

# Check release notes format
gh release view v0.x.0 --json body --jq '.body' | head -100
```

**Verify:**
- ✅ Changelog appears once (not duplicated)
- ✅ Release notes include intro message
- ✅ Release notes include auto-generated changelog
- ✅ Install instructions are present in footer
- ✅ All assets uploaded (binaries, packages, checksums)
- ✅ Docker images tagged correctly

#### 2. Publish Release

If everything looks good:

```bash
gh release edit v0.x.0 --draft=false
```

Or publish via GitHub UI: https://github.com/anomalousventures/tracks/releases

#### 3. Verify Distribution Channels

After publishing, verify all distribution channels work:

**Homebrew:**
```bash
brew install anomalousventures/tap/tracks
tracks version
```

**Scoop:**
```bash
scoop bucket add anomalousventures https://github.com/anomalousventures/scoop-bucket
scoop install tracks
tracks version
```

**Docker:**
```bash
docker pull ghcr.io/anomalousventures/tracks:0.x.0
docker run ghcr.io/anomalousventures/tracks:0.x.0 version
```

**go install:**
```bash
go install github.com/anomalousventures/tracks/cmd/tracks@v0.x.0
tracks version
```

## Troubleshooting

### Common Failure Scenarios

#### Missing CHANGELOG.md

**Error:**
```
Error: open CHANGELOG.md: no such file or directory
```

**Cause:** CHANGELOG.md doesn't exist or wasn't committed before tag creation.

**Fix:**
1. Delete failed tag: `git tag -d v0.x.0 && git push origin :refs/tags/v0.x.0`
2. Delete draft release: `gh release delete v0.x.0 --yes`
3. Create CHANGELOG.md via PR (see Phase 1, Step 3)
4. Merge PR and pull latest main
5. Re-create and push tag

**Prevention:** Always create CHANGELOG before tagging (see checklist)

#### Docker Multi-Arch Build Failure

**Error:**
```
exec /bin/sh: exec format error
```

**Cause:** Missing QEMU/Buildx setup for ARM64 builds on AMD64 runners.

**Fix:** Verify `.github/workflows/release.yml` includes:
```yaml
- name: Set up QEMU
  uses: docker/setup-qemu-action@v3.6.0

- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3.11.1
```

**Occurred:** v0.1.0 attempt #2

#### Homebrew/Scoop Publishing Failure

**Error:**
```
403 Resource not accessible by integration
```

**Cause:** Using default `GITHUB_TOKEN` which can't write to separate repos.

**Fix:**
1. Verify `TRACKS_RELEASER_TOKEN` secret exists in repository settings
2. Verify `.goreleaser.yml` uses the token:
   ```yaml
   homebrew_casks:
     - repository:
         token: "{{ .Env.TRACKS_RELEASER_TOKEN }}"

   scoops:
     - repository:
         token: "{{ .Env.TRACKS_RELEASER_TOKEN }}"
   ```
3. Verify `.github/workflows/release.yml` passes the token:
   ```yaml
   env:
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
     TRACKS_RELEASER_TOKEN: ${{ secrets.TRACKS_RELEASER_TOKEN }}
   ```

**Occurred:** v0.1.0 attempt #3

#### Duplicate Changelog in Release Notes

**Issue:** Changelog appears twice in GitHub release page.

**Cause:** `.goreleaser.yml` header template includes `{{ .ReleaseNotes }}` which duplicates the auto-generated changelog.

**Fix:** Remove `{{ .ReleaseNotes }}` from header template:
```yaml
release:
  header: |
    ## Tracks {{ .Version }} ({{ .Date }})

    Welcome to this new release of Tracks!
```

GoReleaser automatically appends the changelog after the header.

**Occurred:** v0.1.0 attempt #4

### Recovery Procedures

#### Delete Failed Release and Tag

```bash
# Delete draft release
gh release delete v0.x.0 --yes

# Delete local tag
git tag -d v0.x.0

# Delete remote tag
git push origin :refs/tags/v0.x.0
```

#### Re-attempt Release

After fixing issues:

```bash
# Ensure on latest main with all fixes
git checkout main
git pull

# Verify tests and linting pass
make lint
make test

# Re-create and push tag
git tag -a v0.x.0 -m "Release v0.x.0"
git push origin v0.x.0

# Monitor new workflow
gh run watch <NEW_RUN_ID>
```

## Version Numbering

Tracks follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes, incompatible API changes
- **MINOR** (0.X.0): New features, backwards-compatible
- **PATCH** (0.0.X): Bug fixes, backwards-compatible

## Release Artifacts

Each release produces:

### Binaries
- `tracks` CLI tool (6 platforms: linux/darwin/windows × amd64/arm64)
- `tracks-mcp` MCP server (6 platforms)

### Docker Images
- `ghcr.io/anomalousventures/tracks:X.Y.Z` (version-specific)
- `ghcr.io/anomalousventures/tracks:latest` (always newest)
- Multi-arch manifests (AMD64 + ARM64)

### Packages
- `.deb` (Debian/Ubuntu)
- `.rpm` (RHEL/Fedora/CentOS)
- `.apk` (Alpine)
- `.tar.gz` (Linux/macOS archives)
- `.zip` (Windows archives)

### Distribution Channels
- **Homebrew tap:** anomalousventures/homebrew-tap
- **Scoop bucket:** anomalousventures/scoop-bucket
- **go install:** `go install github.com/anomalousventures/tracks/cmd/tracks@vX.Y.Z`
- **Docker:** `docker pull ghcr.io/anomalousventures/tracks:X.Y.Z`

## Future Improvements

Planned enhancements for the release process:

### v0.2.0 Roadmap
- [ ] Shell script installer (`curl | sh` pattern)
- [ ] Automated release notes with highlights extraction
- [ ] Pre-release smoke tests (install and run basic commands)
- [ ] Rollback automation for failed releases

### Documentation
- [ ] Video walkthrough of release process
- [ ] Release announcement templates (blog, social media)
- [ ] User migration guides for breaking changes

## References

- [GoReleaser Documentation](https://goreleaser.com)
- [Keep a Changelog](https://keepachangelog.com)
- [Semantic Versioning](https://semver.org)
- [Conventional Commits](https://www.conventionalcommits.org)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

## Lessons Learned (v0.1.0)

The v0.1.0 release took **5 attempts** over several hours. Here's what we learned:

1. **CHANGELOG before tag** - Always create and merge CHANGELOG.md before tagging
2. **Test Docker setup** - Verify QEMU/Buildx in workflow for multi-arch builds
3. **Use dedicated PAT** - Default GITHUB_TOKEN can't write to distribution repos
4. **Avoid template duplication** - Don't include `{{ .ReleaseNotes }}` in header
5. **Document everything** - Future releases should be quick with this guide

**Release Timeline:**
- Attempt #1: Missing CHANGELOG.md
- Attempt #2: Docker multi-arch build failure (missing QEMU)
- Attempt #3: Homebrew/Scoop 403 errors (wrong token)
- Attempt #4: Duplicate changelog in release notes
- Attempt #5: **SUCCESS!** ✅

With this documentation, future releases should succeed on the first attempt.
