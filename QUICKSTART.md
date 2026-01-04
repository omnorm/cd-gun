# CD-Gun: Quick Start

## 🚀 In 5 minutes

### 1. Compile

```bash
cd cd-gun
make build
```

### 2. Create configuration

```bash
sudo mkdir -p /etc/cd-gun /var/lib/cd-gun
sudo cp examples/simple-deploy.yaml /etc/cd-gun/config.yaml
```

Edit `/etc/cd-gun/config.yaml`:
```yaml
agent:
  name: "my-agent"
repositories:
  - name: "my-repo"
    url: "https://github.com/YOUR_ORG/YOUR_REPO.git"
    branch: "main"
    watch_paths:
      - "src/"
    action:
      type: "shell"
      script: "/path/to/your/deploy.sh"
```

### 3. Install

```bash
sudo make install
```

### 3.5 (If scripts use sudo) Configure sudo

If your deployment script uses `sudo` commands (e.g., nginx restart):

```bash
sudo cp deployments/cd-gun.sudoers /etc/sudoers.d/cd-gun
sudo chmod 0440 /etc/sudoers.d/cd-gun
sudo visudo -c -f /etc/sudoers.d/cd-gun  # Check syntax
```

**More details:** [docs/SUDO_SETUP.md](docs/SUDO_SETUP.md)

### 4. Run

```bash
sudo systemctl start cd-gun
sudo systemctl enable cd-gun
```

### 5. Check logs

```bash
sudo journalctl -u cd-gun -f
```

## 📝 Write your own deployment script

Create file `/opt/cd-gun/scripts/my-deploy.sh`:

```bash
#!/bin/bash
set -e

echo "Deploying $CDGUN_REPO_NAME from $CDGUN_OLD_HASH to $CDGUN_NEW_HASH"

cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# Your deployment logic
npm ci && npm run build && npm run deploy

echo "Deployment complete!"
```

Make executable:
```bash
chmod +x /opt/cd-gun/scripts/my-deploy.sh
```

## 🔧 Available environment variables

Automatically available in your script:

```bash
$CDGUN_REPO_NAME       # Repository name
$CDGUN_REPO_PATH       # Local path
$CDGUN_BRANCH          # Branch
$CDGUN_CHANGED_FILES   # List of changed files
$CDGUN_NEW_HASH        # Current commit
$CDGUN_OLD_HASH        # Previous commit
```

Plus any custom variables from `action.env` in config.

**Full reference:** [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)

## 📚 Examples

- **Simple deployment:** [examples/simple-deploy.yaml](examples/simple-deploy.yaml)
- **Multiple repositories:** [examples/multi-repo.yaml](examples/multi-repo.yaml)
- **With custom variables:** [examples/advanced-config.yaml](examples/advanced-config.yaml)
- **Simple deployment script:** [examples/scripts/deploy-web.sh](examples/scripts/deploy-web.sh)
- **Advanced script:** [examples/scripts/deploy-api-advanced.sh](examples/scripts/deploy-api-advanced.sh)

## 🛠️ Service management

```bash
# Start
sudo systemctl start cd-gun

# Stop
sudo systemctl stop cd-gun

# Reload config (without restart)
sudo kill -HUP $(systemctl show -p MainPID cd-gun --value)

# Force check all repositories
sudo kill -USR1 $(systemctl show -p MainPID cd-gun --value)

# View logs
sudo journalctl -u cd-gun -f

# Status
sudo systemctl status cd-gun
```

## 📖 Full documentation

- [README.md](README.md) - Main documentation
- [ARCH.md](ARCH.md) - System architecture
- [PLAN.md](PLAN.md) - Development plan
- [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) - Environment variables reference

## ⚠️ Important

- Scripts run as user `cd-gun`
- Make sure user `cd-gun` has access to required resources
- Use SSH keys for git repository access
- Logging goes to systemd journal

## 🐛 Debugging

```bash
# Run in debug mode
./bin/cd-gun-agent -config /etc/cd-gun/config.yaml -log-level debug

# Check config
cat /etc/cd-gun/config.yaml

# Check state
cat /var/lib/cd-gun/state.json

# Check local repository cache
ls -la /var/lib/cd-gun/repos/
```

## 🆘 Issues?

1. Check config syntax: `cd-gun-agent -config /etc/cd-gun/config.yaml`
2. Check logs: `sudo journalctl -u cd-gun -n 50`
3. Check permissions on scripts and directories
4. Make sure user `cd-gun` exists: `id cd-gun`

## 📞 Support

- 📖 [Full documentation](docs/ENVIRONMENT_VARIABLES.md)
- 🎯 [Script examples](examples/scripts/)
- 🏗️ [Architecture](ARCH.md)
