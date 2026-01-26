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

### Remaining Phases

- [ ] Phase 5: Aggregation Engine
- [ ] Phase 6: TUI Implementation
- [ ] Phase 7: Main Orchestrator
- [ ] Phase 8: Integration Tests
- [ ] Phase 9: Build System & Docs
