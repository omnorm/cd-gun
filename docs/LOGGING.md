# Logging

CD-Gun supports flexible logging configuration with optional file-based output.

## Configuration

### Log Level

Configure logging verbosity in the `agent` section of your config:

```yaml
agent:
  log_level: "info"  # debug, info, warn, error
```

**Levels:**
- `debug` — Detailed diagnostic information (includes file/function location)
- `info` — General informational messages
- `warn` — Warning messages for potentially problematic conditions
- `error` — Error messages (includes file/function location)

### Log File (Optional)

By default, logs are written to stdout (visible in `journalctl` when running as systemd service).

To write logs to a file, specify `log_file` in the `agent` section:

```yaml
agent:
  name: "cd-gun-agent"
  log_level: "info"
  log_file: "/var/log/cd-gun.log"
  state_dir: "/var/lib/cd-gun"
  cache_dir: "/var/cache/cd-gun/repos"
  poll_interval: "5m"
```

**Log file behavior:**
- File is created if it doesn't exist
- Logs are appended (not overwritten) on startup
- Ensure the cd-gun user has write permissions to the log directory
- File handle is properly closed on shutdown

## Usage Examples

### Log to stdout (default)

```yaml
agent:
  log_level: "info"
  # log_file omitted → logs to stdout/journalctl
```

```bash
# View logs in journalctl
journalctl -u cd-gun -f
```

### Log to file

```yaml
agent:
  log_level: "info"
  log_file: "/var/log/cd-gun.log"
```

```bash
# View logs
tail -f /var/log/cd-gun.log

# Check log file permissions
ls -l /var/log/cd-gun.log
```

### Debug mode with file logging

```yaml
agent:
  log_level: "debug"
  log_file: "/var/log/cd-gun-debug.log"
```

## Log Rotation

If you use file-based logging, configure log rotation with `logrotate`:

### Create `/etc/logrotate.d/cd-gun`:

```
/var/log/cd-gun.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0640 cd-gun cd-gun
    postrotate
        systemctl reload-or-restart cd-gun > /dev/null 2>&1 || true
    endscript
}
```

After SIGHUP, CD-Gun will reopen the log file:

```bash
kill -HUP $(pgrep cd-gun-agent)
```

## Log Format

All log entries include:
- **Timestamp** — `2025-01-27 14:30:45.123456`
- **Level** — `[DEBUG]`, `[INFO]`, `[WARN]`, `[ERROR]`
- **Message** — The actual log message
- **Location** — File and line number (DEBUG and ERROR levels only)

### Example log output:

```
[INFO] 2025-01-27 14:30:45 CD-Gun agent 'cd-gun-agent' initialized successfully
[DEBUG] 2025-01-27 14:30:45 monitor.go:45 Starting repository monitor for 'web-app'
[WARN] 2025-01-27 14:31:15 Monitor timeout for repository 'api-service', retrying...
[ERROR] 2025-01-27 14:32:00 executor.go:123 Failed to execute deploy script: command timed out
```

## Troubleshooting

### No logs appearing

1. Check log level is set to `info` or lower
2. If using file logging, verify:
   ```bash
   ls -l /var/log/cd-gun.log
   cat /var/log/cd-gun.log
   ```
3. If using journalctl:
   ```bash
   journalctl -u cd-gun -n 50  # Last 50 lines
   ```

### Permission denied writing to log file

Ensure the cd-gun user has write permissions:

```bash
sudo chown cd-gun:cd-gun /var/log/cd-gun.log
sudo chmod 640 /var/log/cd-gun.log
```

Or create the log directory with proper permissions:

```bash
sudo mkdir -p /var/log/cd-gun
sudo chown cd-gun:cd-gun /var/log/cd-gun
sudo chmod 755 /var/log/cd-gun
```

### Log file not rotating

If using logrotate, test the configuration:

```bash
sudo logrotate -f /etc/logrotate.d/cd-gun -v
```

## Command-Line Overrides

The `-log-level` flag can override the config file:

```bash
./cd-gun-agent -config /etc/cd-gun/config.yaml -log-level debug
```

Note: The configuration file's `log_level` takes precedence over the command-line flag.
