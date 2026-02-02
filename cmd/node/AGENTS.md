<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# node

## Purpose

Entry point for the etlmon Node daemon. The daemon:
1. Loads configuration from `node.yaml`
2. Initializes SQLite database
3. Starts all configured collectors
4. Exposes HTTP API Gateway
5. Handles graceful shutdown

## Key Files

| File | Description |
|------|-------------|
| `main.go` | Daemon entry point and lifecycle management |

## For AI Agents

### Working In This Directory

- Single `main.go` file only
- All logic should be in `internal/` packages
- Must handle SIGTERM/SIGINT for graceful shutdown
- Default listen address: `0.0.0.0:8080`
- Default config path: `/etc/etlmon/node.yaml`

### Startup Sequence

```go
1. Parse CLI flags (-c config path)
2. Load and validate node.yaml config
3. Initialize SQLite DB (create tables if needed)
4. Start collectors (each in goroutine)
5. Start HTTP server
6. Wait for shutdown signal
7. Graceful stop: stop collectors, close DB, stop HTTP
```

### Example Usage

```bash
# Run with default config
./etlmon-node

# Run with custom config
./etlmon-node -c /path/to/node.yaml

# Run as systemd service
systemctl start etlmon-node
```

### Dependencies

#### Internal
- `internal/config` - Configuration loading
- `internal/db` - Database initialization
- `internal/collector/*` - All collector packages
- `internal/api` - HTTP API server

<!-- MANUAL: -->
