<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# etlmon

## Purpose

A Node-based ETL/Filesystem/Process Monitor for Linux servers. The system consists of a TUI client and daemon architecture where:
- **Node (Daemon)**: Collects metrics, stores in SQLite, exposes HTTP JSON API
- **UI (TUI Client)**: Connects to Nodes via API, displays real-time monitoring data

Target users: ETL operators, data platform engineers, SRE teams.

## Architecture Overview

```
┌─────────────────┐         HTTP JSON API         ┌─────────────────────────────────┐
│   UI (TUI)      │◄─────────────────────────────►│         Node (Daemon)           │
│   Go + tview    │                               │                                 │
│                 │                               │  ┌─────────────────────────┐   │
│  - FS View      │                               │  │   Collector Subsystem   │   │
│  - Path View    │                               │  │  - DiskCollector        │   │
│  - Log View     │                               │  │  - PathScanner          │   │
│  - Process View │                               │  │  - LogTailer            │   │
│  - Cron View    │                               │  │  - ProcessCollector     │   │
│  - Xferlog View │                               │  │  - CronCollector        │   │
└─────────────────┘                               │  │  - XferlogCollector     │   │
                                                  │  └──────────┬──────────────┘   │
                                                  │             │                   │
                                                  │  ┌──────────▼──────────────┐   │
                                                  │  │   SQLite (etlmon.db)    │   │
                                                  │  │   - WAL mode            │   │
                                                  │  │   - Single writer       │   │
                                                  │  └─────────────────────────┘   │
                                                  └─────────────────────────────────┘
```

## Planned Directory Structure

```
etlmon/
├── cmd/
│   ├── node/           # Node daemon entry point
│   │   └── main.go
│   └── ui/             # TUI client entry point
│       └── main.go
├── internal/
│   ├── collector/      # Data collectors
│   │   ├── disk/       # Filesystem usage collector
│   │   ├── path/       # Path scanner (file counts)
│   │   ├── log/        # Log tailer (fsnotify)
│   │   ├── process/    # Process monitor
│   │   ├── cron/       # Cron job parser
│   │   └── xferlog/    # vsftpd xferlog parser
│   ├── api/            # HTTP API Gateway
│   │   ├── handler/    # API handlers
│   │   └── middleware/ # Auth, logging, rate-limit
│   ├── db/             # SQLite repository layer
│   │   ├── schema/     # SQL migrations
│   │   └── repository/ # Data access
│   ├── config/         # Configuration parsing
│   └── controller/     # Command execution (kill, scan trigger)
├── ui/
│   ├── views/          # TUI views (fs, path, log, process, cron, xferlog)
│   ├── components/     # Reusable TUI components
│   └── client/         # HTTP client for Node API
├── pkg/                # Public/shared packages
│   └── models/         # Shared data models
├── configs/            # Example configuration files
│   ├── node.yaml
│   └── ui.yaml
├── scripts/            # Build, deploy, maintenance scripts
├── docs/               # Documentation
└── tests/              # Integration tests
```

## Key Files

| File | Description |
|------|-------------|
| `PRD.txt` | Complete Product Requirements Document (450 lines) |
| `go.mod` | Go module definition (to be created) |
| `Makefile` | Build automation (to be created) |
| `configs/node.yaml` | Node daemon configuration example |
| `configs/ui.yaml` | TUI client configuration example |

## For AI Agents

### Working In This Directory

1. **Language**: Go 1.20+ with Go Modules
2. **Database**: SQLite with WAL mode, single-writer pattern
3. **TUI Framework**: tcell + tview
4. **Key Libraries**:
   - `github.com/rivo/tview` - TUI framework
   - `github.com/fsnotify/fsnotify` - File system notifications
   - `github.com/mattn/go-sqlite3` - SQLite driver
   - `github.com/robfig/cron/v3` - Cron expression parsing

### Code Patterns

1. **Package Structure**:
   - `internal/` for private packages
   - `pkg/` for potentially reusable packages
   - `cmd/` for executables

2. **Error Handling**:
   - Wrap errors with `fmt.Errorf("context: %w", err)`
   - Use `context.Context` for cancellation/timeout

3. **Collector Pattern**:
   - Each collector runs as independent goroutine(s)
   - Collectors write to SQLite, API reads from SQLite
   - Per-path/per-source intervals configurable

4. **Repository Pattern**:
   - DB access through repository interfaces
   - Small transactions, no long-running locks

### Testing Requirements

- Unit tests for: parsers, collectors, DB layer
- Integration tests for: Node startup → Collector → API flow
- Load tests for: Path scanning (100k files), Log ingestion (500 lines/sec)

---

## TDD Rules (MANDATORY)

> **This project follows strict Test-Driven Development. No exceptions.**

### The TDD Cycle

```
┌─────────────────────────────────────────────────────────────────┐
│                     RED → GREEN → REFACTOR                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────┐      ┌─────────┐      ┌───────────┐              │
│   │   RED   │ ───► │  GREEN  │ ───► │ REFACTOR  │ ───┐         │
│   │         │      │         │      │           │    │         │
│   │ Write a │      │ Write   │      │ Clean up  │    │         │
│   │ failing │      │ minimal │      │ the code  │    │         │
│   │ test    │      │ code to │      │ while     │    │         │
│   │ first   │      │ pass    │      │ tests     │    │         │
│   │         │      │         │      │ stay      │    │         │
│   └─────────┘      └─────────┘      │ green     │    │         │
│                                     └───────────┘    │         │
│        ▲                                             │         │
│        └─────────────────────────────────────────────┘         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Golden Rules

| Rule | Description |
|------|-------------|
| **1. Test First** | NEVER write production code without a failing test |
| **2. One Test at a Time** | Write one failing test, make it pass, then write the next |
| **3. Minimal Code** | Write only enough code to make the test pass |
| **4. Refactor with Green** | Only refactor when all tests are passing |
| **5. No Untested Code** | Every line of production code exists because a test required it |

### For AI Agents: TDD Protocol

**BEFORE writing any production code, you MUST:**

1. **Create the test file first** (`*_test.go`)
2. **Write a failing test** that describes the expected behavior
3. **Run the test** to confirm it fails (RED)
4. **Write minimal production code** to make it pass (GREEN)
5. **Run the test** to confirm it passes
6. **Refactor** if needed while keeping tests green
7. **Repeat** for the next piece of functionality

### Test File Naming Convention

```
Production File          Test File
─────────────────────────────────────────
collector.go         →   collector_test.go
parser.go            →   parser_test.go
handler.go           →   handler_test.go
repository.go        →   repository_test.go
```

### Test Function Naming

```go
// Pattern: Test<Function>_<Scenario>_<ExpectedBehavior>

func TestDiskCollector_CollectAll_ReturnsAllMounts(t *testing.T)
func TestDiskCollector_CollectAll_ExcludesPseudoFS(t *testing.T)
func TestDiskCollector_CollectAll_HandlesPermissionError(t *testing.T)

func TestPathScanner_Scan_CountsFilesCorrectly(t *testing.T)
func TestPathScanner_Scan_RespectsMaxDepth(t *testing.T)
func TestPathScanner_Scan_AppliesExcludePatterns(t *testing.T)
func TestPathScanner_Scan_TimesOutOnLongScan(t *testing.T)

func TestXferlogParser_Parse_ExtractsAllFields(t *testing.T)
func TestXferlogParser_Parse_HandlesUploadDirection(t *testing.T)
func TestXferlogParser_Parse_SkipsMalformedLines(t *testing.T)
```

### Table-Driven Tests (Preferred)

```go
func TestCronParser_NextRun(t *testing.T) {
    tests := []struct {
        name     string
        schedule string
        now      time.Time
        want     time.Time
        wantErr  bool
    }{
        {
            name:     "every hour",
            schedule: "0 * * * *",
            now:      time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
            want:     time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
        },
        {
            name:     "invalid expression",
            schedule: "invalid",
            wantErr:  true,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseNextRun(tt.schedule, tt.now)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseNextRun() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && !got.Equal(tt.want) {
                t.Errorf("ParseNextRun() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Coverage Requirements

| Package Type | Minimum Coverage |
|--------------|------------------|
| `collector/*` | 80% |
| `db/repository` | 85% |
| `api/handler` | 80% |
| `config` | 90% |
| `pkg/models` | 70% (mostly data structures) |
| `ui/*` | 60% (UI is harder to test) |

**Overall project target: 80%+**

### Testing Patterns by Component

#### Collectors
```go
func TestDiskCollector_Collect(t *testing.T) {
    // Setup: Create mock filesystem or use temp directory
    // Act: Run collector
    // Assert: Check collected data matches expected
}
```

#### Repositories (Database)
```go
func TestFSRepository_Save(t *testing.T) {
    // Setup: In-memory SQLite database
    db, _ := sql.Open("sqlite3", ":memory:")
    repo := NewFSRepository(db)

    // Act: Save data
    err := repo.Save(ctx, &FilesystemUsage{...})

    // Assert: Data persisted correctly
    got, _ := repo.GetAll(ctx)
    assert.Equal(t, expected, got)
}
```

#### API Handlers
```go
func TestFSHandler_List(t *testing.T) {
    // Setup: Mock repository, create handler
    mockRepo := &MockRepository{...}
    handler := NewFSHandler(mockRepo)

    // Act: Make HTTP request
    req := httptest.NewRequest("GET", "/api/v1/fs", nil)
    w := httptest.NewRecorder()
    handler.List(w, req)

    // Assert: Check response
    assert.Equal(t, http.StatusOK, w.Code)
    // Parse JSON and verify content
}
```

#### Parsers
```go
func TestXferlogParser_Parse(t *testing.T) {
    // Setup: Sample log line
    line := "Mon Jan 15 10:23:45 2026 5 192.168.1.100 1234567 /file.txt b _ i r user ftp 0 * c"

    // Act: Parse
    entry, err := ParseXferlogLine(line)

    // Assert: All fields extracted correctly
    assert.NoError(t, err)
    assert.Equal(t, "192.168.1.100", entry.RemoteHost)
    assert.Equal(t, int64(1234567), entry.Bytes)
    assert.Equal(t, "upload", entry.Direction)
}
```

### TDD Workflow Example: Adding a New Feature

**Task**: Add a new metric "total scan time" to path scanner

**Step 1: Write failing test (RED)**
```go
// internal/collector/path/scanner_test.go

func TestPathScanner_Scan_TracksTotalScanTime(t *testing.T) {
    scanner := NewScanner(testDB, PathConfig{Path: testDir})

    stats, err := scanner.ScanOnce(context.Background())

    require.NoError(t, err)
    assert.Greater(t, stats.ScanDurationMs, int64(0), "scan duration should be tracked")
}
```

**Step 2: Run test → FAILS** (ScanDurationMs doesn't exist yet)

**Step 3: Write minimal code (GREEN)**
```go
// internal/collector/path/scanner.go

type PathStats struct {
    // ... existing fields
    ScanDurationMs int64 // ADD THIS
}

func (s *Scanner) ScanOnce(ctx context.Context) (*PathStats, error) {
    start := time.Now() // ADD THIS

    // ... existing scan logic

    return &PathStats{
        // ... existing fields
        ScanDurationMs: time.Since(start).Milliseconds(), // ADD THIS
    }, nil
}
```

**Step 4: Run test → PASSES**

**Step 5: Refactor if needed** (keeping tests green)

**Step 6: Add more tests** for edge cases

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./internal/collector/path/...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestPathScanner_Scan ./internal/collector/path/

# Run tests with race detection
go test -race ./...
```

### CI/CD Test Requirements

All PRs must:
- [ ] Have tests for all new functionality
- [ ] Pass all existing tests
- [ ] Maintain or improve coverage percentage
- [ ] Pass race detection (`-race` flag)

### What NOT To Do

| Anti-Pattern | Why It's Wrong |
|--------------|----------------|
| Writing tests after code | Defeats the purpose of TDD |
| Testing implementation details | Tests become brittle |
| Ignoring test failures | "It works on my machine" |
| Skipping edge cases | Bugs hide in edge cases |
| Writing tests that always pass | Useless tests |
| Mocking everything | Integration issues missed |
| Large test functions | Hard to understand failures |

### TDD Checklist for Every Change

Before submitting any code:

- [ ] I wrote the test FIRST
- [ ] The test failed before I wrote the production code
- [ ] The test passes now
- [ ] I only wrote code required by the test
- [ ] Test names describe the behavior being tested
- [ ] Coverage did not decrease
- [ ] I ran `go test -race ./...`

---

### API Design Principles

- Query endpoints: GET (read-only)
- Command endpoints: POST/DELETE (with confirmation)
- Response format: JSON
- Pagination: Required for log_lines, xferlog_entries

## Database Schema (SQLite)

| Table | Purpose |
|-------|---------|
| `filesystem_usage` | Mount point disk usage snapshots |
| `path_stats` | File/directory counts per monitored path |
| `log_lines` | Tailed log lines (with retention) |
| `process_stats` | Watched process metrics |
| `cron_jobs` | Parsed cron entries with next_run |
| `xferlog_entries` | Parsed vsftpd transfer logs |
| `meta` | Schema version, migrations |

## API Endpoints (v1)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/fs` | Filesystem usage |
| GET | `/api/v1/paths` | Path statistics |
| POST | `/api/v1/paths/scan` | Trigger immediate scan |
| GET | `/api/v1/logs` | Log file list |
| GET | `/api/v1/logs/{name}` | Log lines (with pagination) |
| GET | `/api/v1/processes` | Process list |
| POST | `/api/v1/processes/{pid}/kill` | Kill process (with confirm) |
| GET | `/api/v1/cron` | Cron jobs |
| GET | `/api/v1/xferlog` | FTP transfer log |
| GET | `/api/v1/health` | Health check |

## Development Roadmap

### MVP (2 weeks)
- [ ] Node: disk collector, path scanner, SQLite, basic API
- [ ] UI: single node connection, FS view, Path view
- [ ] Config: node.yaml, ui.yaml parsing
- [ ] Tests: path scan correctness, API basics

### v1 (4-6 weeks)
- [ ] Log tailer with logrotate handling
- [ ] Process collector with kill functionality
- [ ] Cron parsing and display
- [ ] Xferlog parsing
- [ ] Retention/purge jobs

### v2 (Future)
- [ ] Authentication (TLS, tokens)
- [ ] Web UI option
- [ ] Central aggregator
- [ ] Prometheus metrics

## Dependencies

### External (Go)
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/fsnotify/fsnotify` - File watching
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/robfig/cron/v3` - Cron parsing
- `gopkg.in/yaml.v3` - YAML config parsing

### System
- Linux (target platform)
- `/proc` filesystem access
- SQLite 3.x

<!-- MANUAL: Project-specific notes below this line are preserved on regeneration -->
