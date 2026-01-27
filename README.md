# etlmon

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-90.1%25-brightgreen.svg)](https://github.com/nineking424/etlmon)
[![CGO Free](https://img.shields.io/badge/CGO-Free-success.svg)](https://github.com/nineking424/etlmon)

A Go-based TUI resource monitoring and aggregation tool. Collects system metrics (CPU, Memory, Disk), aggregates them over configurable time windows, persists results to SQLite, and displays data via an interactive terminal UI.

## Quick Start

```bash
# Clone and build
git clone https://github.com/nineking424/etlmon.git
cd etlmon
make build-static

# Run with default config
./etlmon --config configs/config.yaml
```

## Features

| Feature | Description |
|---------|-------------|
| **Real-time Monitoring** | Live CPU, memory, and disk usage metrics updated at configurable intervals |
| **Time-Window Aggregation** | AVG, MAX, MIN, LAST calculations over 1m, 5m, 1h windows |
| **Persistent Storage** | SQLite database with WAL mode for historical data |
| **Interactive TUI** | tview-based terminal UI with keyboard navigation |
| **Single Binary** | No external runtime dependencies required |
| **CGO-Free** | Pure Go SQLite driver (modernc.org/sqlite) for easy cross-compilation |
| **Cross-Platform** | Builds for Linux, macOS, Windows |

## Screenshots

```
┌─ Real-time Metrics (Press Tab to switch, Q to quit) ─────────────────┐
│ ━━━ CPU ━━━                                                          │
│   usage_percent      [green]23.5%[white]                             │
│                                                                      │
│ ━━━ MEMORY ━━━                                                       │
│   available_bytes    12.4 GB                                         │
│   total_bytes        16.0 GB                                         │
│   usage_percent      [yellow]77.2%[white]                            │
│   used_bytes         12.4 GB                                         │
│                                                                      │
│ ━━━ DISK ━━━                                                         │
│   usage_percent      [green]45.1%[white] (/)                         │
│                                                                      │
│ Last updated: 15:04:05                                               │
└──────────────────────────────────────────────────────────────────────┘
│ [green]Running[white] | Last: 15:04:05 | etlmon                      │
```

## Installation

### From Source (Recommended)

```bash
# Clone repository
git clone https://github.com/nineking424/etlmon.git
cd etlmon

# Build static binary (CGO-free, recommended)
make build-static

# Or standard build
make build

# Or cross-compile for Linux
make build-linux
```

### Using Go Install

```bash
go install github.com/nineking424/etlmon/cmd/etlmon@latest
```

### Pre-built Binaries

Download from [Releases](https://github.com/nineking424/etlmon/releases).

## Usage

### Command Line Options

```bash
# Run with configuration file (required)
./etlmon --config configs/config.yaml

# Override database path
./etlmon --config configs/config.yaml --db /tmp/metrics.db

# Show version
./etlmon --version

# Show help
./etlmon --help
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Switch between Realtime and History views |
| `R` | Switch to Realtime view |
| `H` | Switch to History view |
| `1` | Filter history by 1-minute window |
| `5` | Filter history by 5-minute window |
| `0` | Filter history by 1-hour window |
| `Q` / `Esc` / `Ctrl+C` | Quit application |

## Configuration

Create a YAML configuration file (see `configs/config.yaml` for basic example or `configs/config-advanced.yaml` for advanced disk configuration):

```yaml
# Collection interval (how often to collect metrics)
interval: 10s

# Resources to monitor
resources:
  - cpu      # CPU usage percentage
  - memory   # Memory usage, available, total
  - disk     # Disk usage per mount point

# Disk-specific configuration (optional)
# If omitted, all partitions are auto-discovered with stats method
disk:
  default_method: stats  # stats (default), df, or du
  paths:
    - path: /
      method: stats      # Optional: override default method for this path
    - path: /home
      method: df

# Aggregation time windows
windows:
  - 1m       # 1-minute aggregation
  - 5m       # 5-minute aggregation
  - 1h       # 1-hour aggregation

# Aggregation functions to apply
aggregations:
  - avg      # Average value over window
  - max      # Maximum value over window
  - min      # Minimum value over window
  - last     # Last value in window

# Database settings
database:
  path: ./etlmon.db  # SQLite database file path
```

### Configuration Reference

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `interval` | duration | Metric collection interval | `10s` |
| `resources` | list | Resources to monitor: `cpu`, `memory`, `disk` | all |
| `disk.default_method` | string | Disk collection method: `stats`, `df`, `du` | `stats` |
| `disk.paths` | list | Specific filesystem paths to monitor (optional) | auto-discover all |
| `disk.paths[].path` | string | Filesystem path (mount point or directory) | - |
| `disk.paths[].method` | string | Collection method override for this path | default_method |
| `windows` | list | Aggregation windows (Go duration format) | `1m, 5m, 1h` |
| `aggregations` | list | Functions: `avg`, `max`, `min`, `last` | all |
| `database.path` | string | SQLite database file path | `./etlmon.db` |

### Disk Collection Methods

etlmon supports three methods for collecting disk metrics:

| Method | Description | Speed | Use Case |
|--------|-------------|-------|----------|
| **stats** | syscall.Statfs (filesystem-level) | Fast | Default; accurate for partitions |
| **df** | Parses df command output | Medium | Cross-validation; useful for validation |
| **du** | Recursive directory scan | Slow | Actual file usage in directories (e.g., `/var/log`) |

Example: Mix methods for different paths:
```yaml
disk:
  default_method: stats
  paths:
    - path: /              # Root: use fast filesystem stats
      method: stats
    - path: /var/log       # Logs: measure actual files recursively
      method: du
```

See `configs/config-advanced.yaml` for detailed examples.

## Architecture

```
etlmon/
├── cmd/etlmon/           # Application entry point
│   └── main.go           # CLI flags, component wiring, main loop
├── internal/
│   ├── config/           # YAML configuration parsing and validation
│   ├── collector/        # System metric collectors (CPU, Memory, Disk)
│   ├── aggregator/       # Time-window aggregation engine
│   ├── storage/          # SQLite DAO with WAL mode
│   └── tui/              # tview-based terminal UI
├── configs/              # Example configuration files
└── testdata/             # Test fixtures
```

### Data Flow

```
┌───────────┐     ┌────────────┐     ┌─────────┐     ┌─────────┐
│ Collectors│────▶│ Aggregator │────▶│ Storage │────▶│   TUI   │
│ (CPU/Mem/ │     │ (Windows)  │     │ (SQLite)│     │ (tview) │
│   Disk)   │     └────────────┘     └─────────┘     └─────────┘
└───────────┘           │                                 │
      │                 │                                 │
      └─────────────────┴─────────────────────────────────┘
                    Real-time Updates
```

1. **Collectors** gather raw metrics at configured intervals (in-memory only)
2. **Aggregator** maintains time-window buffers, calculates aggregations when windows complete
3. **Storage** persists completed aggregations to SQLite (not raw metrics)
4. **TUI** displays real-time metrics and historical aggregated data

## Development

### Prerequisites

- Go 1.22 or later
- Make (optional, for convenience)

### Build Commands

```bash
make build          # Build binary
make build-static   # Build CGO-free static binary
make build-linux    # Cross-compile for Linux amd64
make build-all      # Build for all platforms
```

### Test Commands

```bash
make test           # Run all tests
make test-race      # Run tests with race detector
make test-cover     # Run tests with coverage summary
make coverage       # Generate HTML coverage report
```

### Other Commands

```bash
make fmt            # Format code with go fmt
make lint           # Run golangci-lint (if installed)
make tidy           # Run go mod tidy
make clean          # Remove build artifacts
make run            # Build and run with default config
make help           # Show all available commands
```

### Test Coverage

| Package | Coverage | Target |
|---------|----------|--------|
| `aggregator` | 97.9% | >= 90% |
| `collector` | 95.9% | >= 70% |
| `config` | 93.6% | >= 80% |
| `storage` | 83.0% | >= 80% |
| `tui` | 73.0% | >= 50% |
| **Overall** | **90.1%** | >= 75% |

## Database Schema

etlmon stores only aggregated metrics (not raw data) in SQLite:

```sql
CREATE TABLE aggregated_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,       -- Unix timestamp (window end time)
    resource_type TEXT NOT NULL,      -- cpu, memory, disk
    metric_name TEXT NOT NULL,        -- usage_percent, used_bytes, etc.
    aggregated_value REAL NOT NULL,   -- The calculated value
    window_size TEXT NOT NULL,        -- 1m, 5m, 1h
    aggregation_type TEXT NOT NULL,   -- avg, max, min, last
    labels TEXT                       -- JSON-encoded labels (e.g., {"mountpoint": "/", "method": "stats"})
);

-- Indexes for efficient querying
CREATE INDEX idx_metrics_timestamp ON aggregated_metrics(timestamp);
CREATE INDEX idx_metrics_resource ON aggregated_metrics(resource_type);
CREATE INDEX idx_metrics_window ON aggregated_metrics(window_size);
CREATE INDEX idx_metrics_composite ON aggregated_metrics(resource_type, window_size, timestamp);
CREATE INDEX idx_metrics_labels ON aggregated_metrics(labels);
```

### Labels

Disk metrics include labels to identify collection method and filesystem details:

```json
{
  "mountpoint": "/",
  "device": "/dev/sda1",
  "fstype": "ext4",
  "method": "stats"
}
```

Labels are stored as JSON strings and can be used to filter or group historical metrics by collection method or filesystem properties.

## Troubleshooting

### "Config file required" error
```bash
# Specify config file path
./etlmon --config /path/to/config.yaml
```

### "Permission denied" for database
```bash
# Ensure write permission to database directory
chmod 755 /path/to/db/directory
```

### High CPU usage
- Increase collection interval in config (e.g., `interval: 30s`)
- Reduce number of monitored resources

### TUI not displaying correctly
- Ensure terminal supports Unicode and 256 colors
- Try resizing terminal window
- Use a modern terminal emulator (iTerm2, Alacritty, etc.)

## Metrics Reference

### CPU Metrics
| Metric | Description |
|--------|-------------|
| `usage_percent` | Average CPU utilization across all cores (0-100) |

### Memory Metrics
| Metric | Description |
|--------|-------------|
| `usage_percent` | Memory utilization percentage (0-100) |
| `used_bytes` | Bytes of memory in use |
| `available_bytes` | Bytes of memory available |
| `total_bytes` | Total system memory |

### Disk Metrics
| Metric | Description |
|--------|-------------|
| `usage_percent` | Disk utilization per mount point (0-100) |
| `used_bytes` | Bytes used on disk |
| `total_bytes` | Total disk capacity (available with stats/df methods) |

**Collection Methods:**
- Disk metrics are collected using one of three methods (configurable per path):
  - **stats**: Filesystem-level metrics via syscall.Statfs (default, fastest)
  - **df**: Parses df command output (useful for validation)
  - **du**: Recursive directory scan measuring actual files (slowest, most accurate for directories)

**Labels:** Each disk metric includes labels identifying the collection method (`stats`, `df`, `du`) and filesystem details (mountpoint, device, fstype).

**Note:** Pseudo-filesystems (tmpfs, proc, sysfs, etc.) are automatically filtered out during auto-discovery.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests first (TDD approach)
4. Implement your changes
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Acknowledgments

- [tview](https://github.com/rivo/tview) - Terminal UI library
- [modernc.org/sqlite](https://modernc.org/sqlite) - CGO-free SQLite driver
- [gopsutil](https://github.com/shirou/gopsutil) - System metrics collection
