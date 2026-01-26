# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

etlmon is a Go-based TUI resource monitoring and aggregation tool. It collects system metrics (CPU, Memory, Disk), aggregates them over configurable time windows, persists results to SQLite, and displays data via a terminal UI.

**Key constraints from PRD:**
- Single binary distribution (no external runtime dependencies)
- CGO-free SQLite driver (use modernc.org/sqlite)
- Local monitoring only (no distributed/remote features)
- Linux primary target

## Build Commands

```bash
# Build
go build -o etlmon ./cmd/etlmon

# Build for Linux (cross-compile)
GOOS=linux GOARCH=amd64 go build -o etlmon ./cmd/etlmon

# Run
./etlmon --config configs/config.yaml

# Test
go test ./...

# Test with coverage
go test -cover ./...

# Test single package
go test ./internal/aggregator/...

# Lint (if golangci-lint installed)
golangci-lint run
```

## Architecture

```
cmd/etlmon/          # Application entry point
internal/
  collector/         # System metric collection (CPU, Memory, Disk)
  aggregator/        # Time-window aggregation (AVG, MAX, MIN, LAST)
  storage/           # SQLite DAO layer
  tui/               # tview-based terminal UI
  config/            # YAML configuration parsing
configs/             # Example YAML configuration files
```

### Data Flow

1. **Collectors** gather raw metrics at configured intervals (memory only, not persisted)
2. **Aggregator** processes raw data over time windows (1m, 5m, 1h)
3. **Storage** persists aggregated results to SQLite when windows complete
4. **TUI** displays real-time and historical aggregated data

### SQLite Schema

```sql
-- Core aggregation results table
CREATE TABLE aggregated_metrics (
    timestamp       INTEGER,  -- window end time (Unix)
    resource_type   TEXT,     -- cpu, memory, disk
    metric_name     TEXT,     -- usage_percent, etc.
    aggregated_value REAL,
    window_size     TEXT      -- 1m, 5m, 1h
);
```

### Configuration Structure

```yaml
interval: 10s           # Collection interval
resources:
  - cpu
  - memory
  - disk
windows:
  - 1m
  - 5m
  - 1h
aggregations:
  - avg
  - max
  - min
  - last
```

## Dependencies

- **tview**: Terminal UI framework
- **modernc.org/sqlite**: CGO-free SQLite driver
- **gopkg.in/yaml.v3**: YAML parsing
- **gopsutil** (or similar): System metrics collection

## Testing Strategy

- **Unit tests**: Aggregation logic, SQLite DAO operations
- **Integration tests**: Full collect → aggregate → persist pipeline
- **Load tests**: High-frequency collection impact on resources

## Git Workflow Rules

**Auto-commit and push rules for Claude Code:**

1. **Commit After Each Phase Completion**: When a phase (e.g., Phase 2: Config, Phase 3: Storage) is fully implemented and tests pass, automatically commit with a descriptive message.

2. **Commit Message Format**:
   ```
   <type>(<scope>): <description>

   - Detail 1
   - Detail 2

   Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
   ```
   Types: `feat`, `fix`, `test`, `refactor`, `docs`, `chore`

3. **Auto-Push Policy**: Push to origin after each successful commit.

4. **Pre-Commit Checks**: Before committing, ensure:
   - All tests pass: `go test ./...`
   - Build succeeds: `go build ./...`
   - No lint errors (if golangci-lint available)

5. **Branch Strategy**: Work on `main` branch for this project.

6. **Files to Exclude from Commits**:
   - `.omc/` directory (local state)
   - `etlmon` binary
   - Any `.db` files
   - IDE/editor config files
