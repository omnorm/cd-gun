# Sudo Setup for CD-Gun

## Problem

When running deployment scripts that use `sudo` commands, an error occurs:

```
sudo: effective uid is not 0, is sudo installed setuid root?
```

This happens because the cd-gun service runs under a regular user (`cd-gun`), not root, and it needs elevated privileges for some operations.

## Solution

Use the `sudoers` file to grant the `cd-gun` user the ability to run necessary commands **without entering a password**.

### Installation

1. **Install the sudoers configuration:**

```bash
sudo cp deployments/cd-gun.sudoers /etc/sudoers.d/cd-gun
sudo chmod 0440 /etc/sudoers.d/cd-gun
```

2. **Ensure the file is syntactically correct:**

```bash
sudo visudo -c -f /etc/sudoers.d/cd-gun
```

### Configuration Contents

The `/etc/sudoers.d/cd-gun` file contains rules allowing the `cd-gun` user to run:

- `cp` - copying files (for web applications)
- `systemctl` - service management (nginx, apache2, etc.)
- `docker` and `docker-compose` - container management
- Other file operations (`chmod`, `chown`, `rm`, `mkdir`)

All commands are executed **without a password** thanks to the `NOPASSWD` flag.

### Example Usage in Scripts

In scripts, you can freely use `sudo`:

```bash
# Copying files
sudo cp -r dist/* /var/www/myapp/

# Reloading nginx
sudo systemctl reload nginx

# Docker operations
sudo docker-compose down
sudo docker-compose up -d
```

### Security

**⚠️ Important:** The rules provided in `cd-gun.sudoers` are quite permissive (broad).

For production environments, it is recommended to:

1. **Restrict commands to specific paths:**

```bash
# Instead of allowing all paths:
cd-gun ALL=(ALL) NOPASSWD: /bin/cp -r /var/lib/cd-gun/repos/web-app/dist/* /var/www/myapp/

# Instead of allowing all systemctl services:
cd-gun ALL=(ALL) NOPASSWD: /usr/bin/systemctl reload nginx
cd-gun ALL=(ALL) NOPASSWD: /usr/bin/systemctl reload apache2
```

2. **Restrict docker commands:**

```bash
# Instead of allowing all docker commands:
cd-gun ALL=(ALL) NOPASSWD: /usr/bin/docker-compose -f /var/lib/cd-gun/repos/*/docker-compose.yml up -d
```

3. **Use audit logging:**

```bash
sudo auditctl -w /etc/sudoers.d/ -p wa -k sudoers_changes
```

### Verification

Make sure everything works:

```bash
sudo -u cd-gun sudo systemctl status nginx
sudo -u cd-gun sudo cp /tmp/test /tmp/test2
sudo -u cd-gun sudo docker ps
```

All commands should execute without prompting for a password.

### Debugging

If commands don't work:

1. **Check absolute paths:**
   - Use `which cp` and `which systemctl` to get full paths
   - Make sure absolute paths are used in sudoers

2. **Check sudoers syntax:**
   ```bash
   sudo visudo -c -f /etc/sudoers.d/cd-gun
   ```

3. **Check sudo logs:**
   ```bash
   sudo tail -f /var/log/auth.log  # Ubuntu/Debian
   sudo tail -f /var/log/secure    # CentOS/RHEL
   ```

4. **Make sure the cd-gun user exists:**
   ```bash
   id cd-gun
   ```
