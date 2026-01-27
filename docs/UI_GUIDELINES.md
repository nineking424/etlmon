# etlmon UI Guidelines

## Overview

etlmon is a Terminal User Interface (TUI) application built with [tview](https://github.com/rivo/tview) for displaying real-time system resource metrics and historical aggregated data. The UI provides two modal views:

- **Real-time View:** Current system metrics updated at collection intervals
- **History View:** Aggregated historical data with filtering by time window

**Key Characteristics:**
- Single-window modal view switching
- Color-coded status indicators for resource thresholds
- Thread-safe concurrent updates from background goroutines
- Scrollable views with responsive rendering

## Layout Architecture

The etlmon TUI uses a hierarchical flex-based layout:

```
┌────────────────────────────────────────────────────────┐
│ PAGES (tview.Pages) - Main Content Area (weight: 1)   │
│                                                        │
│ ┌─ RealtimeView (default)                           │
│ │  ├─ Metrics grouped by resource type              │
│ │  └─ Updated on collection intervals               │
│ │                                                   │
│ └─ HistoryView (on Tab)                             │
│    ├─ Aggregated data in table format               │
│    └─ Filterable by window size (1m/5m/1h)         │
├────────────────────────────────────────────────────────┤
│ StatusBar (fixed 1 line) - Status and last update time │
└────────────────────────────────────────────────────────┘
```

**Components:**
- Root: `tview.Flex` (FlexRow direction)
- Content: `tview.Pages` (proportional, weight = 1)
- Status: Fixed 1-line height

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Toggle between Real-time and History views |
| `R` | Switch to Real-time view |
| `H` | Switch to History view |
| `1`, `5`, `0` | Filter history by 1m, 5m, or 1h window |
| `Q`, `Esc`, `Ctrl+C` | Quit application |

Keybindings configured via `SetInputCapture()` in `App.setupKeybindings()`.

## Color Scheme

Colors use tview's markup syntax: `[colorname]text[white]`

### Theme Colors

| Color | Usage |
|-------|-------|
| **Green** `[green]` | Headers, normal values, positive status |
| **Yellow** `[yellow]` | Warning state (70-90% metrics) |
| **Red** `[red]` | Critical state (>90% metrics) |
| **Blue** `[blue]` | Column headers in tables |
| **Gray** `[gray]` | Labels, metadata, timestamps |
| **White** `[white]` | Default text, reset color |

### Threshold-Based Coloring for Percentages

```
< 70%   → [green]XX.X%[white]     (Normal)
70-90%  → [yellow]XX.X%[white]    (Warning)
> 90%   → [red]XX.X%[white]       (Critical)
```

## Components

### Real-time View

**Title:** `Real-time Metrics (Press Tab to switch, Q to quit)`

**Features:**
- Dynamic colors enabled for markup
- Scrollable content
- No text wrapping
- Bordered with left-aligned title

**Organization:**
- Metrics grouped by ResourceType (CPU, DISK, MEMORY)
- Resource types sorted alphabetically
- Metrics within group sorted alphabetically by name

**Example Output:**
```
━━━ CPU ━━━
  core0_usage_percent  30.2%
  core1_usage_percent  55.8%
  usage_percent        [green]45.5%[white]

━━━ DISK ━━━
  free_bytes           100.5 GB [gray](/, sda1, ext4, stats)[white]
  total_bytes          465.6 GB [gray](/, sda1, ext4, stats)[white]
  usage_percent        [yellow]78.5%[white] [gray](/, sda1, ext4, stats)[white]
  used_bytes           365.1 GB [gray](/, sda1, ext4, stats)[white]

━━━ MEMORY ━━━
  available_bytes      6.0 GB
  total_bytes          16.0 GB
  usage_percent        [green]62.3%[white]
  used_bytes           10.0 GB

Last updated: 09:38:56
```

**Value Formatting:**
- **Percentage metrics** (name contains "percent"): `XX.X%` with threshold coloring
- **Byte metrics** (name contains "bytes"): Human-readable (B, KB, MB, GB, TB)
- **Other metrics**: `XX.XX` (2 decimal places)

**Disk Labels:** For disk metrics, labels show in format: `[gray](mountpoint, device, fstype, method)[white]`
- `mountpoint`: Path where disk is mounted
- `device`: Device name (e.g., sda1)
- `fstype`: Filesystem type (e.g., ext4)
- `method`: Collection method (stats, df, or du)

**Update Mechanism:**
- Called via `App.OnMetricsCollected(metrics)` from collector goroutine
- Data copied to internal `currentMetrics` with read-write lock
- `render()` formats and updates view text
- `app.Draw()` called only if running

### History View

**Title:** `Aggregated History (1/5/0 for window, Tab to switch, Q to quit)`

**Features:**
- Dynamic colors enabled
- Scrollable content
- No text wrapping
- Filterable by window size (1m default, 5m, 1h)

**Data Management:**
- Maximum 1000 recent entries in memory
- Older entries pruned when limit exceeded
- Sorted by timestamp descending (most recent first)
- Display limited to 50 results per screen

**Table Format:**

```
Window: 1m

TIME         RESOURCE        METRIC     TYPE     VALUE
──────────────────────────────────────────────────────────
09:38:56     cpu             usage_per… avg      45.5%
09:38:56     disk            usage_per… avg      78.5%
09:37:56     memory          used_bytes avg      10.0 GB

Showing 3 of 1000 results
```

**Column Specifications:**

| Column | Width | Format | Notes |
|--------|-------|--------|-------|
| TIME | 12 chars | `HH:MM:SS` | Left-aligned |
| RESOURCE | 15 chars | Text | cpu, memory, disk |
| METRIC | 10 chars | Text | Truncated with "…" |
| TYPE | 8 chars | Text | avg, max, min, last |
| VALUE | Remaining | Formatted | Percentage, bytes, or numeric |

**Row Styling:**
- Header row: `[blue]`
- Separator: em-dashes (70 chars)
- Data rows: white text
- Footer: `Showing N of M results`

**Update Mechanism:**
- Called via `App.OnAggregationComplete(results)` from aggregator goroutine
- Results appended to internal `results` slice with read-write lock
- Memory trimmed to 1000 most recent entries
- `render()` applies current filters and updates view text

### Status Bar

**Position:** Bottom of screen (fixed 1 line)

**Features:**
- Center-aligned text
- Dynamic color support
- Thread-safe updates with `sync.RWMutex`

**Status States:**

| State | Display |
|-------|---------|
| Initializing | `[green]Initializing...[white] \| [gray]etlmon[white]` |
| Running | `[green]Running[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` |
| Collecting | `[green]Collecting metrics[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` |
| Aggregating | `[green]Aggregating data[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` |
| Persisting | `[green]Persisting to database[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` |

**Update Methods:**
- `SetStatus(status string)` - Changes status text
- `SetLastUpdate(time.Time)` - Updates timestamp
- Both trigger `render()` for display update

## Value Formatting

### Bytes

Converts numeric byte values to human-readable format (unit size = 1024 bytes):

```
< 1 KB       →  "X B"       (e.g., "500 B")
< 1 MB       →  "X.X KB"    (e.g., "512.5 KB")
< 1 GB       →  "X.X MB"    (e.g., "25.8 MB")
< 1 TB       →  "X.X GB"    (e.g., "100.5 GB")
>= 1 TB      →  "X.X TB"    (e.g., "1.2 TB")
```

Function: `FormatBytes(bytes float64) string` (internal/tui/app.go)

### Percentages

Format: `XX.X%` (1 decimal place precision)

Applied to any metric with "percent" in the name. Threshold-based coloring applied in Real-time view only.

### Duration

Converts time durations to abbreviated format:

```
>= 1 hour    →  "Xh"   (e.g., "2h")
>= 1 minute  →  "Xm"   (e.g., "5m")
< 1 minute   →  "Xs"   (e.g., "30s")
```

Function: `FormatDuration(d time.Duration) string` (internal/tui/app.go)

## Thread Safety

All TUI components use `sync.RWMutex` for concurrent access:

**Lock Patterns:**

| Operation | Lock Type | Pattern |
|-----------|-----------|---------|
| Reading shared data | Read lock | `RLock()` / `RUnlock()` |
| Writing shared data | Write lock | `Lock()` / `Unlock()` |
| Updating text view | Write lock | `Lock()` / `Unlock()` |

**Key Guideline:** Always acquire locks in the same order across codebase to prevent deadlocks. All components follow: operation → lock → unlock.

**Safe Update Pattern:**
1. Acquire write lock
2. Copy/update internal state
3. Release lock
4. Call `render()` (acquires its own locks)
5. Update text view (within write lock)

**tview Threading:**
- tview's Application is NOT fully thread-safe
- All UI operations must happen in tview event loop
- Use `app.Draw()` only after verifying running state

## Data Flow

Simplified flow from collection through TUI rendering:

```
1. COLLECTOR GOROUTINE → Collects metrics at interval
                      ↓
2. APP.OnMetricsCollected() → Receives metrics
                             ├─ realtimeView.Update(metrics)
                             ├─ statusBar.SetLastUpdate(time.Now())
                             └─ app.Draw() if running

3. REALTIME VIEW UPDATE → Updates internal state
                        ├─ Acquires write lock
                        ├─ Copies metric data
                        ├─ Calls render()
                        └─ Updates tview.TextView

4. AGGREGATOR GOROUTINE (parallel) → Applies aggregations
                                    ↓
5. APP.OnAggregationComplete() → Receives results
                               ├─ historyView.Update(results)
                               └─ app.Draw() if running

6. HISTORY VIEW UPDATE → Updates internal state
                       ├─ Acquires write lock
                       ├─ Appends new results
                       ├─ Trims to 1000 entries
                       ├─ Calls render() with filters
                       └─ Updates tview.TextView
```

**Concurrency:** Collector, aggregator, and TUI event loop run in separate goroutines with mutex protection for shared memory.

## Testing Captures

The `internal/tui/capture_test.go` file generates realistic UI output samples for documentation and regression testing.

**Generated Files** (in `docs/ui/`):

| File | Content |
|------|---------|
| `realtime_view.txt` | Sample Real-time View with 11 metrics across 3 resource types |
| `history_view.txt` | Sample History View with 1-minute window filter (19 results) |
| `status_bar.txt` | Status bar in 5 different states |

**Regenerate Captures:**

```bash
go test ./internal/tui -run TestCaptureTUIOutput -v
```

**Test Data:**
- **RealtimeView:** 11 metrics (CPU: 3, Memory: 4, Disk: 4)
- **HistoryView:** 19 aggregation results (multiple time windows)
- **StatusBar:** 5 status states

## Architecture Integration

### Connection Points

**Collector Callback:**
```go
func (a *App) OnMetricsCollected(metrics []collector.Metric) {
    a.realtimeView.Update(metrics)
    a.statusBar.SetLastUpdate(time.Now())
    if running { a.app.Draw() }
}
```

**Aggregator Callback:**
```go
func (a *App) OnAggregationComplete(results []aggregator.AggregationResult) {
    a.historyView.Update(results)
    if running { a.app.Draw() }
}
```

**Storage Interface:**
- `GetMetrics(opts)` - Load historical data for History view
- `GetLatestMetrics(resourceType, windowSize)` - Load recent metrics

### Module Structure

```
internal/tui/
├── app.go          # App struct, layout, keybindings, formatting
├── realtime.go     # RealtimeView component
├── history.go      # HistoryView component
├── status.go       # StatusBar component
└── capture_test.go # UI capture tests
```

## Common Pitfalls

1. **Not Checking Running State:** Always verify `running` flag before `app.Draw()` to avoid panics during shutdown.

2. **Forgetting Read Locks:** Reading shared state requires locks to prevent data races.

3. **Lock Ordering:** Always acquire locks in the same order across codebase to prevent deadlocks.

4. **Blocking Operations in Render:** `render()` should only format text, not perform I/O.

5. **Color Syntax Errors:** Always close with `[white]` after colored sections; test in target terminal.

## Troubleshooting

**Colors not showing:**
- Verify terminal supports 256-color mode
- Check `$TERM` env variable (should be `xterm-256color` or similar)
- Try different terminal emulator (kitty, alacritty, iterm2 have best support)

**Text misaligned:**
- Ensure terminal window is at least 80 characters wide
- Check for unsupported Unicode characters if using custom fonts

**Updates not appearing:**
- Confirm collector goroutine is running and reporting metrics
- Check for panics in logs
- Verify mutex isn't deadlocked

**High CPU usage:**
- Reduce collection interval in config
- Check if `render()` is being called excessively
- Profile with `pprof`: use `-cpuprofile` flag

**Memory growth:**
- History view limited to 1000 results; should be stable
- Monitor goroutine count via `debug/pprof`
