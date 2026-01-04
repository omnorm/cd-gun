# CD-Gun: Documentation Index

Quick search and navigation through project documentation.

## 📍 Main Documents

| Document | Contains |
|----------|----------|
| [README.md](README.md) | Project overview, quick start, variable table |
| [ARCH.md](ARCH.md) | Architecture, components, lifecycle |
| [PLAN.md](PLAN.md) | Development status, completed phases, plans |
| [STATUS.md](STATUS.md) | Current status, what's done, what's in development |

## 📚 Documentation in docs/

| Document | Contains |
|----------|----------|
| [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) | Complete reference for variables + script examples |
| [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md) | Sudo configuration for privileged operations |

## 🛠 Examples in examples/

| File | Purpose |
|------|-----------|
| [examples/simple-deploy.yaml](examples/simple-deploy.yaml) | Simple configuration example |
| [examples/multi-repo.yaml](examples/multi-repo.yaml) | Multiple repositories |
| [examples/advanced-config.yaml](examples/advanced-config.yaml) | With custom variables |
| [examples/scripts/deploy-web.sh](examples/scripts/deploy-web.sh) | Simple deployment script |
| [examples/scripts/deploy-api.sh](examples/scripts/deploy-api.sh) | Script for Docker/API |
| [examples/scripts/deploy-api-advanced.sh](examples/scripts/deploy-api-advanced.sh) | Advanced with notifications |

## 🎯 Quick Navigation by Topics

### Want to get started quickly
1. [README.md](README.md) → installation and quick start
2. [examples/simple-deploy.yaml](examples/simple-deploy.yaml) → config example
3. [examples/scripts/deploy-web.sh](examples/scripts/deploy-web.sh) → script example

### Need environment variables for scripts
→ [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) — complete reference with examples in different languages

### Want to understand architecture
→ [ARCH.md](ARCH.md) — components, lifecycle, signals

### Need config examples
→ [examples/](examples/) — 3 examples from simple to complex

### Need sudo setup
→ [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md) — how to give privileges to scripts

### Want to know project status
→ [STATUS.md](STATUS.md) or [PLAN.md](PLAN.md) — development phases, plans

### Need management commands
→ [README.md](README.md#control-signals) — signals, build, tests

## ✅ Complete File List

### Root Documents
- [README.md](README.md) — main documentation
- [ARCH.md](ARCH.md) — system architecture
- [PLAN.md](PLAN.md) — development plan
- [STATUS.md](STATUS.md) — development status
- [Makefile](Makefile) — build commands
- [go.mod](go.mod) — Go dependencies

### Documentation (docs/)
- [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) — environment variables for scripts
- [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md) — sudo configuration

### Examples (examples/)
- [examples/simple-deploy.yaml](examples/simple-deploy.yaml)
- [examples/multi-repo.yaml](examples/multi-repo.yaml)
- [examples/advanced-config.yaml](examples/advanced-config.yaml)
- [examples/scripts/deploy-web.sh](examples/scripts/deploy-web.sh)
- [examples/scripts/deploy-api.sh](examples/scripts/deploy-api.sh)
- [examples/scripts/deploy-api-advanced.sh](examples/scripts/deploy-api-advanced.sh)

### Source Code (internal/ and cmd/)
- [cmd/cd-gun-agent/main.go](cmd/cd-gun-agent/main.go) — entry point
- [internal/app/app.go](internal/app/app.go) — main application
- [internal/config/](internal/config/) — Config Manager
- [internal/executor/executor.go](internal/executor/executor.go) — Action Executor
- [internal/logger/logger.go](internal/logger/logger.go) — Logging
- [internal/monitor/](internal/monitor/) — Repository Monitor
- [internal/state/](internal/state/) — State Store

### Deployment (deployments/)
- [deployments/cd-gun.service](deployments/cd-gun.service) — systemd unit file
- [deployments/cd-gun.sudoers](deployments/cd-gun.sudoers) — sudoers configuration

---

**See also:** [PLAN.md](PLAN.md) for development plan and [STATUS.md](STATUS.md) for current status

## 🔧 Splitting configuration into multiple files

When working with a large number of repositories, it's convenient to split the configuration:
→ [docs/CONFIGURATION_SPLIT.md](docs/CONFIGURATION_SPLIT.md) — how to split config into multiple files

**Examples of split configuration:**
- [examples/config-with-includes.yaml](examples/config-with-includes.yaml) — main config with includes
- [examples/repositories/](examples/repositories/) — repository file examples