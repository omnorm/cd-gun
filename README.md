# CD-Gun: Universal CD/GitOps Agent

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue)](https://golang.org/doc/devel/release)

A lightweight, universal CD/GitOps service for deploying anything to anywhere. CD-Gun is a systemd-based agent that monitors git repositories and executes actions when changes are detected.

**Status**: 🚀 v0.1.1 (Early Development)

## About

This project was developed with the assistance of **GitHub Copilot**. While the concept, architecture, design decisions, and comprehensive testing were provided by the project author, the implementation code was generated and refined using AI assistance. This approach demonstrates modern collaborative development practices combining human creativity with AI-powered code generation.

## Features

- 🌍 **Universal**: Works with any git-compatible repository (GitHub, GitLab, Gitea, self-hosted, etc.)
- 🔧 **Flexible**: Define custom actions in shell scripts
- 🪶 **Lightweight**: ~6MB binary, minimal resource usage
- 🔌 **Independent**: No dependencies on Kubernetes, Docker, or specific platforms
- 📋 **Simple**: Easy to configure with YAML
- 🛡️ **Secure**: Runs as unprivileged systemd service

## Quick Start

### Installation

```bash
git clone https://github.com/omnorm/cd-gun.git
cd cd-gun

make build
sudo make install

# Start the service
systemctl start cd-gun
```

### Basic Usage

1. Create a config file `/etc/cd-gun/config.yaml`:

```yaml
agent:
  name: "cd-gun-agent"
  log_level: "info"
  log_file: "/var/log/cd-gun.log"  # Optional: omit to log to stdout/journalctl
  state_dir: "/var/lib/cd-gun"
  cache_dir: "/var/cache/cd-gun/repos"
  poll_interval: "5m"

repositories:
  - name: "my-app"
    url: "https://github.com/myorg/app.git"
    branch: "main"
    watch_paths:
      - "src/"
      - "package.json"
    poll_interval: "5m"
    action:
      type: "shell"
      script: "/opt/cd-gun/scripts/deploy.sh"
      timeout: "10m"
```

2. Create a deployment script `/opt/cd-gun/scripts/deploy.sh`:

```bash
#!/bin/bash
set -e

echo "Deploying $CDGUN_REPO_NAME from $CDGUN_OLD_HASH to $CDGUN_NEW_HASH"

cd "$CDGUN_REPO_PATH"
npm ci
npm run build
rsync -av dist/ /var/www/app/
systemctl reload nginx

echo "Deployment successful!"
```

3. Check status:

```bash
systemctl status cd-gun
journalctl -u cd-gun -f
```

## Environment Variables for Scripts

When executing your deployment scripts, CD-Gun provides these variables:

| Variable | Description |
|----------|-------------|
| `CDGUN_REPO_NAME` | Repository name from config |
| `CDGUN_REPO_URL` | Repository URL |
| `CDGUN_REPO_PATH` | Local cache path |
| `CDGUN_BRANCH` | Branch being monitored |
| `CDGUN_CHANGED_FILES` | Changed files (comma-separated) |
| `CDGUN_OLD_HASH` | Previous commit hash |
| `CDGUN_NEW_HASH` | Current commit hash |

**Full reference:** [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)

## Control Signals

```bash
# Reload configuration
kill -HUP $(pgrep cd-gun-agent)

# Force check all repositories
kill -USR1 $(pgrep cd-gun-agent)

# Graceful shutdown
systemctl stop cd-gun
```

## Build & Development

```bash
# Build
make build

# Run locally for testing
./bin/cd-gun-agent -config examples/simple-deploy.yaml -log-level debug

# Run tests
make test

# Clean
make clean
```

## Project Structure

```
cmd/cd-gun-agent/      # Application entry point
internal/
├── app/                # Main application & event loop
├── config/             # Configuration management
├── executor/           # Action execution
├── monitor/            # Repository monitoring
├── state/              # State management
└── logger/             # Logging
examples/               # Example configurations
deployments/            # systemd service file
docs/                   # Documentation
```

## Documentation

- **[ARCH.md](ARCH.md)** — Architecture and design
- **[PLAN.md](PLAN.md)** — Implementation status
- **[STATUS.md](STATUS.md)** — Current development status
- **[CONTRIBUTING.md](CONTRIBUTING.md)** — How to contribute
- **[SECURITY.md](SECURITY.md)** — Security policy and best practices
- **[RELEASE.md](RELEASE.md)** — Release process
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** — Community guidelines
- **[docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)** — Environment variables guide
- **[docs/CONFIGURATION_SPLIT.md](docs/CONFIGURATION_SPLIT.md)** — Splitting config into multiple files
- **[docs/SUDO_SETUP.md](docs/SUDO_SETUP.md)** — Sudo configuration for privileged operations
- **[examples/](examples/)** — Configuration and script examples

## Common Use Cases

- **Web application deployment** — Auto-deploy on git push
- **Configuration management** — Auto-update service configs
- **Multi-repo coordination** — Monitor and sync multiple repositories
- **Custom deployment pipelines** — Run any shell script on changes

## Compared to Alternatives

If you're wondering how CD-Gun compares to other tools:

| Tool | Best For | vs CD-Gun |
|------|----------|----------|
| **Argo CD** / **Flux** | Kubernetes clusters | CD-Gun is simpler, works on plain servers without K8s |
| **Jenkins** / **GitHub Actions** | Full CI/CD pipelines | CD-Gun is lighter, uses pull model instead of webhooks |
| **Ansible pull** | Config management | Similar pull-based approach, but CD-Gun is script-agnostic |
| **Webhook services** | Git → script execution | CD-Gun polls instead of requiring open webhooks; works with private networks |

**CD-Gun is best for**: Single/multiple servers, simple deployments, no K8s, minimal dependencies, custom bash-based workflows.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Permishen Denaev
