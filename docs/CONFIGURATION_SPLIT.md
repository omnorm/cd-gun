# CD-Gun: Splitting Configuration into Multiple Files

When working with a large number of repositories, one configuration file can become inconvenient. CD-Gun supports flexible configuration splitting into multiple files via the `include_repositories` parameter.

## Main Approach: `include_repositories`

The `include_repositories` parameter accepts a list of paths and glob patterns for including repository configuration files:

**`/etc/cd-gun/config.yaml`:**
```yaml
agent:
  name: "cd-gun-agent"
  log_level: "info"
  log_file: "/var/log/cd-gun.log"
  state_dir: "/var/lib/cd-gun"
  cache_dir: "/var/cache/cd-gun/repos"
  poll_interval: "5m"

# List of paths/patterns for includes
# Supports:
#   - Glob patterns (/etc/cd-gun/*.yaml)
#   - Directories (/etc/cd-gun/repos/) - will load all .yaml files
#   - Direct file paths (/etc/cd-gun/special.yaml)
include_repositories:
  - "/etc/cd-gun/repositories/*.yaml"       # Glob pattern
  - "/etc/cd-gun/production/"                # Directory
  - "/etc/cd-gun/special-config.yaml"        # Specific file
  - "/etc/cd-gun/projects/*/*.yaml"          # Nested globs

# Repositories can also be specified here (optional)
repositories: []
```

## Usage Examples

### Example 1: One Directory with Glob Pattern

```yaml
include_repositories:
  - "/etc/cd-gun/repositories/*.yaml"

repositories: []
```

**File structure:**
```
/etc/cd-gun/
├── config.yaml
└── repositories/
    ├── frontend.yaml
    ├── api.yaml
    └── workers.yaml
```

### Example 2: Multiple Directories by Projects

```yaml
include_repositories:
  - "/etc/cd-gun/project1/*.yaml"
  - "/etc/cd-gun/project2/*.yaml"
  - "/etc/cd-gun/production/*.yaml"

repositories: []
```

**File structure:**
```
/etc/cd-gun/
├── config.yaml
├── project1/
│   ├── frontend.yaml
│   └── api.yaml
├── project2/
│   ├── worker.yaml
│   └── scheduler.yaml
└── production/
    ├── critical-app.yaml
    └── monitoring.yaml
```

### Example 3: Hybrid Approach - Includes + Local Repositories

```yaml
include_repositories:
  - "/etc/cd-gun/repositories/*.yaml"
  - "/etc/cd-gun/staging/*.yaml"

repositories:
  - name: "local-debug-app"
    url: "https://github.com/myorg/debug.git"
    branch: "develop"
    watch_paths:
      - "src/"
    action:
      type: "shell"
      script: "/opt/cd-gun/scripts/deploy-debug.sh"
```

### Example 4: Nested Patterns

```yaml
include_repositories:
  - "/etc/cd-gun/services/*/*.yaml"         # Will load all .yaml in subfolders
  - "/etc/cd-gun/special-deploys/**/*.yaml" # Recursive search (if supported)
```

## Include File Formats

### Format 1: Repository Array (recommended)

`/etc/cd-gun/repositories/frontend.yaml`:
```yaml
- name: "frontend"
  url: "https://github.com/myorg/frontend.git"
  branch: "main"
  watch_paths:
    - "dist/"
    - "src/"
  action:
    type: "shell"
    script: "/opt/cd-gun/scripts/deploy-frontend.sh"
    timeout: "15m"

- name: "frontend-staging"
  url: "https://github.com/myorg/frontend.git"
  branch: "staging"
  watch_paths:
    - "dist/"
  action:
    type: "shell"
    script: "/opt/cd-gun/scripts/deploy-frontend-staging.sh"
```

### Format 2: Single Repository (per-file)

`/etc/cd-gun/repositories/api.yaml`:
```yaml
name: "api-service"
url: "https://github.com/myorg/api.git"
branch: "main"
watch_paths:
  - "src/"
  - "docker-compose.yml"
action:
  type: "shell"
  script: "/opt/cd-gun/scripts/deploy-api.sh"
  timeout: "20m"
  env:
    DEPLOY_ENV: "production"
```

Both formats are supported and can be mixed.

## Recommended Directory Structures

### For Organization by Projects

```
/etc/cd-gun/
├── config.yaml
├── project1/
│   ├── frontend.yaml
│   ├── api.yaml
│   └── workers.yaml
├── project2/
│   ├── services.yaml
│   └── jobs.yaml
└── shared/
    ├── monitoring.yaml
    └── shared-libs.yaml
```

**config.yaml:**
```yaml
agent:
  name: "cd-gun-agent"
  ...

include_repositories:
  - "/etc/cd-gun/project1/*.yaml"
  - "/etc/cd-gun/project2/*.yaml"
  - "/etc/cd-gun/shared/*.yaml"
```

### For Organization by Environments
├── config.yaml                    # Global agent parameters
├── repositories/                  # Repository configurations
│   ├── frontend.yaml
│   ├── api.yaml
│   ├── config-service.yaml
│   ├── database.yaml
│   └── ...
└── auth/
    ├── ssh/
    │   ├── id_rsa
    │   └── known_hosts
    └── .netrc
```

### For Organization by Environments

```
/etc/cd-gun/
├── config.yaml
├── repositories/
│   ├── production/
│   │   ├── frontend.yaml
│   │   ├── api.yaml
│   │   └── ...
│   ├── staging/
│   │   ├── frontend.yaml
│   │   ├── api.yaml
│   │   └── ...
│   └── development/
│       ├── ...
└── auth/
```

In this case use a glob pattern:
```yaml
repositories_include: "/etc/cd-gun/repositories/production/*.yaml"
```

### For Organization by Services

```
/etc/cd-gun/
├── config.yaml
├── repositories/
│   ├── backend/
│   │   ├── api.yaml
│   │   ├── auth-service.yaml
│   │   └── database.yaml
│   ├── frontend/
│   │   ├── web.yaml
│   │   ├── admin.yaml
│   │   └── cdn.yaml
│   └── infrastructure/
│       ├── config-service.yaml
│       └── monitoring.yaml
└── auth/
```

Use multiple glob patterns in one file:
```yaml
# You can create multiple configs with different patterns,
# or use a broader pattern:
repositories_include: "/etc/cd-gun/repositories/**/*.yaml"
```

## Repository File Format

Repository files contain an array of repositories in YAML format.

### Repository Array

**`repos.yaml`:**
```yaml
- name: "frontend"
  url: "https://github.com/myorg/frontend.git"
  branch: "main"
  watch_paths:
    - "dist/"
  action:
    type: "shell"
    script: "/opt/cd-gun/scripts/deploy.sh"

- name: "api"
  url: "https://github.com/myorg/api.git"
  branch: "main"
  watch_paths:
    - "src/"
  action:
    type: "shell"
    script: "/opt/cd-gun/scripts/deploy.sh"
```

### Single Repository (YAML object)

If a file contains a single repository, it will still be loaded correctly:

**`frontend.yaml`:**
```yaml
name: "frontend"
url: "https://github.com/myorg/frontend.git"
branch: "main"
watch_paths:
  - "dist/"
action:
  type: "shell"
  script: "/opt/cd-gun/scripts/deploy.sh"
```

## Command Examples

### Installation with Split Configuration

```bash
# 1. Create directory for repositories
mkdir -p /etc/cd-gun/repositories

# 2. Copy main config
sudo cp examples/config-with-includes.yaml /etc/cd-gun/config.yaml

# 3. Copy repository examples
sudo cp examples/repositories/*.yaml /etc/cd-gun/repositories/

# 4. Edit configs for your needs
sudo nano /etc/cd-gun/repositories/frontend.yaml
sudo nano /etc/cd-gun/repositories/api.yaml

# 5. Reload config
systemctl reload cd-gun
```

### Adding a New Repository

```bash
# Create new repository file
cat > /etc/cd-gun/repositories/new-service.yaml << 'EOF'
- name: "new-service"
  url: "https://github.com/myorg/new-service.git"
  branch: "main"
  watch_paths:
    - "src/"
  action:
    type: "shell"
    script: "/opt/cd-gun/scripts/deploy.sh"
EOF

# Reload config
systemctl reload cd-gun

# Check logs
journalctl -u cd-gun -f
```

### Testing Config Before Applying

```bash
# Run agent locally with new config
./bin/cd-gun-agent -config /etc/cd-gun/config.yaml -log-level debug
```

## Repository Loading Order

1. First, repositories from `repositories` are loaded (if any)
2. Then repositories from `repositories_include` are added (if specified)
3. Then repositories from `repositories_include_dir` are added (if specified)

This allows using all three approaches simultaneously, if needed.

## Config Reload Effect (SIGHUP)

When reloading config (SIGHUP):
- ALL repositories are reloaded (main file and includes)
- Already running checks are completed
- New monitors are created for new repositories
- Old monitors are stopped

```bash
# Reload config
kill -HUP $(pgrep cd-gun-agent)

# Or via systemctl
systemctl reload cd-gun
```

## Best Practices

1. **One repository — one file**: Each file should contain configuration for one or several related repositories
2. **Clear file names**: `frontend.yaml`, `api.yaml`, `database.yaml`, not `repo1.yaml`, `repo2.yaml`
3. **Organize by structure**: Use subfolders for organization (by environments, by services, etc.)
4. **Comments in files**: Add comments to explain configuration
5. **Version control**: Use git for config versioning (the `/etc/cd-gun` folder itself can be a git repository)

## Possible Errors

### File not loading

```
Error: failed to load repositories from glob pattern: no files match pattern
```

Solution: Check that the pattern is correct and files exist

```bash
# Check which files match the pattern
ls /etc/cd-gun/repositories/*.yaml
```

### Syntax error in repository file

```
Error: failed to parse repositories: ...
```

Solution: Check YAML syntax

```bash
# Check syntax (if yamllint is installed)
yamllint /etc/cd-gun/repositories/frontend.yaml
```

### Duplicate repository names

```
Error: config validation failed: ...
```

Solution: Check that different files don't have repositories with the same names

```bash
# Find duplicate names
grep -r "name:" /etc/cd-gun/repositories/ | sort | uniq -d
```
