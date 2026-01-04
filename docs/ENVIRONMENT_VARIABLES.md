# CD-Gun: Environment Variables for Deployment Scripts

When CD-Gun executes a shell script upon detecting changes in a repository, it automatically sets environment variables that contain useful information about the event. This allows scripts to get context about the changes and use them in deployment logic.

## Built-in Environment Variables

These variables are set automatically for each script execution:

### Basic Variables

| Variable | Type | Description |
|------------|-----|---------|
| `CDGUN_REPO_NAME` | string | Repository name from configuration |
| `CDGUN_REPO_URL` | string | Repository URL (as specified in config.yaml) |
| `CDGUN_REPO_PATH` | string | Local path to cached repository on host |
| `CDGUN_BRANCH` | string | Branch that CD-Gun is monitoring |

### Change Information

| Variable | Type | Description |
|------------|-----|---------|
| `CDGUN_CHANGED_FILES` | string (CSV) | List of changed files, comma-separated |
| `CDGUN_OLD_HASH` | string | Hash of previous commit (empty on first run) |
| `CDGUN_NEW_HASH` | string | Hash of current commit |

### Custom Variables

You can add your own variables in the configuration in the `action.env` section:

```yaml
repositories:
  - name: "my-app"
    url: "https://github.com/..."
    branch: "main"
    watch_paths:
      - "src/"
    action:
      type: "shell"
      script: "/opt/cd-gun/scripts/deploy.sh"
      env:
        DEPLOY_ENV: "production"
        DOCKER_REGISTRY: "docker.mycompany.com"
        SLACK_WEBHOOK: "https://hooks.slack.com/..."
```

In this example, the script will receive additional variables:
- `DEPLOY_ENV=production`
- `DOCKER_REGISTRY=docker.mycompany.com`
- `SLACK_WEBHOOK=https://hooks.slack.com/...`

## Usage Examples

### Example 1: Simple Web Application Deployment

```bash
#!/bin/bash
set -e

echo "Repository: $CDGUN_REPO_NAME"
echo "Branch: $CDGUN_BRANCH"
echo "Changed files: $CDGUN_CHANGED_FILES"

# Navigate to repository and update it
cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# Install dependencies and build
npm ci
npm run build

# Deploy
cp -r dist/* /var/www/myapp/
systemctl reload nginx

# Log the change
echo "Deployed: $CDGUN_REPO_NAME from $CDGUN_OLD_HASH to $CDGUN_NEW_HASH"
```

### Example 2: Deployment with Docker

```bash
#!/bin/bash
set -e

# Get information about values
REGISTRY="${DOCKER_REGISTRY:-docker.io}"
VERSION="${CDGUN_NEW_HASH:0:8}"  # First 8 characters of hash

echo "Building image: $REGISTRY/myapp:$VERSION"

cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# Check which files changed
echo "Changed files: $CDGUN_CHANGED_FILES"

# If Dockerfile or dependencies changed - rebuild image
if echo "$CDGUN_CHANGED_FILES" | grep -E "(Dockerfile|package\.json|requirements\.txt)"; then
    docker build -t "$REGISTRY/myapp:$VERSION" .
    docker push "$REGISTRY/myapp:$VERSION"
    
    # Update deployment
    kubectl set image deployment/myapp myapp="$REGISTRY/myapp:$VERSION"
fi

echo "Deployment complete"
```

### Example 3: Deployment with Notifications

```bash
#!/bin/bash

REPO="$CDGUN_REPO_NAME"
OLD_HASH="$CDGUN_OLD_HASH"
NEW_HASH="$CDGUN_NEW_HASH"
BRANCH="$CDGUN_BRANCH"
SLACK_WEBHOOK="$SLACK_WEBHOOK"

# Function to send to Slack
notify_slack() {
    local message="$1"
    local status="$2"
    
    curl -X POST "$SLACK_WEBHOOK" \
        -H 'Content-Type: application/json' \
        -d @- <<EOF
{
    "text": "$message",
    "attachments": [{
        "color": "$status",
        "fields": [
            {"title": "Repository", "value": "$REPO"},
            {"title": "Branch", "value": "$BRANCH"},
            {"title": "From", "value": "${OLD_HASH:0:8}"},
            {"title": "To", "value": "${NEW_HASH:0:8}"}
        ]
    }]
}
EOF
}

# Start deployment
notify_slack "🚀 Deployment started for $REPO" "warning"

if deploy_application; then
    notify_slack "✅ Deployment successful for $REPO" "good"
else
    notify_slack "❌ Deployment failed for $REPO" "danger"
    exit 1
fi
```

### Example 4: Conditional Execution Based on Changed Files

```bash
#!/bin/bash
set -e

CHANGED="$CDGUN_CHANGED_FILES"

cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# If configuration files changed
if echo "$CHANGED" | grep -q "config/"; then
    echo "Configuration changed, reloading..."
    systemctl reload myapp
fi

# If frontend files changed
if echo "$CHANGED" | grep -q "static/"; then
    echo "Static files changed, rebuilding..."
    npm run build
    systemctl reload nginx
fi

# If database migration files changed
if echo "$CHANGED" | grep -q "migrations/"; then
    echo "Migrations found, running database updates..."
    python manage.py migrate
fi

echo "Update completed"
```

## Accessing Environment Variables

### Bash/Shell

```bash
#!/bin/bash
echo "Repository: $CDGUN_REPO_NAME"
echo "Path: $CDGUN_REPO_PATH"
```

### Python

```python
#!/usr/bin/env python3
import os

repo_name = os.environ.get('CDGUN_REPO_NAME')
repo_path = os.environ.get('CDGUN_REPO_PATH')
changed_files = os.environ.get('CDGUN_CHANGED_FILES', '').split(',')

print(f"Repository: {repo_name}")
print(f"Path: {repo_path}")
print(f"Changed files: {changed_files}")
```

### Go

```go
package main

import (
    "os"
    "fmt"
)

func main() {
    repoName := os.Getenv("CDGUN_REPO_NAME")
    repoPath := os.Getenv("CDGUN_REPO_PATH")
    changedFiles := os.Getenv("CDGUN_CHANGED_FILES")
    
    fmt.Printf("Repository: %s\n", repoName)
    fmt.Printf("Path: %s\n", repoPath)
    fmt.Printf("Changed files: %s\n", changedFiles)
}
```

## Security Best Practices

### 1. Always escape variables in shell scripts

```bash
# ✅ Good - variables in quotes
cd "$CDGUN_REPO_PATH"
git reset --hard "$CDGUN_NEW_HASH"

# ❌ Bad - variables without quotes (injection risk)
cd $CDGUN_REPO_PATH
git reset --hard $CDGUN_NEW_HASH
```

### 2. Use `set -e` to stop on errors

```bash
#!/bin/bash
set -e  # Stop script on first error

# Now if any command fails with an error, the script will stop
npm ci
npm run build
npm run test
```

### 3. Log actions

```bash
#!/bin/bash
set -e

echo "[$(date)] Starting deployment of $CDGUN_REPO_NAME"
echo "[$(date)] Branch: $CDGUN_BRANCH"
echo "[$(date)] New hash: $CDGUN_NEW_HASH"

# ... rest of code

echo "[$(date)] Deployment completed successfully"
```

### 4. Check variables before using

```bash
#!/bin/bash
set -e

if [ -z "$CDGUN_REPO_PATH" ]; then
    echo "ERROR: CDGUN_REPO_PATH is not set"
    exit 1
fi

if [ -z "$CDGUN_NEW_HASH" ]; then
    echo "ERROR: CDGUN_NEW_HASH is not set"
    exit 1
fi

# Continue script...
```

## Common Usage Scenarios

### Deploy only when specific files change

```bash
#!/bin/bash
set -e

# Check if Dockerfile or source code changed
if echo "$CDGUN_CHANGED_FILES" | grep -E "^(Dockerfile|src/)" > /dev/null; then
    echo "Code or Dockerfile changed, rebuilding..."
    docker build -t myapp:latest .
    docker push myapp:latest
    systemctl restart myapp
else
    echo "No relevant changes, skipping deployment"
fi
```

### Send error notifications

```bash
#!/bin/bash

set +e  # Don't exit on error immediately

if ! deploy_application; then
    # Send error notification
    curl -X POST "$SLACK_WEBHOOK" \
        -H 'Content-Type: application/json' \
        -d "{\"text\": \"❌ Deployment failed for $CDGUN_REPO_NAME\"}"
    exit 1
fi

set -e
```

### Track deployment versions

```bash
#!/bin/bash
set -e

DEPLOY_LOG="/var/log/deployments.log"

echo "$(date '+%Y-%m-%d %H:%M:%S') | $CDGUN_REPO_NAME | $CDGUN_BRANCH | $CDGUN_OLD_HASH -> $CDGUN_NEW_HASH" >> "$DEPLOY_LOG"

# Perform deployment...
```

## Debugging

If the script receives incorrect variable values, you can temporarily add debugging:

```bash
#!/bin/bash

# Output all CDGUN_* variables
env | grep "^CDGUN_" | sort

# Then perform deployment
```

Then check the logs:

```bash
sudo journalctl -u cd-gun -f
```

## Limitations and Features

- **Variables available only in scripts** - they are not passed to webhooks (planned in future version)
- **CDGUN_CHANGED_FILES** - this is a CSV list of files, comma-separated
- **CDGUN_OLD_HASH** - empty on first script run for repository
- **Custom variables** - override built-in ones if names match
- **All variables are strings** - if you need a number, convert in script

## See Also

- [README.md](../README.md) - main documentation
- [ARCH.md](../ARCH.md) - system architecture
- [examples/](../examples/) - configuration and script examples
