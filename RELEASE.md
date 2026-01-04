# Release Process

## Version Numbering

CD-Gun uses [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., `0.1.1`)
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

## Creating a Release

### 1. Prepare the Release

```bash
# Update version in Makefile
vim Makefile

# Update CHANGELOG.md with all changes
vim CHANGELOG.md

# Commit changes
git add Makefile CHANGELOG.md
git commit -m "chore: prepare release v0.2.0"

# Push to main
git push origin main
```

### 2. Tag the Release

```bash
# Create a signed tag (recommended)
git tag -s v0.2.0 -m "Release v0.2.0"

# Or unsigned tag
git tag v0.2.0 -m "Release v0.2.0"

# Push the tag
git push origin v0.2.0
```

### 3. GitHub Actions

The `.github/workflows/release.yml` will automatically:
- Build binaries for Linux (amd64, arm64) and macOS (amd64, arm64)
- Generate checksums
- Create a GitHub Release with binaries attached

### 4. Post-Release

- Monitor the release on [GitHub Releases](https://github.com/omnorm/cd-gun/releases)
- Announce the release on your channels (blog, social media, etc.)
- Close associated GitHub issues

## Changelog Format

Use these prefixes in CHANGELOG.md:
- `Added` — New features
- `Changed` — Changes to existing functionality
- `Deprecated` — Soon-to-be removed features
- `Removed` — Removed features
- `Fixed` — Bug fixes
- `Security` — Security fixes

Example:
```markdown
## [0.2.0] - 2025-01-15

### Added
- Support for webhook-based triggers
- Configuration validation command

### Fixed
- Memory leak in repository monitor
- Incorrect handling of symlinks
```

## Pre-Release Checks

Before tagging, ensure:
- [ ] All tests pass: `make test`
- [ ] Code is formatted: `make fmt`
- [ ] No linter warnings: `make lint`
- [ ] CHANGELOG.md is updated
- [ ] README examples work correctly
- [ ] Documentation is up to date
