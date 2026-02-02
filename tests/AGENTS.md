<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# tests

## Purpose

Integration and end-to-end tests. Unit tests are co-located with source code (`*_test.go`).

## Key Files

| File | Description |
|------|-------------|
| `integration_test.go` | Node startup and API tests |
| `e2e_test.go` | Full UI-to-Node tests |
| `load_test.go` | Performance and load tests |
| `testdata/` | Test fixtures and sample data |

## For AI Agents

### Working In This Directory

- Use `go test` with `-tags=integration` for integration tests
- Keep test data in `testdata/` subdirectory
- Mock external dependencies where appropriate
- Tests should be runnable in CI

### Test Categories

| Category | Tag | Description |
|----------|-----|-------------|
| Unit | (default) | Fast, isolated tests |
| Integration | `integration` | Tests with real DB |
| E2E | `e2e` | Full stack tests |
| Load | `load` | Performance tests |

### Integration Test Pattern

```go
//go:build integration

package tests

import (
    "context"
    "testing"
    "time"

    "github.com/yourorg/etlmon/internal/db"
    "github.com/yourorg/etlmon/internal/collector/disk"
)

func TestDiskCollector_Integration(t *testing.T) {
    // Setup test database
    testDB, err := db.NewDB(":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer testDB.Close()

    // Create collector
    collector := disk.NewCollector(testDB, 1*time.Second)

    // Run collection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    go collector.Start(ctx)
    time.Sleep(2 * time.Second)

    // Verify data was collected
    usage, err := testDB.GetFilesystemUsage(ctx)
    if err != nil {
        t.Fatal(err)
    }

    if len(usage) == 0 {
        t.Error("expected filesystem usage data")
    }
}
```

### Load Test Pattern

```go
//go:build load

package tests

func TestPathScanner_LargeDirectory(t *testing.T) {
    // Create test directory with 100k files
    testDir := createTestFiles(t, 100000)
    defer os.RemoveAll(testDir)

    scanner := path.NewScanner(testDB, path.Config{
        Path:         testDir,
        ScanInterval: 1 * time.Minute,
        MaxDepth:     10,
        Timeout:      5 * time.Minute,
    })

    start := time.Now()
    err := scanner.ScanOnce(context.Background())
    duration := time.Since(start)

    if err != nil {
        t.Fatal(err)
    }

    t.Logf("Scanned 100k files in %v", duration)

    // Assert reasonable performance
    if duration > 30*time.Second {
        t.Errorf("scan took too long: %v", duration)
    }
}
```

### Test Data

```
testdata/
├── logs/
│   ├── sample.log      # Sample log file
│   └── rotated.log.1   # Rotated log sample
├── xferlog/
│   └── sample.xferlog  # Sample vsftpd log
└── cron/
    └── sample.crontab  # Sample crontab
```

<!-- MANUAL: -->
