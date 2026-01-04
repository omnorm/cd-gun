# CD-Gun: Architecture

## System Overview

**CD-Gun** is a universal, lightweight CD/GitOps service for deployment automation. Runs as a systemd service on Linux and monitors git repositories, executing user-defined scripts when changes are detected.

### Key Characteristics

- **Universality**: works with any git repositories (GitHub, GitLab, Gitea, etc.)
- **Independence**: no dependency on Kubernetes, Docker or specific platforms
- **Simplicity**: YAML configuration, easy to deploy
- **Flexibility**: user defines actions (shell scripts)
- **Lightweight**: ~6 MB binary, minimal dependencies
- **Security**: runs under unprivileged user, supports sudo

## System Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                  Linux Host (systemd enabled)                 │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │         cd-gun-agent (systemd Service)                  │ │
│  │                                                          │ │
│  │  ┌─────────────────────────────────────────────────────┐│ │
│  │  │  Config Manager                                      ││ │
│  │  │  - Parsing config.yaml                               ││ │
│  │  │  - Configuration validation                          ││ │
│  │  │  - Hot reload (SIGHUP)                               ││ │
│  │  └─────────────────────────────────────────────────────┘│ │
│  │                         │                               │ │
│  │  ┌──────────────────────▼──────────────────────────────┐│ │
│  │  │  Monitor (per repository)                           ││ │
│  │  │  - Periodic git change checking                     ││ │
│  │  │  - Comparing commit hashes                          ││ │
│  │  │  - Detecting changes in watch_paths                 ││ │
│  │  │  - Generating ChangeEvent                           ││ │
│  │  └──────────────────┬───────────────────────────────────┘│ │
│  │                     │                                   │ │
│  │  ┌──────────────────▼──────────────────────────────────┐│ │
│  │  │  Executor                                           ││ │
│  │  │  - Executing shell scripts                          ││ │
│  │  │  - Setting environment variables (CDGUN_*)          ││ │
│  │  │  - Handling timeouts and errors                     ││ │
│  │  │  - Saving results                                   ││ │
│  │  └──────────────────┬───────────────────────────────────┘│ │
│  │                     │                                   │ │
│  │  ┌──────────────────▼──────────────────────────────────┐│ │
│  │  │  State Store                                        ││ │
│  │  │  - Persistent storage (state.json)                  ││ │
│  │  │  - Tracking latest repository hashes                ││ │
│  │  │  - History of executed actions                      ││ │
│  │  └──────────────────────────────────────────────────────┘│ │
│  │                                                          │ │
│  │  ┌──────────────────────────────────────────────────────┐│ │
│  │  │  Logger                                             ││ │
│  │  │  - Structured logging                               ││ │
│  │  │  - Levels: DEBUG, INFO, WARN, ERROR                 ││ │
│  │  └──────────────────────────────────────────────────────┘│ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  Local storage                                          │ │
│  │  - /var/cache/cd-gun/repos/*      (git cache)           │ │
│  │  - /var/lib/cd-gun/state.json     (state)               │ │
│  │  - /var/log/cd-gun.log            (logs)                │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
              │                              │
         (git fetch)                   (script execution)
              │                              │
    ┌─────────▼──────────┐      ┌────────────▼────────────┐
    │  External Git       │      │  Deployment             │
    │  repositories       │      │  - Shell scripts        │
    │  (GitHub, GitLab,   │      │  - systemctl commands   │
    │   Gitea, etc.)      │      │  - Docker, K8s, etc.    │
    └────────────────────┘      │  - Webhook notifications │
                                 └────────────────────────┘
```

## Lifecycle

1. **Startup**: systemd starts cd-gun-agent with configuration
2. **Initialization**: Monitors are created for each repository
3. **Monitoring**: each Monitor periodically checks for changes
4. **Detection**: when changes in watch_paths are found, a ChangeEvent is generated
5. **Execution**: Executor runs the action script with environment variables
6. **Persistence**: State Store saves the results and new hash
7. **Signals**: SIGHUP reloads config, SIGUSR1 forces a check

## System Components

### 1. **Config Manager**
**Files**: `internal/config/config.go`, `internal/config/types.go`

Loads and validates YAML configuration, and provides access to parameters for other components.

**Config structure:**
```yaml
agent:
  name: "cd-gun-agent"
  log_level: "info"
  log_file: "/var/log/cd-gun.log"  # Optional: path to log file (if omitted, logs to stdout/journalctl)
  state_dir: "/var/lib/cd-gun"
  cache_dir: "/var/cache/cd-gun/repos"
  poll_interval: "5m"

# Include repositories from files/patterns (multiple sources supported)
include_repositories:
  - "/etc/cd-gun/repositories/*.yaml"       # Glob pattern
  - "/etc/cd-gun/production/"                # Directory (loads all .yaml files)
  - "/etc/cd-gun/special-config.yaml"        # Direct file path

# Repositories can also be defined inline
repositories:
  - name: "web-app"
    url: "https://github.com/myorg/app.git"
    branch: "main"
    auth:
      type: "ssh" | "https" | "none"
    watch_paths:
      - "src/"
      - "package.json"
      - "docker-compose.yaml"
    poll_interval: "5m"
    action:
      type: "shell"
      script: "/opt/cd-gun/scripts/deploy.sh"
      timeout: "10m"
      env:
        DEPLOY_ENV: "production"
```

### 2. **Repository Monitor**
**Files**: `internal/monitor/monitor.go`, `internal/monitor/git_helper.go`

Monitors each repository in a separate goroutine. Periodically runs `git fetch` and compares hashes of files in `watch_paths`.

**Main loop:**
```
┌─ poll_interval interval
│  ├─ git fetch origin <branch>
│  ├─ Get hash of <branch>
│  ├─ Compare with stored hash (state.json)
│  └─ If changed:
│     └─ Send ChangeEvent → Executor
└─ Repeat
```

**ChangeEvent structure:**
```go
type ChangeEvent struct {
    RepositoryName string    // Name from config
    Files []string           // Changed files
    OldHash string           // Previous commit
    NewHash string           // Current commit
    DetectedAt time.Time     // Detection time
}
```

### 3. **Action Executor**
**Files**: `internal/executor/executor.go`

Executes actions when receiving a ChangeEvent from Monitor.

**Supported action types:**
- `shell`: run bash script with environment variables and timeout
- `webhook`: HTTP POST request (basic support)

**Environment variables passed to the script:**
```bash
CDGUN_REPO_NAME       # Repository name
CDGUN_REPO_URL        # Repository URL
CDGUN_REPO_PATH       # Path to local cache
CDGUN_BRANCH          # Branch
CDGUN_CHANGED_FILES   # List of changed files (CSV)
CDGUN_OLD_HASH        # Old hash
CDGUN_NEW_HASH        # New hash
```

Plus any custom variables from `action.env` in the config.

### 4. **State Store**
**Files**: `internal/state/store.go`, `internal/state/types.go`

Persists state in JSON for recovery after restart.

**Example contents of state.json:**
```json
{
  "version": "1.0",
  "last_updated": "2025-12-26T10:30:00Z",
  "repositories": {
    "web-app": {
      "last_commit_hash": "abc123...",
      "last_check": "2025-12-26T10:29:00Z",
      "last_execution": {
        "status": "success",
        "duration": 120,
        "timestamp": "2025-12-26T10:20:00Z"
      }
    }
  }
}
```

### 5. **Logger**
**Files**: `internal/logger/logger.go`

Structured logging with levels DEBUG, INFO, WARN, ERROR.

Logs:
- Initialization and startup
- Check results (changes/no changes)
- Script execution (success/errors)
- Signals and lifecycle management

### 6. **Main Application**
**Files**: `internal/app/app.go`, `cmd/cd-gun-agent/main.go`

Coordinates all components:
- Creates Config Manager
- Initializes Monitor for each repository
- Manages goroutines (start/stop)
- Handles signals (SIGHUP, SIGUSR1, SIGTERM)
- Graceful shutdown

## Event Lifecycle

```
1. Initialization (startup)
   ├─ Read config
   ├─ Validate
   ├─ Initialize local git cache
   └─ Load last state

2. Monitoring (main loop)
   ├─ For each repository:
   │  ├─ Run git fetch
   │  ├─ Compare hashes of tracked files
   │  └─ If changed:
   │     └─ Generate ChangeEvent
   ├─ Wait for next interval
   └─ Handle signals (SIGHUP, SIGUSR1, SIGTERM)

3. Action execution
   ├─ Start Action Executor
   ├─ Run script (with environment variables)
   ├─ Wait for completion (with timeout)
   ├─ Write result to logs
   └─ Update state in state.json
```

## Control Signals

| Signal | Action |
|--------|--------|
| `SIGHUP` | Reload configuration, reinitialize monitors |
| `SIGUSR1` | Force immediate check of all repositories |
| `SIGTERM` | Graceful shutdown (finish current operations and exit) |

## Deployment

### Systemd Unit File

```ini
[Unit]
Description=CD-Gun - Universal CD/GitOps Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=cd-gun
Group=cd-gun
ExecStart=/usr/local/bin/cd-gun-agent --config /etc/cd-gun/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Control signals
ExecReload=/bin/kill -HUP $MAINPID

[Install]
WantedBy=multi-user.target
```

### Host file structure

```
/etc/cd-gun/
├── config.yaml           # Main configuration
└── auth/
    ├── ssh/
    │   ├── id_rsa        # SSH key for repositories
    │   └── known_hosts
    └── .netrc            # Credentials for HTTPS

/var/lib/cd-gun/
├── state.json            # Monitoring state
├── repos/                # Local git cache
│   ├── web-app/
│   ├── api-service/
│   └── ...
└── logs/                 # Execution history

/opt/cd-gun/
├── scripts/              # User deployment scripts
│   ├── deploy-web.sh
│   ├── deploy-api.sh
│   └── restart-service.sh
└── hooks/                # Custom event handlers
```

## Usage examples

Configuration examples are available in the [examples/](examples/) directory:
- [simple-deploy.yaml](examples/simple-deploy.yaml) - basic example
- [multi-repo.yaml](examples/multi-repo.yaml) - multiple repositories
- [advanced-config.yaml](examples/advanced-config.yaml) - with custom variables

Script examples in [examples/scripts/](examples/scripts/):
- [deploy-web.sh](examples/scripts/deploy-web.sh) - simple web application deployment
- [deploy-api.sh](examples/scripts/deploy-api.sh) - API deployment
- [deploy-api-advanced.sh](examples/scripts/deploy-api-advanced.sh) - advanced example

## Security

- **Isolation**: dedicated system user `cd-gun` with minimal privileges
- **SSH**: use of the user's local SSH key
- **Logging**: all actions are logged for audit
- **Timeout**: all actions have a timeout (protection from hangs)
- **Validation**: configuration is checked on startup
- **sudo support**: see [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md) for privileged operations
