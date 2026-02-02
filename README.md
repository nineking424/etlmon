# etlmon

<p align="center">
  <strong>Node-based ETL / Filesystem / Process Monitor</strong>
</p>

<p align="center">
  A lightweight, real-time monitoring solution for ETL pipelines, filesystems, and server processes.<br>
  Built with Go for high performance and minimal resource usage.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#api-reference">API Reference</a> •
  <a href="#development">Development</a>
</p>

---

## Overview

**etlmon** is a monitoring tool designed for ETL operators, data platform engineers, and SREs who need real-time visibility into their data pipelines and server infrastructure. It provides:

- **Filesystem monitoring** — Track disk usage across all mounted filesystems
- **Path scanning** — Count files and directories with configurable intervals per path
- **Log tailing** — Real-time log viewing with logrotate support
- **Process monitoring** — Watch specific processes with CPU/memory metrics and kill capability
- **Cron monitoring** — Parse crontabs and display next run times
- **FTP transfer logs** — Parse vsftpd xferlog for transfer analytics

The system follows a **client-daemon architecture** where:
- **Node (Daemon)** runs on each monitored server, collects metrics, stores in SQLite, and exposes an HTTP API
- **UI (TUI Client)** connects to multiple nodes and displays real-time data in a terminal interface

---

## Features

### Core Monitoring

| Feature | Description |
|---------|-------------|
| **Filesystem Usage** | Real-time disk usage for all mounts with configurable warning thresholds |
| **Path Statistics** | File/directory counts with per-path scan intervals and exclusion patterns |
| **Log Tailing** | Live log viewing with fsnotify, automatic logrotate handling |
| **Process Metrics** | CPU%, memory RSS, runtime for watched processes |
| **Cron Jobs** | Parse system/user crontabs, calculate and display next run times |
| **FTP Transfers** | Parse vsftpd xferlog with filtering by user, host, filename |

### Architecture Highlights

| Feature | Description |
|---------|-------------|
| **Single Binary** | Both node and UI compile to single static Go binaries |
| **SQLite Storage** | Local, WAL-mode database with automatic retention policies |
| **HTTP JSON API** | RESTful API for all queries and commands |
| **Multi-Node Support** | UI can connect to multiple nodes simultaneously |
| **Vim-style Navigation** | Keyboard-driven TUI with familiar keybindings |
| **Graceful Handling** | Context-based cancellation, proper signal handling |

### Safety Features

| Feature | Description |
|---------|-------------|
| **Non-blocking Scans** | Path scans never block API responses |
| **Scan Timeouts** | Configurable timeouts prevent runaway operations |
| **NFS Throttling** | Rate limiting for operations on network filesystems |
| **Kill Confirmation** | Process kill requires explicit confirmation |
| **Single Writer** | SQLite single-writer pattern prevents lock contention |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MONITORED SERVERS                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                        etlmon Node (Daemon)                          │   │
│   │                                                                      │   │
│   │   ┌─────────────────────────────────────────────────────────────┐   │   │
│   │   │                    Collector Subsystem                       │   │   │
│   │   │                                                              │   │   │
│   │   │   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │   │   │
│   │   │   │   Disk   │  │   Path   │  │   Log    │  │ Process  │   │   │   │
│   │   │   │Collector │  │ Scanner  │  │ Tailer   │  │Collector │   │   │   │
│   │   │   └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │   │   │
│   │   │        │             │             │             │          │   │   │
│   │   │   ┌────┴─────┐  ┌────┴─────┐                               │   │   │
│   │   │   │   Cron   │  │ Xferlog  │                               │   │   │
│   │   │   │Collector │  │ Parser   │                               │   │   │
│   │   │   └────┬─────┘  └────┬─────┘                               │   │   │
│   │   │        │             │                                      │   │   │
│   │   └────────┼─────────────┼──────────────────────────────────────┘   │   │
│   │            │             │                                          │   │
│   │            ▼             ▼                                          │   │
│   │   ┌─────────────────────────────────────────────────────────────┐   │   │
│   │   │                  SQLite Database (WAL)                       │   │   │
│   │   │   filesystem_usage | path_stats | log_lines | process_stats  │   │   │
│   │   │   cron_jobs | xferlog_entries | meta                         │   │   │
│   │   └─────────────────────────────────────────────────────────────┘   │   │
│   │            │                                                        │   │
│   │            ▼                                                        │   │
│   │   ┌─────────────────────────────────────────────────────────────┐   │   │
│   │   │                    HTTP API Gateway                          │   │   │
│   │   │   GET /api/v1/fs | /paths | /logs | /processes | /cron      │   │   │
│   │   │   POST /api/v1/processes/{pid}/kill | /paths/scan           │   │   │
│   │   └─────────────────────────────────────────────────────────────┘   │   │
│   │                              │                                      │   │
│   └──────────────────────────────┼──────────────────────────────────────┘   │
│                                  │                                          │
└──────────────────────────────────┼──────────────────────────────────────────┘
                                   │
                                   │ HTTP JSON API
                                   │
┌──────────────────────────────────┼──────────────────────────────────────────┐
│                                  │                                          │
│   ┌──────────────────────────────┴──────────────────────────────────────┐   │
│   │                       etlmon UI (TUI Client)                         │   │
│   │                                                                      │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │  Node Tabs: [prod-etl-01] [prod-etl-02] [staging]            │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   │                                                                      │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │                         Views                                 │   │   │
│   │   │   [1] FS    [2] Paths   [3] Logs                             │   │   │
│   │   │   [4] Procs [5] Cron    [6] Xferlog                          │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   │                                                                      │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │  Mount        Total      Used       Avail      Use%          │   │   │
│   │   │  ─────────────────────────────────────────────────────────── │   │   │
│   │   │  /            100G       45G        55G        45%           │   │   │
│   │   │  /data        500G       420G       80G        84%  ⚠        │   │   │
│   │   │  /backup      1T         200G       800G       20%           │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   │                                                                      │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │  Node: prod-etl-01 | Status: OK | Updated: 10:23:45          │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   │                                                                      │   │
│   └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│                              OPERATOR WORKSTATION                            │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Collection**: Each collector runs as independent goroutines, gathering metrics at configurable intervals
2. **Storage**: Collectors write to SQLite database using short transactions (single-writer pattern)
3. **Query**: API Gateway reads from database and returns JSON responses
4. **Display**: UI polls API endpoints and renders data in terminal

---

## Installation

### Prerequisites

- **Go 1.20+** (for building from source)
- **Linux** (target platform; macOS for development)
- **SQLite 3.x** (included via go-sqlite3)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/nineking424/etlmon.git
cd etlmon

# Build both binaries
make build

# Or build individually
go build -o bin/etlmon-node ./cmd/node
go build -o bin/etlmon-ui ./cmd/ui
```

### Install

```bash
# Run installation script (requires sudo)
./scripts/install.sh

# Or manually:
sudo cp bin/etlmon-node /usr/local/bin/
sudo cp bin/etlmon-ui /usr/local/bin/
sudo mkdir -p /etc/etlmon /var/lib/etlmon
sudo cp configs/node.yaml /etc/etlmon/
```

### systemd Service

```bash
# Install and enable service
sudo cp scripts/etlmon-node.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable etlmon-node
sudo systemctl start etlmon-node
```

---

## Quick Start

### 1. Configure the Node

Edit `/etc/etlmon/node.yaml`:

```yaml
node:
  listen: 0.0.0.0:8080
  node_name: my-server
  db_path: /var/lib/etlmon/etlmon.db

paths:
  - path: /data/input
    scan_interval: 5m

logs:
  - name: app
    path: /var/log/app.log
    follow: true

process_watch:
  - name: etl_worker
    match: "etl_worker"
```

### 2. Start the Node

```bash
# Start directly
etlmon-node -c /etc/etlmon/node.yaml

# Or via systemd
sudo systemctl start etlmon-node
```

### 3. Configure the UI

Create `~/.config/etlmon/ui.yaml`:

```yaml
nodes:
  - name: my-server
    address: http://localhost:8080

ui:
  refresh_interval: 2s
  default_node: my-server
```

### 4. Run the UI

```bash
etlmon-ui -c ~/.config/etlmon/ui.yaml
```

### 5. Verify with curl

```bash
# Check health
curl http://localhost:8080/api/v1/health

# Get filesystem usage
curl http://localhost:8080/api/v1/fs

# Get path statistics
curl http://localhost:8080/api/v1/paths
```

---

## Configuration

### Node Configuration (`node.yaml`)

```yaml
# =============================================================================
# etlmon Node Configuration
# =============================================================================

# Node identity and network settings
node:
  # Address to listen on
  listen: 0.0.0.0:8080

  # Unique name for this node (shown in UI)
  node_name: prod-etl-01

  # SQLite database path
  db_path: /var/lib/etlmon/etlmon.db

# =============================================================================
# Collection Intervals
# =============================================================================
refresh:
  # Filesystem usage collection interval
  disk: 15s

  # Default interval for path scanning (can be overridden per path)
  default_path_scan: 5m

  # Process statistics collection interval
  process: 5s

# =============================================================================
# Path Monitoring
# =============================================================================
# Monitor directories for file counts
paths:
  - path: /data/logs
    scan_interval: 1m        # Override default interval
    max_depth: 5             # Maximum directory depth to scan
    timeout: 30s             # Abort scan after this duration
    exclude:                 # Glob patterns to exclude
      - "*.tmp"
      - "*.swp"
      - ".git"

  - path: /data/input
    scan_interval: 5m
    max_depth: 3

  - path: /data/output
    scan_interval: 10m

# =============================================================================
# Log Monitoring
# =============================================================================
# Tail log files in real-time
logs:
  - name: application        # Identifier (used in API)
    path: /var/log/app.log   # File path
    follow: true             # Enable real-time tailing
    buffer_lines: 500        # Lines to buffer before DB insert

  - name: error
    path: /var/log/error.log
    follow: true
    buffer_lines: 100

  - name: access
    path: /var/log/access.log
    follow: true

# =============================================================================
# Process Monitoring
# =============================================================================
# Watch specific processes
process_watch:
  # Match by process name pattern (regex)
  - name: etl_worker
    match: "etl_worker.*"

  # Match by command line pattern
  - name: scheduler
    match: "python.*scheduler.py"

  # Match by exact PID (optional, for debugging)
  # - pid: 12345

# =============================================================================
# Cron Monitoring
# =============================================================================
cron:
  enabled: true
  # Crontab locations are auto-detected:
  # - /etc/crontab
  # - /etc/cron.d/*
  # - /var/spool/cron/crontabs/*

# =============================================================================
# FTP Transfer Log (vsftpd)
# =============================================================================
xferlog:
  path: /var/log/xferlog
  # Only parse entries after this timestamp (ISO 8601)
  parse_start: "2026-01-01T00:00:00Z"

# =============================================================================
# Data Retention
# =============================================================================
retention:
  # Log lines retention
  log_lines_days: 7          # Keep for N days
  log_lines_max: 100000      # Or maximum N rows (whichever is reached first)

  # Xferlog retention
  xferlog_days: 30
  xferlog_max: 500000
```

### UI Configuration (`ui.yaml`)

```yaml
# =============================================================================
# etlmon UI Configuration
# =============================================================================

# Nodes to connect to
nodes:
  - name: prod-etl-01
    address: http://10.0.0.11:8080

  - name: prod-etl-02
    address: http://10.0.0.12:8080

  - name: staging
    address: http://staging.internal:8080

# UI behavior
ui:
  # Data refresh interval
  refresh_interval: 2s

  # Node to connect to on startup
  default_node: prod-etl-01
```

---

## UI Navigation

### Global Keys

| Key | Action |
|-----|--------|
| `q` | Quit application |
| `?` | Show help |
| `Tab` | Next node |
| `Shift+Tab` | Previous node |
| `1` | Filesystem view |
| `2` | Paths view |
| `3` | Logs view |
| `4` | Processes view |
| `5` | Cron view |
| `6` | Xferlog view |
| `r` | Force refresh |

### List Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Select / expand |
| `Esc` | Back / cancel |

### View-Specific Keys

| View | Key | Action |
|------|-----|--------|
| Logs | `f` | Toggle follow mode |
| Logs | `p` | Pause |
| Logs | `/` | Search |
| Processes | `d` | Kill process (with confirmation) |
| Paths | `s` | Trigger immediate scan |

### Color Coding

| Color | Meaning |
|-------|---------|
| 🟢 Green | Normal / healthy |
| 🟡 Yellow | Warning (e.g., disk > 75%) |
| 🔴 Red | Critical (e.g., disk > 90%) |
| 🔵 Blue | Informational |
| ⚪ Gray | Inactive / stale data |

---

## API Reference

### Base URL

```
http://<node-address>:8080/api/v1
```

### Endpoints

#### Health Check

```http
GET /api/v1/health
```

**Response:**
```json
{
  "data": {
    "status": "ok",
    "node_name": "prod-etl-01",
    "uptime_sec": 86400,
    "db_size_bytes": 10485760
  }
}
```

#### Filesystem Usage

```http
GET /api/v1/fs
```

**Response:**
```json
{
  "data": [
    {
      "mount_point": "/data",
      "total_bytes": 1073741824000,
      "used_bytes": 536870912000,
      "avail_bytes": 536870912000,
      "used_percent": 50.0,
      "collected_at": "2026-01-15T10:00:00Z"
    }
  ]
}
```

#### Path Statistics

```http
GET /api/v1/paths
GET /api/v1/paths?limit=10&offset=0
```

**Response:**
```json
{
  "data": [
    {
      "path": "/data/logs",
      "file_count": 15234,
      "dir_count": 42,
      "scan_duration_ms": 1523,
      "status": "OK",
      "collected_at": "2026-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "limit": 10,
    "offset": 0
  }
}
```

#### Trigger Path Scan

```http
POST /api/v1/paths/scan
Content-Type: application/json

{
  "paths": ["/data/logs", "/data/input"]
}
```

**Response:**
```json
{
  "data": {
    "status": "scan_triggered",
    "paths": ["/data/logs", "/data/input"]
  }
}
```

#### Log Files

```http
GET /api/v1/logs
```

**Response:**
```json
{
  "data": [
    {
      "name": "application",
      "path": "/var/log/app.log",
      "following": true,
      "line_count": 50000
    }
  ]
}
```

#### Log Lines

```http
GET /api/v1/logs/{name}?limit=100&offset=0
```

**Response:**
```json
{
  "data": [
    {
      "id": 50000,
      "log_name": "application",
      "line": "2026-01-15 10:00:00 INFO Processing batch 1234",
      "created_at": "2026-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 50000,
    "limit": 100,
    "offset": 0
  }
}
```

#### Processes

```http
GET /api/v1/processes
```

**Response:**
```json
{
  "data": [
    {
      "pid": 12345,
      "process_name": "etl_worker",
      "cmdline": "/usr/bin/python etl_worker.py --config prod.yaml",
      "start_time": "2026-01-14T08:00:00Z",
      "runtime_sec": 93600,
      "cpu_percent": 25.5,
      "mem_rss_bytes": 536870912,
      "status": "RUNNING",
      "collected_at": "2026-01-15T10:00:00Z"
    }
  ]
}
```

#### Kill Process

```http
POST /api/v1/processes/{pid}/kill
Content-Type: application/json

{
  "signal": "SIGTERM",
  "confirm": true
}
```

**Response:**
```json
{
  "data": {
    "status": "killed",
    "pid": 12345,
    "signal": "SIGTERM"
  }
}
```

**Errors:**
- `400 Bad Request` — Missing confirmation or invalid signal
- `403 Forbidden` — PID not in watch list
- `404 Not Found` — Process not found
- `500 Internal Server Error` — Kill failed

#### Cron Jobs

```http
GET /api/v1/cron
```

**Response:**
```json
{
  "data": [
    {
      "job_id": 1,
      "schedule": "0 * * * *",
      "command": "/usr/local/bin/hourly-job.sh",
      "user": "root",
      "source": "system",
      "next_run": "2026-01-15T11:00:00Z",
      "last_checked": "2026-01-15T10:00:00Z"
    }
  ]
}
```

#### Xferlog Entries

```http
GET /api/v1/xferlog?limit=100&offset=0
GET /api/v1/xferlog?user=ftpuser&direction=upload
```

**Response:**
```json
{
  "data": [
    {
      "id": 1000,
      "log_time": "2026-01-15T10:00:00Z",
      "remote_host": "192.168.1.100",
      "username": "ftpuser",
      "filename": "/uploads/data.csv",
      "bytes": 1048576,
      "transfer_time_sec": 5,
      "direction": "upload",
      "status": "complete"
    }
  ],
  "meta": {
    "total": 5000,
    "limit": 100,
    "offset": 0
  }
}
```

#### Database Maintenance

```http
POST /api/v1/admin/db/compact
```

**Response:**
```json
{
  "data": {
    "status": "compacted",
    "freed_bytes": 10485760
  }
}
```

### Error Responses

All errors follow this format:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": "Additional context if available"
}
```

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `BAD_REQUEST` | Invalid request parameters |
| 403 | `FORBIDDEN` | Operation not allowed |
| 404 | `NOT_FOUND` | Resource not found |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |

---

## Development

### Project Structure

```
etlmon/
├── cmd/
│   ├── node/           # Node daemon entry point
│   └── ui/             # TUI client entry point
├── internal/
│   ├── collector/      # Data collectors
│   │   ├── disk/       # Filesystem usage
│   │   ├── path/       # Path scanner
│   │   ├── log/        # Log tailer
│   │   ├── process/    # Process monitor
│   │   ├── cron/       # Cron parser
│   │   └── xferlog/    # vsftpd log parser
│   ├── api/            # HTTP API Gateway
│   │   ├── handler/    # Request handlers
│   │   └── middleware/ # Middleware
│   ├── db/             # Database layer
│   │   ├── schema/     # SQL migrations
│   │   └── repository/ # Data access
│   ├── config/         # Configuration
│   └── controller/     # Command execution
├── ui/
│   ├── views/          # TUI views
│   ├── components/     # Reusable widgets
│   └── client/         # HTTP client
├── pkg/
│   └── models/         # Shared data models
├── configs/            # Example configurations
├── scripts/            # Build and deploy scripts
├── docs/               # Documentation
└── tests/              # Integration tests
```

### Building

```bash
# Build all
make build

# Build with version
VERSION=1.0.0 make build

# Run tests
make test

# Run integration tests
make test-integration

# Run linter
make lint

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (requires SQLite)
go test -tags=integration ./tests/...

# Load tests
go test -tags=load ./tests/... -v

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` and `goimports`
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Context propagation for all I/O operations
- Interfaces defined where they're used

### Adding a New Collector

1. Create package in `internal/collector/<name>/`
2. Implement the `Collector` interface:
   ```go
   type Collector interface {
       Start(ctx context.Context) error
       Name() string
       Status() CollectorStatus
   }
   ```
3. Add repository methods in `internal/db/repository/`
4. Add API handlers in `internal/api/handler/`
5. Add UI view in `ui/views/`
6. Update configuration structs in `internal/config/`

---

## Roadmap

### MVP (Current)

- [x] Project structure and documentation
- [ ] Node daemon with disk and path collectors
- [ ] SQLite storage with schema migrations
- [ ] Basic HTTP API
- [ ] TUI with filesystem and path views

### v1.0

- [ ] Log tailer with logrotate handling
- [ ] Process collector with kill functionality
- [ ] Cron parser and display
- [ ] Xferlog parser
- [ ] Retention and purge jobs
- [ ] Full TUI with all views

### v2.0 (Future)

- [ ] Authentication (TLS, token-based)
- [ ] Web UI option
- [ ] Central aggregator for multi-node views
- [ ] Prometheus metrics endpoint
- [ ] Alerting integration

---

## Operations

### Monitoring etlmon Itself

```bash
# Check node status
curl http://localhost:8080/api/v1/health

# Check systemd service
sudo systemctl status etlmon-node

# View node logs
sudo journalctl -u etlmon-node -f
```

### Database Maintenance

```bash
# Trigger compaction via API
curl -X POST http://localhost:8080/api/v1/admin/db/compact

# Manual SQLite maintenance
sqlite3 /var/lib/etlmon/etlmon.db "PRAGMA wal_checkpoint(TRUNCATE);"
sqlite3 /var/lib/etlmon/etlmon.db "VACUUM;"
```

### Backup

```bash
# Backup database (hot backup with WAL)
sqlite3 /var/lib/etlmon/etlmon.db ".backup /backup/etlmon-$(date +%Y%m%d).db"
```

### Troubleshooting

| Issue | Solution |
|-------|----------|
| UI can't connect | Check node is running, firewall allows port 8080 |
| Slow path scans | Reduce `max_depth`, add `exclude` patterns |
| High disk I/O | Increase `scan_interval`, reduce `buffer_lines` |
| Database growing | Check `retention` settings, trigger compaction |
| Process kill fails | Verify node has sufficient permissions |

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- Code passes `make lint`
- Tests pass with `make test`
- New features include tests and documentation

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [tview](https://github.com/rivo/tview) — Terminal UI library
- [tcell](https://github.com/gdamore/tcell) — Terminal handling
- [fsnotify](https://github.com/fsnotify/fsnotify) — File system notifications
- [go-sqlite3](https://github.com/mattn/go-sqlite3) — SQLite driver
- [robfig/cron](https://github.com/robfig/cron) — Cron expression parsing

---

<p align="center">
  Made with ❤️ for ETL operators everywhere
</p>
