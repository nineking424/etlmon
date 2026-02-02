<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# configs

## Purpose

Example configuration files for Node daemon and UI client. Users should copy these and customize for their environment.

## Key Files

| File | Description |
|------|-------------|
| `node.yaml` | Example Node daemon configuration |
| `ui.yaml` | Example UI client configuration |

## For AI Agents

### Working In This Directory

- Files are EXAMPLES - should be well-commented
- Show all available options with sensible defaults
- Include explanations for non-obvious settings

### node.yaml Example

```yaml
# etlmon Node Configuration
# Copy to /etc/etlmon/node.yaml and customize

node:
  # Address to listen on (default: 0.0.0.0:8080)
  listen: 0.0.0.0:8080

  # Node name shown in UI
  node_name: prod-etl-01

  # SQLite database path
  db_path: /var/lib/etlmon/etlmon.db

# Collection intervals
refresh:
  # Filesystem usage collection interval
  disk: 15s

  # Default path scan interval (can be overridden per path)
  default_path_scan: 5m

  # Process stats collection interval
  process: 5s

# Paths to monitor for file counts
paths:
  - path: /data/logs
    scan_interval: 1m
    max_depth: 5
    exclude:
      - "*.tmp"
      - "*.swp"
    timeout: 30s

  - path: /data/input
    scan_interval: 5m
    max_depth: 3

# Log files to tail
logs:
  - name: app
    path: /var/log/app.log
    follow: true
    buffer_lines: 500

  - name: error
    path: /var/log/error.log
    follow: true

# Processes to monitor
process_watch:
  - name: etl_worker
    match: "etl_worker"

  - name: scheduler
    match: "cron.*scheduler"

# Cron monitoring
cron:
  enabled: true

# FTP transfer log (vsftpd)
xferlog:
  path: /var/log/xferlog
  # Only parse entries after this time
  parse_start: "2026-01-01T00:00:00Z"

# Data retention
retention:
  # Keep log lines for N days
  log_lines_days: 7
  # Maximum log lines to keep
  log_lines_max: 100000
  # Keep xferlog entries for N days
  xferlog_days: 30
```

### ui.yaml Example

```yaml
# etlmon UI Configuration
# Copy to ~/.config/etlmon/ui.yaml and customize

# Nodes to connect to
nodes:
  - name: prod-etl-01
    address: http://10.0.0.11:8080

  - name: prod-etl-02
    address: http://10.0.0.12:8080

  - name: staging
    address: http://staging.internal:8080

# UI settings
ui:
  # How often to refresh data from nodes
  refresh_interval: 2s

  # Which node to connect to on startup
  default_node: prod-etl-01

  # Theme (future)
  # theme: dark
```

<!-- MANUAL: -->
