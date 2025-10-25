# Release Checklist

Quick reference for creating a new Tracks release. For detailed instructions, see [docs/RELEASE_PROCESS.md](../docs/RELEASE_PROCESS.md).

## Pre-Release

- [ ] On latest `main` branch
- [ ] Working tree clean (`git status`)
- [ ] All tests pass (`make test`)
- [ ] All linters pass (`make lint`)

## CHANGELOG

- [ ] Generate changelog: `go tool git-chglog -o CHANGELOG.md`
- [ ] Add release summary with highlights to CHANGELOG.md
- [ ] Fix markdown linting: `make lint-md`
- [ ] Create PR for CHANGELOG
- [ ] **Merge CHANGELOG PR before tagging**
- [ ] Pull latest main: `git checkout main && git pull`

## Create Release

- [ ] Create tag: `git tag -a v0.x.0 -m "Release v0.x.0"`
- [ ] Push tag: `git push origin v0.x.0`
- [ ] Monitor workflow: `gh run watch <RUN_ID>`
- [ ] Verify workflow success (no red ✗)

## Verify Draft Release

- [ ] Draft release created: `gh release list --limit 1`
- [ ] Changelog appears once (not duplicated)
- [ ] All assets uploaded (binaries, packages, archives, checksums)
- [ ] Homebrew tap updated: `gh api repos/anomalousventures/homebrew-tap/commits --jq '.[0].commit.message'`
- [ ] Scoop bucket updated: `gh api repos/anomalousventures/scoop-bucket/commits --jq '.[0].commit.message'`

## Publish

- [ ] Review release notes: `gh release view v0.x.0`
- [ ] Publish: `gh release edit v0.x.0 --draft=false`
- [ ] Verify on GitHub: https://github.com/anomalousventures/tracks/releases

## Post-Release Verification

- [ ] Homebrew install works: `brew install anomalousventures/tap/tracks`
- [ ] Scoop install works: `scoop install tracks` (after adding bucket)
- [ ] Docker pull works: `docker pull ghcr.io/anomalousventures/tracks:0.x.0`
- [ ] go install works: `go install github.com/anomalousventures/tracks/cmd/tracks@v0.x.0`

## If Release Fails

See [docs/RELEASE_PROCESS.md#troubleshooting](../docs/RELEASE_PROCESS.md#troubleshooting) for recovery procedures.

**Quick recovery:**
```bash
# Delete failed release
gh release delete v0.x.0 --yes
git tag -d v0.x.0
git push origin :refs/tags/v0.x.0

# Fix issues, then retry
git checkout main && git pull
git tag -a v0.x.0 -m "Release v0.x.0"
git push origin v0.x.0
```

---

**First time releasing?** Read the comprehensive [docs/RELEASE_PROCESS.md](../docs/RELEASE_PROCESS.md) guide.
