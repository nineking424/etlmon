<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# collector

## Purpose

Data collection subsystem. Each collector is responsible for gathering specific metrics and writing them to SQLite. Collectors run as independent goroutines with configurable intervals.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `disk/` | Filesystem/mount usage collection (see `disk/AGENTS.md`) |
| `path/` | File/directory counting (see `path/AGENTS.md`) |
| `log/` | Log file tailing (see `log/AGENTS.md`) |
| `process/` | Process monitoring (see `process/AGENTS.md`) |
| `cron/` | Cron job parsing (see `cron/AGENTS.md`) |
| `xferlog/` | vsftpd transfer log parsing (see `xferlog/AGENTS.md`) |

## For AI Agents

### Working In This Directory

- Each collector is a separate package
- Collectors only **write** to DB (never read for API responses)
- Use `context.Context` for cancellation/timeout
- Each collector should implement a common interface

### Collector Interface

```go
// Collector is the common interface for all collectors
type Collector interface {
    // Start begins collection. Blocks until context is cancelled.
    Start(ctx context.Context) error

    // Name returns the collector identifier
    Name() string

    // Status returns current collector health
    Status() CollectorStatus
}

type CollectorStatus struct {
    Running     bool
    LastRun     time.Time
    LastError   error
    ItemsCount  int64
}
```

### Common Patterns

1. **Ticker-based collection**:
```go
func (c *DiskCollector) Start(ctx context.Context) error {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := c.collect(ctx); err != nil {
                c.lastError = err
            }
        }
    }
}
```

2. **Per-item intervals** (for paths, logs):
```go
// Each watched path has its own goroutine and interval
for _, path := range paths {
    go c.watchPath(ctx, path, path.Interval)
}
```

3. **Safe concurrent writes**:
```go
// Use small batches, short transactions
func (c *Collector) saveBatch(items []Item) error {
    tx, err := c.db.Begin()
    // INSERT batch
    return tx.Commit()
}
```

### Dependencies

#### Internal
- `internal/db/repository` - For writing collected data
- `internal/config` - For collector configuration

#### External
- `github.com/fsnotify/fsnotify` - For log/file watching

<!-- MANUAL: -->
