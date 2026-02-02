<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# path

## Purpose

Scans configured directories to count files and subdirectories. Writes to `path_stats` table. Supports per-path intervals and exclusion patterns.

## Key Files

| File | Description |
|------|-------------|
| `scanner.go` | PathScanner implementation |
| `scanner_test.go` | Unit tests |
| `worker.go` | Worker pool for parallel scanning |

## For AI Agents

### Working In This Directory

- **Critical**: Never block API on long scans
- Each path has its own goroutine with independent ticker
- Use worker pool for directory traversal
- Implement scan timeout via `context.WithTimeout`
- Track scan status: OK, SCANNING, ERROR

### Data Model

```go
type PathStats struct {
    Path           string
    FileCount      int64
    DirCount       int64
    ScanDurationMs int64
    Status         string // "OK", "SCANNING", "ERROR"
    ErrorMessage   string
    CollectedAt    time.Time
}

type PathConfig struct {
    Path         string
    ScanInterval time.Duration
    MaxDepth     int
    Exclude      []string // glob patterns
    Timeout      time.Duration
}
```

### Safety Mechanisms

1. **Skip if previous scan running**:
```go
if !c.scanLock.TryLock() {
    return // Previous scan still running
}
defer c.scanLock.Unlock()
```

2. **Timeout protection**:
```go
ctx, cancel := context.WithTimeout(ctx, path.Timeout)
defer cancel()
```

3. **NFS throttling**:
```go
// Limit operations per second for NFS paths
rateLimiter := rate.NewLimiter(rate.Limit(maxOpsPerSec), 1)
```

4. **Memory limits**:
```go
// Use filepath.WalkDir (not Walk) for memory efficiency
// Don't accumulate full file list in memory
```

### Parallelization

```go
// Worker pool pattern
type Scanner struct {
    workerCount int
    dirQueue    chan string
    results     chan scanResult
}
```

<!-- MANUAL: -->
