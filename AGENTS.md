<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# etlmon

## Purpose

A Node-based ETL/Filesystem/Process Monitor for Linux servers. The system consists of a TUI client and daemon architecture where:
- **Node (Daemon)**: Collects metrics, stores in SQLite, exposes HTTP JSON API
- **UI (TUI Client)**: Connects to Nodes via API, displays real-time monitoring data

Target users: ETL operators, data platform engineers, SRE teams.

## Architecture Overview

```
┌─────────────────┐         HTTP JSON API         ┌─────────────────────────────────┐
│   UI (TUI)      │◄─────────────────────────────►│         Node (Daemon)           │
│   Go + tview    │                               │                                 │
│                 │                               │  ┌─────────────────────────┐   │
│  - FS View      │                               │  │   Collector Subsystem   │   │
│  - Path View    │                               │  │  - DiskCollector        │   │
│  - Log View     │                               │  │  - PathScanner          │   │
│  - Process View │                               │  │  - LogTailer            │   │
│  - Cron View    │                               │  │  - ProcessCollector     │   │
│  - Xferlog View │                               │  │  - CronCollector        │   │
└─────────────────┘                               │  │  - XferlogCollector     │   │
                                                  │  └──────────┬──────────────┘   │
                                                  │             │                   │
                                                  │  ┌──────────▼──────────────┐   │
                                                  │  │   SQLite (etlmon.db)    │   │
                                                  │  │   - WAL mode            │   │
                                                  │  │   - Single writer       │   │
                                                  │  └─────────────────────────┘   │
                                                  └─────────────────────────────────┘
```

## Planned Directory Structure

```
etlmon/
├── cmd/
│   ├── node/           # Node daemon entry point
│   │   └── main.go
│   └── ui/             # TUI client entry point
│       └── main.go
├── internal/
│   ├── collector/      # Data collectors
│   │   ├── disk/       # Filesystem usage collector
│   │   ├── path/       # Path scanner (file counts)
│   │   ├── log/        # Log tailer (fsnotify)
│   │   ├── process/    # Process monitor
│   │   ├── cron/       # Cron job parser
│   │   └── xferlog/    # vsftpd xferlog parser
│   ├── api/            # HTTP API Gateway
│   │   ├── handler/    # API handlers
│   │   └── middleware/ # Auth, logging, rate-limit
│   ├── db/             # SQLite repository layer
│   │   ├── schema/     # SQL migrations
│   │   └── repository/ # Data access
│   ├── config/         # Configuration parsing
│   └── controller/     # Command execution (kill, scan trigger)
├── ui/
│   ├── views/          # TUI views (fs, path, log, process, cron, xferlog)
│   ├── components/     # Reusable TUI components
│   └── client/         # HTTP client for Node API
├── pkg/                # Public/shared packages
│   └── models/         # Shared data models
├── configs/            # Example configuration files
│   ├── node.yaml
│   └── ui.yaml
├── scripts/            # Build, deploy, maintenance scripts
├── docs/               # Documentation
└── tests/              # Integration tests
```

## Key Files

| File | Description |
|------|-------------|
| `PRD.txt` | Complete Product Requirements Document (450 lines) |
| `go.mod` | Go module definition (to be created) |
| `Makefile` | Build automation (to be created) |
| `configs/node.yaml` | Node daemon configuration example |
| `configs/ui.yaml` | TUI client configuration example |

## For AI Agents

### Working In This Directory

1. **Language**: Go 1.20+ with Go Modules
2. **Database**: SQLite with WAL mode, single-writer pattern
3. **TUI Framework**: tcell + tview
4. **Key Libraries**:
   - `github.com/rivo/tview` - TUI framework
   - `github.com/fsnotify/fsnotify` - File system notifications
   - `github.com/mattn/go-sqlite3` - SQLite driver
   - `github.com/robfig/cron/v3` - Cron expression parsing

### Code Patterns

1. **Package Structure**:
   - `internal/` for private packages
   - `pkg/` for potentially reusable packages
   - `cmd/` for executables

2. **Error Handling**:
   - Wrap errors with `fmt.Errorf("context: %w", err)`
   - Use `context.Context` for cancellation/timeout

3. **Collector Pattern**:
   - Each collector runs as independent goroutine(s)
   - Collectors write to SQLite, API reads from SQLite
   - Per-path/per-source intervals configurable

4. **Repository Pattern**:
   - DB access through repository interfaces
   - Small transactions, no long-running locks

### Testing Requirements

- Unit tests for: parsers, collectors, DB layer
- Integration tests for: Node startup → Collector → API flow
- Load tests for: Path scanning (100k files), Log ingestion (500 lines/sec)

### API Design Principles

- Query endpoints: GET (read-only)
- Command endpoints: POST/DELETE (with confirmation)
- Response format: JSON
- Pagination: Required for log_lines, xferlog_entries

## Database Schema (SQLite)

| Table | Purpose |
|-------|---------|
| `filesystem_usage` | Mount point disk usage snapshots |
| `path_stats` | File/directory counts per monitored path |
| `log_lines` | Tailed log lines (with retention) |
| `process_stats` | Watched process metrics |
| `cron_jobs` | Parsed cron entries with next_run |
| `xferlog_entries` | Parsed vsftpd transfer logs |
| `meta` | Schema version, migrations |

## API Endpoints (v1)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/fs` | Filesystem usage |
| GET | `/api/v1/paths` | Path statistics |
| POST | `/api/v1/paths/scan` | Trigger immediate scan |
| GET | `/api/v1/logs` | Log file list |
| GET | `/api/v1/logs/{name}` | Log lines (with pagination) |
| GET | `/api/v1/processes` | Process list |
| POST | `/api/v1/processes/{pid}/kill` | Kill process (with confirm) |
| GET | `/api/v1/cron` | Cron jobs |
| GET | `/api/v1/xferlog` | FTP transfer log |
| GET | `/api/v1/health` | Health check |

## Development Roadmap

### MVP (2 weeks)
- [ ] Node: disk collector, path scanner, SQLite, basic API
- [ ] UI: single node connection, FS view, Path view
- [ ] Config: node.yaml, ui.yaml parsing
- [ ] Tests: path scan correctness, API basics

### v1 (4-6 weeks)
- [ ] Log tailer with logrotate handling
- [ ] Process collector with kill functionality
- [ ] Cron parsing and display
- [ ] Xferlog parsing
- [ ] Retention/purge jobs

### v2 (Future)
- [ ] Authentication (TLS, tokens)
- [ ] Web UI option
- [ ] Central aggregator
- [ ] Prometheus metrics

## Dependencies

### External (Go)
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/fsnotify/fsnotify` - File watching
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/robfig/cron/v3` - Cron parsing
- `gopkg.in/yaml.v3` - YAML config parsing

### System
- Linux (target platform)
- `/proc` filesystem access
- SQLite 3.x

<!-- MANUAL: Project-specific notes below this line are preserved on regeneration -->
