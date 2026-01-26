# etlmon

A Go-based TUI resource monitoring and aggregation tool. Collects system metrics (CPU, Memory, Disk), aggregates them over configurable time windows, persists results to SQLite, and displays data via a terminal UI.

## Features

- **Real-time Monitoring**: Live CPU, memory, and disk metrics
- **Time-Window Aggregation**: AVG, MAX, MIN, LAST over 1m, 5m, 1h windows
- **Persistent Storage**: SQLite database for historical data
- **Terminal UI**: Interactive TUI with tview
- **Single Binary**: No external runtime dependencies
- **CGO-Free**: Uses pure Go SQLite driver for easy cross-compilation

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/etlmon/etlmon.git
cd etlmon

# Build
make build

# Or build static binary (CGO-free)
make build-static
```

### Pre-built Binaries

Download from [Releases](https://github.com/etlmon/etlmon/releases).

## Usage

### Basic Usage

```bash
# Run with configuration file
./etlmon --config configs/config.yaml

# Override database path
./etlmon --config configs/config.yaml --db /tmp/metrics.db

# Show version
./etlmon --version
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Switch between Realtime and History views |
| `R` | Switch to Realtime view |
| `H` | Switch to History view |
| `1` | Filter history by 1m window |
| `5` | Filter history by 5m window |
| `0` | Filter history by 1h window |
| `Q` / `Esc` / `Ctrl+C` | Quit |

## Configuration

Create a YAML configuration file:

```yaml
# Collection interval
interval: 10s

# Resources to monitor
resources:
  - cpu
  - memory
  - disk

# Aggregation windows
windows:
  - 1m
  - 5m
  - 1h

# Aggregation functions
aggregations:
  - avg
  - max
  - min
  - last

# Database settings
database:
  path: ./etlmon.db
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `interval` | Collection interval (Go duration) | `10s` |
| `resources` | List: `cpu`, `memory`, `disk` | all |
| `windows` | Aggregation windows (Go duration) | `1m, 5m, 1h` |
| `aggregations` | Functions: `avg`, `max`, `min`, `last` | all |
| `database.path` | SQLite database path | `./etlmon.db` |

## Architecture

```
cmd/etlmon/          # Application entry point
internal/
  collector/         # System metric collection (CPU, Memory, Disk)
  aggregator/        # Time-window aggregation (AVG, MAX, MIN, LAST)
  storage/           # SQLite DAO layer
  tui/               # tview-based terminal UI
  config/            # YAML configuration parsing
```

### Data Flow

1. **Collectors** gather raw metrics at configured intervals
2. **Aggregator** processes raw data over time windows
3. **Storage** persists aggregated results when windows complete
4. **TUI** displays real-time and historical data

## Development

### Prerequisites

- Go 1.22+
- Make (optional)

### Commands

```bash
# Build
make build

# Run tests
make test

# Run with race detection
make test-race

# Generate coverage report
make coverage

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint
```

### Testing

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/aggregator/...
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| `aggregator` | 97.9% |
| `collector` | 95.9% |
| `config` | 93.6% |
| `storage` | 83.0% |
| `tui` | 73.0% |

## Database Schema

```sql
CREATE TABLE aggregated_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,      -- Unix timestamp (window end)
    resource_type TEXT NOT NULL,     -- cpu, memory, disk
    metric_name TEXT NOT NULL,       -- usage_percent, etc.
    aggregated_value REAL NOT NULL,  -- The aggregated value
    window_size TEXT NOT NULL,       -- 1m, 5m, 1h
    aggregation_type TEXT NOT NULL   -- avg, max, min, last
);
```

## License

MIT License

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests (TDD)
4. Submit a pull request
