# Storage Package

## Overview

The `storage` package provides the persistence layer for etlmon's aggregated metrics. It uses SQLite as the storage backend with a CGO-free driver (modernc.org/sqlite) for maximum portability and single-binary distribution.

## Architecture

### Components

- **SQLiteStore**: Thread-safe SQLite storage implementation
- **AggregatedMetric**: Core data structure for aggregated metrics
- **GetMetricsOptions**: Flexible query filtering options

### Database Schema

```sql
CREATE TABLE aggregated_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    resource_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    aggregated_value REAL NOT NULL,
    window_size TEXT NOT NULL,
    aggregation_type TEXT NOT NULL
);

-- Indexes for optimized queries
CREATE INDEX idx_metrics_timestamp ON aggregated_metrics(timestamp);
CREATE INDEX idx_metrics_resource ON aggregated_metrics(resource_type);
CREATE INDEX idx_metrics_window ON aggregated_metrics(window_size);
CREATE INDEX idx_metrics_composite ON aggregated_metrics(resource_type, window_size, timestamp);
```

## Usage

### Creating a Store

```go
store, err := storage.NewSQLiteStore("etlmon.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Initialize schema
if err := store.Initialize(); err != nil {
    log.Fatal(err)
}
```

### Saving Metrics

```go
// Save single metric
metric := &storage.AggregatedMetric{
    Timestamp:       time.Now().Unix(),
    ResourceType:    "cpu",
    MetricName:      "usage_percent",
    AggregatedValue: 45.5,
    WindowSize:      "1m",
    AggregationType: "avg",
}
err := store.SaveAggregatedMetric(metric)

// Save batch of metrics (uses transaction)
metrics := []*storage.AggregatedMetric{
    {Timestamp: now, ResourceType: "cpu", ...},
    {Timestamp: now, ResourceType: "memory", ...},
}
err := store.SaveBatch(metrics)
```

### Querying Metrics

```go
// Get all metrics
metrics, err := store.GetMetrics(storage.GetMetricsOptions{})

// Filter by resource type
metrics, err := store.GetMetrics(storage.GetMetricsOptions{
    ResourceType: "cpu",
})

// Filter by time range
metrics, err := store.GetMetrics(storage.GetMetricsOptions{
    StartTime: startTime,
    EndTime:   endTime,
})

// Complex filter with limit
metrics, err := store.GetMetrics(storage.GetMetricsOptions{
    ResourceType:    "cpu",
    WindowSize:      "5m",
    AggregationType: "max",
    Limit:           100,
})

// Get latest metrics for a resource/window
metrics, err := store.GetLatestMetrics("cpu", "1m")
```

## Features

### Thread-Safety
- All operations are protected by mutex locks
- Safe for concurrent use from multiple goroutines
- Tested with race detector

### Performance Optimizations
- WAL mode for improved concurrency
- Composite indexes for common query patterns
- Batch inserts use transactions
- Prepared statements for batch operations

### Error Handling
- All errors are wrapped with context
- Connection tested on initialization
- Idempotent operations (Initialize, Close)

## Testing

### Running Tests

```bash
# Run all tests
go test ./internal/storage/... -v

# Check coverage
go test ./internal/storage/... -cover

# Run with race detector
go test -race ./internal/storage/...

# Generate coverage report
go test ./internal/storage/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

Current coverage: **83.0%**

Tests cover:
- Database creation and initialization
- Schema creation (idempotent)
- Single and batch metric insertion
- All query filters (resource, window, time range, aggregation type)
- Limit functionality
- Latest metrics retrieval
- Thread-safety (via race detector)
- Error conditions

## Design Decisions

### CGO-Free SQLite Driver
Uses `modernc.org/sqlite` instead of `mattn/go-sqlite3` to avoid CGO dependencies. This ensures:
- Single binary distribution
- Cross-compilation without C toolchain
- Consistent behavior across platforms

### WAL Mode
Write-Ahead Logging mode is enabled for:
- Better concurrent read/write performance
- Reduced blocking for readers
- Improved crash recovery

### Composite Index
The `(resource_type, window_size, timestamp)` composite index optimizes the most common query pattern: fetching metrics for a specific resource and window, ordered by time.

## Integration Points

### With Aggregator
```go
// Aggregator saves completed window results
result := aggregator.GetWindowResult()
metric := &storage.AggregatedMetric{
    Timestamp:       result.EndTime,
    ResourceType:    result.ResourceType,
    MetricName:      result.MetricName,
    AggregatedValue: result.Value,
    WindowSize:      result.WindowSize,
    AggregationType: result.AggType,
}
store.SaveAggregatedMetric(metric)
```

### With TUI
```go
// TUI fetches latest metrics for display
metrics, err := store.GetLatestMetrics("cpu", "1m")
for _, m := range metrics {
    display.UpdateMetric(m.MetricName, m.AggregatedValue)
}
```

## Future Enhancements

Potential improvements:
- Metric retention policy (auto-delete old metrics)
- Data compression for historical data
- Export functionality (CSV, JSON)
- Aggregation-on-read for custom time ranges
- Database migration tooling
