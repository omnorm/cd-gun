# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-12-26

### Added

- **Support for splitting configuration into multiple files**
  - Added `repositories_include` field for loading repositories by glob pattern
  - Added `repositories_include_dir` field for loading all `.yaml` files from directory
  - Both approaches can be used simultaneously with inline repositories in main config
  - Complete documentation in [docs/CONFIGURATION_SPLIT.md](docs/CONFIGURATION_SPLIT.md)

- **Configuration Examples**
  - `examples/config-with-includes.yaml` — example main config with includes
  - `examples/repositories/` — examples of separate repository configs (frontend, api, config)

- **Documentation**
  - New section in [INDEX.md](INDEX.md) with information about config splitting
  - Link to `CONFIGURATION_SPLIT.md` in [README.md](README.md)

### Changed

- `Load()` function in `config.Manager` now processes includes after loading main config
- Repositories from includes are added to main list in order: main config → glob pattern → directory

### Technical

- Added methods in `internal/config/config.go`:
  - `loadRepositoriesFromGlob(pattern string)` — load by glob pattern
  - `loadRepositoriesFromDir(dirPath string)` — load from directory
  - `loadRepositoriesFromFile(filePath string)` — load from single file

- Added fields in `Config` structure (`internal/config/types.go`):
  - `RepositoriesInclude string` — glob pattern for files
  - `RepositoriesIncludeDir string` — path to directory

### Compatibility

- ✅ Full backward compatibility with existing configs
- ✅ Old format with `repositories` continues to work without changes
- ✅ New fields are optional and completely ignored if not specified

---

## [0.1.0] - 2025-12-25

### Initial Release

- Implemented core CD-Gun functionality:
  - Pull-based Git repository monitoring (periodic synchronization)
  - Shell script execution on detecting changes
  - Webhook actions support (basic)
  - State persistence in JSON
  - Systemd integration
  - Signal handling (SIGTERM, SIGHUP)

- Documentation:
  - [README.md](README.md) — user guide
  - [ARCH.md](ARCH.md) — system architecture
  - [PLAN.md](PLAN.md) — development roadmap
  - [INDEX.md](INDEX.md) — documentation navigation
  - [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) — environment variables
  - [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md) — sudo setup for actions

- Examples:
  - [examples/simple-deploy.yaml](examples/simple-deploy.yaml) — simple config example
  - [examples/multi-repo.yaml](examples/multi-repo.yaml) — multiple repositories
  - [examples/advanced-config.yaml](examples/advanced-config.yaml) — advanced configuration
  - Shell scripts for deployment in [examples/scripts/](examples/scripts/)

---

## Versioning

Versions in `X.Y.Z` format (Semantic Versioning):

- **X** — major version (breaking changes, incompatibility)
- **Y** — minor version (new features, backward compatible)
- **Z** — patch version (bug fixes, backward compatible)

**Current version:** [0.1.1]

