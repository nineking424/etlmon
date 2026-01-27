# Changelog

All notable changes to etlmon will be documented in this file.

## [0.1.0] - 2026-01-27

### Added

#### Phase 1: Project Initialization
- Initialized Go module `github.com/etlmon/etlmon`
- Created directory structure:
  - `cmd/etlmon/` - Application entry point
  - `internal/config/` - Configuration parsing
  - `internal/storage/` - SQLite DAO layer
  - `internal/collector/` - Metric collectors
  - `internal/aggregator/` - Time-window aggregation
  - `internal/tui/` - Terminal UI
  - `configs/` - Example configurations
  - `testdata/` - Test fixtures
- Added core dependencies:
  - `modernc.org/sqlite` v1.44.3 (CGO-free SQLite)
  - `github.com/rivo/tview` v0.42.0 (Terminal UI)
  - `gopkg.in/yaml.v3` v3.0.1 (YAML parsing)
  - `github.com/shirou/gopsutil/v3` v3.24.5 (System metrics)
  - `github.com/gdamore/tcell/v2` v2.13.8 (Terminal handling)
- Verified CGO-free build works

#### Phase 2: Configuration Management (TDD)
- **Test Coverage: 93.6%** (target: >=80%)
- Implemented `internal/config/config.go`:
  - `Config` struct with YAML parsing
  - `Load()` function for reading config files
  - `Validate()` method with comprehensive validation
  - `ParseWindow()` for duration parsing
  - `GetWindowDurations()` helper
- Validation rules:
  - Required: interval, resources, windows, aggregations
  - Valid resources: cpu, memory, disk
  - Valid aggregations: avg, max, min, last
  - Window format: Go duration strings (1m, 5m, 1h)
- Created example config: `configs/config.yaml`

#### Phase 3: Storage Layer (TDD)
- **Test Coverage: 83.0%** (target: >=80%)
- Implemented `internal/storage/sqlite.go`:
  - `SQLiteStore` with thread-safe operations
  - `NewSQLiteStore()` creates DB with WAL mode
  - `Initialize()` creates schema with indexes
  - `SaveAggregatedMetric()` single insert
  - `SaveBatch()` transactional batch insert
  - `GetMetrics()` with flexible filtering
  - `GetLatestMetrics()` convenience method
  - `Close()` idempotent cleanup
- Schema (`internal/storage/schema.go`):
  - `aggregated_metrics` table
  - Indexes on timestamp, resource_type, window_size
  - Composite index for common queries
- Filter options:
  - By resource type, metric name, window size
  - By aggregation type
  - By time range (start/end)
  - With limit

#### Phase 4: Metric Collectors (TDD)
- **Test Coverage: 95.9%** (target: >=70%)
- Implemented collector interface and types:
  - `Metric` struct with ResourceType, Name, Value, Timestamp, Labels
  - `Collector` interface with Type() and Collect()
  - `Manager` for orchestrating multiple collectors
- CPU Collector (`internal/collector/cpu.go`):
  - 200ms measurement interval for accurate readings
  - Returns average usage across all cores
- Memory Collector (`internal/collector/memory.go`):
  - Collects: usage_percent, used_bytes, available_bytes, total_bytes
- Disk Collector (`internal/collector/disk.go`):
  - Filters pseudo-filesystems (tmpfs, proc, sysfs, etc.)
  - Per-partition metrics with labels
  - Collects: usage_percent, used_bytes, total_bytes
- Manager features:
  - Register multiple collectors
  - Start() with context cancellation
  - CollectOnce() for immediate collection

### Technical Details

- **TDD Methodology**: All code written test-first
- **Race Detection**: All packages pass `-race` flag
- **CGO-Free**: Builds successfully with `CGO_ENABLED=0`
- **Go Version**: 1.22+ required

#### Phase 5: Aggregation Engine (TDD)
- **Test Coverage: 97.9%** (target: >=80%)
- Implemented `internal/aggregator/aggregator.go`:
  - `Aggregator` with windowing logic
  - Support for multiple aggregation types: AVG, MAX, MIN, LAST
  - Per-metric, per-window state tracking
  - Thread-safe operations with mutex
  - Automatic window completion detection
- Aggregation features:
  - Sliding time windows (1m, 5m, 1h)
  - Configurable aggregation functions
  - Window-based data retention
  - Automatic stale data cleanup

#### Phase 6: TUI Implementation
- **Test Coverage: 73.0%** (target: >=60%)
- Implemented `internal/tui/app.go`:
  - tview-based terminal UI
  - Realtime and History view tabs
  - Live metric updates
  - Storage integration for historical data
  - Keyboard navigation and shortcuts
- UI features:
  - Tab-based navigation
  - Window filtering (1m, 5m, 1h)
  - Real-time metric streaming
  - Historical data browsing
  - Graceful shutdown handling

#### Phase 7: Main Application Orchestrator
- Implemented `cmd/etlmon/main.go`:
  - Command-line flag parsing
  - Configuration loading and validation
  - Component initialization and lifecycle
  - Goroutine coordination
  - Signal handling (SIGINT, SIGTERM)
  - Backpressure handling for metrics pipeline
- Integration features:
  - Collector → Aggregator → Storage pipeline
  - TUI updates on metric collection
  - TUI updates on aggregation completion
  - Graceful shutdown with context cancellation

#### Phase 8: Integration Testing
- **Overall Test Coverage: 90.1%**
- Created comprehensive integration tests:
  - End-to-end pipeline testing
  - Storage persistence verification
  - TUI integration testing
  - Error handling scenarios
- Performance characteristics:
  - 10-second collection interval
  - Sub-second aggregation processing
  - Efficient SQLite batch operations
  - Minimal memory footprint

#### Phase 9: Build System & Documentation
- Created `Makefile` with targets:
  - `build`, `build-static`, `build-linux`, `build-all`
  - `test`, `test-race`, `test-cover`, `coverage`
  - `lint`, `fmt`, `tidy`, `clean`, `run`, `install`
- Comprehensive `README.md` with:
  - Feature overview
  - Installation instructions
  - Usage examples
  - Configuration guide
  - Architecture documentation
  - Development workflow
  - Test coverage metrics
  - Database schema
- Version management:
  - Semantic versioning (0.1.0)
  - Git commit embedding in binary
  - `--version` flag support

### Project Statistics

- **Total Lines of Code**: ~2,500
- **Test Coverage**: 90.1% overall
- **Development Time**: Single sprint (TDD approach)
- **CGO-Free**: Yes (pure Go binary)
- **Platform Support**: Linux, macOS, cross-compilation ready

### Verification Status

All phases completed and verified:
- [x] Phase 1: Project Initialization
- [x] Phase 2: Configuration Management (TDD)
- [x] Phase 3: Storage Layer (TDD)
- [x] Phase 4: Metric Collectors (TDD)
- [x] Phase 5: Aggregation Engine (TDD)
- [x] Phase 6: TUI Implementation
- [x] Phase 7: Main Application Orchestrator
- [x] Phase 8: Integration Testing
- [x] Phase 9: Build System & Documentation
