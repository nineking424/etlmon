# etlmon UI Guidelines

## Overview

etlmon is a Terminal User Interface (TUI) application built with the [tview](https://github.com/rivo/tview) library for displaying real-time system resource metrics and historical aggregated data. The UI provides two distinct views: real-time metrics and aggregated history, enabling users to monitor system performance at both immediate and longer-term perspectives.

**Key Characteristics:**
- Single-window TUI with modal view switching
- Color-coded status indicators for resource thresholds
- Scrollable views for large datasets
- Responsive updates from background collector and aggregator goroutines
- Thread-safe rendering with minimal latency

## Layout Architecture

The etlmon TUI uses a hierarchical flex-based layout with fixed and proportional sizing:

```
┌────────────────────────────────────────────────────────┐
│  PAGES (tview.Pages) - Proportional (1 unit)           │
│  ┌────────────────────────────────────────────────────┐│
│  │ RealtimeView (default page)                        ││
│  │   - Displays current system metrics                ││
│  │   - Updated on collection intervals                ││
│  │   - Grouped by resource type                       ││
│  │                                                    ││
│  │ OR                                                 ││
│  │                                                    ││
│  │ HistoryView (secondary page)                       ││
│  │   - Displays aggregated historical data            ││
│  │   - Filterable by window size (1m/5m/1h)           ││
│  │   - Table format with metrics                      ││
│  └────────────────────────────────────────────────────┘│
├────────────────────────────────────────────────────────┤
│ StatusBar (fixed 1 line)                               │
│   - Shows current status and last update time          │
│   - Center-aligned                                     │
└────────────────────────────────────────────────────────┘
```

**Layout Implementation:**
- **Root Container:** `tview.Flex` with `FlexRow` direction
- **Main Content Area:** `tview.Pages` (proportional sizing = 0, weight = 1)
- **Status Bar:** Fixed height of 1 line (0 proportion, 0 weight)
- **Page Management:** Two pages added to tview.Pages:
  - `"realtime"` - RealtimeView (initially visible)
  - `"history"` - HistoryView (hidden by default)

### Component Sizing

| Component | Size | Proportional | Weight | Behavior |
|-----------|------|--------------|--------|----------|
| Pages | Dynamic | 0 | 1 | Expands to fill available space |
| StatusBar | 1 line | 0 | 0 | Fixed height, no expansion |

## Keyboard Shortcuts

Users can interact with the TUI using the following keyboard shortcuts:

| Key | Action | Scope |
|-----|--------|-------|
| `Tab` | Toggle between Real-time and History views | Global |
| `R` | Switch to Real-time view | Global |
| `H` | Switch to History view | Global |
| `1` | Filter history by 1-minute window | History view |
| `5` | Filter history by 5-minute window | History view |
| `0` | Filter history by 1-hour window | History view |
| `Q` | Quit application | Global |
| `Esc` | Quit application | Global |
| `Ctrl+C` | Quit application | Global |

**Implementation Notes:**
- Keybindings are configured in `App.setupKeybindings()` using `SetInputCapture()`
- Both `KeyEscape` and `tcell` raw rune events are captured
- Tab toggles the view state in the internal `currentView` variable
- Quit commands cleanly stop the application with `app.Stop()`

## Color Scheme

The etlmon TUI uses a consistent color palette for visual clarity and status communication. All colors are specified using tview's color markup syntax `[colorname]text[white]`.

### Theme Colors

| Color | Usage | Context |
|-------|-------|---------|
| **Green** `[green]` | Headers, normal values, positive status | Resource type headers (CPU, DISK, MEMORY), section dividers, "Running" status |
| **Yellow** `[yellow]` | Warning state, elevated values | Percentage metrics in 70-90% range, waiting messages, caution indicators |
| **Red** `[red]` | Critical state, high values | Percentage metrics >90%, error conditions, critical alerts |
| **Blue** `[blue]` | Information, labels | Column headers in History view, window filter labels |
| **Gray** `[gray]` | Secondary info, metadata, timestamps | Mount points, labels, secondary text, footer information |
| **White** `[white]` | Default text, reset color | Body text, metric values, general content |

### Threshold-Based Coloring for Percentage Metrics

Percentage-based metrics (e.g., CPU usage, memory usage, disk usage) are dynamically colored based on the value:

```
Value < 70%  →  [green]XX.X%[white]   (Normal state)
Value 70-90% →  [yellow]XX.X%[white]  (Warning state)
Value > 90%  →  [red]XX.X%[white]     (Critical state)
```

**Implementation:** The coloring logic is applied during formatting in `RealtimeView.formatMetrics()` by checking metric name patterns and numeric values.

### Color Compatibility

- **Terminal Support:** Colors are rendered via tcell, which supports most modern terminal emulators
- **Fallback:** tview handles color downgrades gracefully for limited color terminals
- **Testing:** The `[gray]` color is used selectively for semantic markup; it may appear differently depending on terminal background

## Components

### 5.1 Real-time View

The Real-time View displays current system metrics collected at configurable intervals. It updates immediately as new metrics are collected.

**Component:** `RealtimeView` (internal/tui/realtime.go)

**Title:** ` Real-time Metrics (Press Tab to switch, Q to quit) `

**Features:**
- Dynamic color rendering enabled (`SetDynamicColors(true)`)
- Scrollable content area (`SetScrollable(true)`)
- No text wrapping (`SetWrap(false)`)
- Border enabled with left-aligned title
- Thread-safe updates with `sync.RWMutex`

**Data Organization:**
- Metrics grouped by `ResourceType` (CPU, DISK, MEMORY, etc.)
- Resource types sorted alphabetically
- Metrics within each group sorted alphabetically by name
- Organized in collapsible sections with visual dividers

**Visual Rendering:**
```
━━━ CPU ━━━
  core0_usage_percent  30.2%
  core1_usage_percent  55.8%
  usage_percent        45.5%

━━━ DISK ━━━
  free_bytes           100.5 GB (/)
  total_bytes          465.6 GB (/)
  usage_percent        78.5% (/)
  used_bytes           365.1 GB (/)

━━━ MEMORY ━━━
  available_bytes      6.0 GB
  total_bytes          16.0 GB
  usage_percent        62.3%
  used_bytes           10.0 GB

Last updated: 09:38:56
```

**Value Formatting Rules:**
- **Percentage metrics** (name contains "percent"): `XX.X%` with threshold-based coloring
- **Byte metrics** (name contains "bytes"): Human-readable format (B, KB, MB, GB, TB)
- **Other metrics**: Fixed 2 decimal places (`XX.XX`)

**Labels:** Mount points and other labels are displayed in gray parentheses: `[gray](mountpoint)[white]`

**Update Mechanism:**
- Called via `App.OnMetricsCollected(metrics)` from collector goroutine
- Data copied to internal `currentMetrics` slice with read-write lock
- `render()` method formats and updates view text
- UI drawn only when app is running (`app.Draw()`)

### 5.2 History View

The History View displays aggregated historical metrics in a table format with filtering capabilities.

**Component:** `HistoryView` (internal/tui/history.go)

**Title:** ` Aggregated History (1/5/0 for window, Tab to switch, Q to quit) `

**Features:**
- Dynamic color rendering enabled (`SetDynamicColors(true)`)
- Scrollable content area (`SetScrollable(true)`)
- No text wrapping (`SetWrap(false)`)
- Border enabled with left-aligned title
- Thread-safe updates with `sync.RWMutex`

**Filtering:**
- **Window Size Filter:** 1-minute (`1m`), 5-minute (`5m`), or 1-hour (`1h`)
  - Triggered by keys `1`, `5`, `0` respectively
  - Stored in `windowFilter` variable
  - Default filter: `1m`
- **Resource Type Filter:** Optional filtering by resource type (CPU, DISK, MEMORY)
  - Not exposed in default keybindings
  - Programmatically settable via `SetResourceFilter()`

**Data Management:**
- Results accumulated in memory with maximum of 1000 recent entries
- Older entries automatically pruned when limit exceeded
- Sorted by timestamp descending (most recent first)
- Limited display to 50 results per screen

**Table Format:**

```
Window: 1m

TIME         RESOURCE        METRIC     TYPE     VALUE
──────────────────────────────────────────────────────────────────────
09:38:56     cpu             usage_per… avg      45.5%
09:38:56     cpu             usage_per… max      58.2%
09:38:56     cpu             usage_per… min      32.1%
09:38:56     cpu             usage_per… last     47.3%
09:38:56     memory          usage_per… avg      62.3%
09:38:56     memory          usage_per… max      65.8%
09:38:56     memory          used_bytes avg      10.0 GB
09:38:56     disk            usage_per… avg      78.5%
09:38:56     disk            usage_per… max      79.2%
09:38:56     disk            free_bytes avg      100.5 GB
09:37:56     cpu             usage_per… avg      42.1%
09:37:56     memory          usage_per… avg      61.5%
09:37:56     disk            usage_per… avg      78.3%

Showing 13 of 13 results
```

**Column Specifications:**

| Column | Width | Format | Notes |
|--------|-------|--------|-------|
| TIME | 12 chars | `HH:MM:SS` | Left-aligned, timestamp only (no date) |
| RESOURCE | 15 chars | Text | Resource type (cpu, memory, disk) |
| METRIC | 10 chars | Text | Metric name, truncated with "…" if exceeds width |
| TYPE | 8 chars | Text | Aggregation type (avg, max, min, last) |
| VALUE | Remaining | Formatted | Percentage, bytes, or numeric |

**Row Styling:**
- Header row colored blue `[blue]`
- Separator line using em-dashes (70 characters total)
- Data rows in default white text
- Footer shows count: `Showing N of M results`

**Value Formatting:**
- **Percentage metrics:** `XX.X%` (no coloring in table)
- **Byte metrics:** Human-readable format (B, KB, MB, GB, TB)
- **Other metrics:** Fixed 2 decimal places

**Update Mechanism:**
- Called via `App.OnAggregationComplete(results)` from aggregator goroutine
- New results appended to internal `results` slice with read-write lock
- Memory trimmed to 1000 most recent entries
- `render()` applies current filters and updates view text
- UI drawn only when app is running

### 5.3 Status Bar

The Status Bar is a fixed single-line component at the bottom of the window displaying application status and timing information.

**Component:** `StatusBar` (internal/tui/status.go)

**Position:** Bottom of screen

**Features:**
- Center-aligned text (`SetTextAlign(tview.AlignCenter)`)
- Dynamic color support enabled
- Thread-safe updates with `sync.RWMutex`

**Status States and Format:**

| State | Display | Trigger |
|-------|---------|---------|
| Initializing | `[green]Initializing...[white] | [gray]etlmon[white]` | Application startup |
| Running | `[green]Running[white] | Last: HH:MM:SS \| [gray]etlmon[white]` | `SetStatus("Running")` after app starts |
| Collecting | `[green]Collecting metrics[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` | During metric collection |
| Aggregating | `[green]Aggregating data[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` | During aggregation phase |
| Persisting | `[green]Persisting to database[white] \| Last: HH:MM:SS \| [gray]etlmon[white]` | During database persistence |

**Visual Examples:**

```
=== Initializing ===
[empty - status not set initially]

=== Running ===
Running | Last: 09:38:56 | etlmon

=== Collecting ===
Collecting metrics | Last: 09:38:56 | etlmon

=== Aggregating ===
Aggregating data | Last: 09:38:56 | etlmon

=== Persisting ===
Persisting to database | Last: 09:38:56 | etlmon
```

**Update Mechanism:**
- Status text changed via `SetStatus(status string)`
- Last update timestamp set via `SetLastUpdate(time.Time)`
- Both methods trigger `render()` to update display
- Called from collector and aggregator goroutines via `app.OnMetricsCollected()` and `app.OnAggregationComplete()`

## Value Formatting

Consistent value formatting is critical for readability and understanding metric units.

### Bytes Formatting

Function: `FormatBytes(bytes float64) string` (internal/tui/app.go)

Converts numeric byte values to human-readable format:

```
< 1 KB       →  "X B"           (e.g., "500 B")
< 1 MB       →  "X.X KB"        (e.g., "512.5 KB")
< 1 GB       →  "X.X MB"        (e.g., "25.8 MB")
< 1 TB       →  "X.X GB"        (e.g., "100.5 GB")
>= 1 TB      →  "X.X TB"        (e.g., "1.2 TB")
```

**Implementation:**
- Unit size = 1024 bytes (binary measurement)
- Divisor multiplied for each magnitude level
- Result formatted with 1 decimal place for KB and above
- Zero decimals for bytes below 1 KB
- Uses `KMGTPE` exponent characters for suffix

**Usage:** Applied to any metric with "bytes" in the name (available_bytes, used_bytes, free_bytes, total_bytes)

### Percentage Formatting

Format: `XX.X%` (1 decimal place precision)

**Threshold-Based Coloring:**
```go
if value > 90 {
    return fmt.Sprintf("[red]%.1f%%[white]", value)
} else if value > 70 {
    return fmt.Sprintf("[yellow]%.1f%%[white]", value)
} else {
    return fmt.Sprintf("[green]%.1f%%[white]", value)
}
```

**Thresholds:**
- **< 70%:** Green (normal operation)
- **70-90%:** Yellow (elevated, monitoring recommended)
- **> 90%:** Red (critical, action may be needed)

**Usage:** Applied to any metric with "percent" in the name (usage_percent, available_percent, etc.)

### Duration Formatting

Function: `FormatDuration(d time.Duration) string` (internal/tui/app.go)

Converts time durations to abbreviated format:

```
>= 1 hour    →  "Xh"            (e.g., "2h")
>= 1 minute  →  "Xm"            (e.g., "5m")
< 1 minute   →  "Xs"            (e.g., "30s")
```

**Implementation:**
- No decimal places, integer conversion
- Cascading comparison from largest unit downward
- Useful for window size display in History view

**Usage:** Applied to `WindowSize` duration fields when displayed in History view

### Generic Numeric Formatting

Format: `XX.XX` (2 decimal places)

Used for metrics that don't match percentage or byte patterns.

## Thread Safety

All TUI components employ thread-safe patterns to handle concurrent updates from background goroutines.

### Synchronization Mechanisms

**Component-Level Locking:**
Each major component (RealtimeView, HistoryView, StatusBar, App) uses `sync.RWMutex`:

```go
type RealtimeView struct {
    view           *tview.TextView
    currentMetrics []collector.Metric
    mu             sync.RWMutex  // Protects currentMetrics and view.SetText()
}
```

**Lock Patterns:**

| Operation | Lock Type | Code Pattern |
|-----------|-----------|--------------|
| Reading metrics | Read lock | `v.mu.RLock()` / `defer v.mu.RUnlock()` |
| Writing metrics | Write lock | `v.mu.Lock()` / `v.mu.Unlock()` |
| Updating text view | Write lock | `v.mu.Lock()` / `defer v.mu.Unlock()` |
| Getting text (safe read) | Read lock | `v.mu.RLock()` / `defer v.mu.RUnlock()` |

**Safe Update Pattern:**

```go
// Update from background goroutine
func (v *RealtimeView) Update(metrics []collector.Metric) {
    v.mu.Lock()
    v.currentMetrics = make([]collector.Metric, len(metrics))
    copy(v.currentMetrics, metrics)
    v.mu.Unlock()

    v.render()  // render() acquires its own locks as needed
}

// Render with proper locking
func (v *RealtimeView) render() {
    v.mu.RLock()
    metrics := make([]collector.Metric, len(v.currentMetrics))
    copy(metrics, v.currentMetrics)
    v.mu.RUnlock()

    text := v.formatMetrics(metrics)

    v.mu.Lock()
    v.view.SetText(text)
    v.mu.Unlock()
}
```

### Application-Level Coordination

The `App` struct manages view state and coordinates updates:

```go
type App struct {
    app          *tview.Application
    pages        *tview.Pages
    realtimeView *RealtimeView
    historyView  *HistoryView
    statusBar    *StatusBar
    currentView  ViewMode
    store        Store
    running      bool
    mu           sync.RWMutex  // Protects running, currentView, store
}
```

**Running State Check:**
```go
func (a *App) OnMetricsCollected(metrics []collector.Metric) {
    a.realtimeView.Update(metrics)
    a.statusBar.SetLastUpdate(time.Now())

    // Only draw if app is running to avoid nil pointer panics
    a.mu.RLock()
    running := a.running
    a.mu.RUnlock()

    if running {
        a.app.Draw()
    }
}
```

### Safe UI Updates from Goroutines

**tview's Threading Model:**
- tview's `Application` is NOT fully thread-safe
- All UI operations must happen in the tview event loop
- Use `QueueUpdateDraw()` for updates from background goroutines

**When NOT using `QueueUpdateDraw()`:**
- Only perform safe operations that don't interact with tview's event loop
- Copy data to internal structs (under mutex)
- Call `render()` which updates internal text
- Only call `app.Draw()` with running state check

**Best Practice:**
Use `App.QueueUpdateDraw(func())` for safety:
```go
a.app.QueueUpdateDraw(func() {
    a.realtimeView.Update(metrics)
    a.statusBar.SetLastUpdate(time.Now())
})
```

## Data Flow

The data flow from metric collection through TUI rendering follows this sequence:

```
1. COLLECTOR GOROUTINE
   ├─ Runs at configured interval (e.g., 10s)
   ├─ Collects CPU, memory, disk metrics
   └─ Calls app.OnMetricsCollected(metrics [])

2. APP EVENT HANDLER
   ├─ Receives metrics from collector
   ├─ Calls realtimeView.Update(metrics)
   ├─ Updates statusBar.SetLastUpdate(time.Now())
   └─ Calls app.Draw() if running

3. REALTIME VIEW UPDATE
   ├─ Acquires write lock on currentMetrics
   ├─ Copies metric data
   ├─ Releases lock
   ├─ Calls render()
   └─ Updates tview.TextView with formatted text

4. AGGREGATOR GOROUTINE (Parallel)
   ├─ Maintains time windows (1m, 5m, 1h)
   ├─ Applies aggregation functions (avg, max, min, last)
   └─ Calls app.OnAggregationComplete(results)

5. APP EVENT HANDLER
   ├─ Receives aggregation results
   ├─ Calls historyView.Update(results)
   └─ Calls app.Draw() if running

6. HISTORY VIEW UPDATE
   ├─ Acquires write lock on results
   ├─ Appends new results
   ├─ Trims to 1000 most recent
   ├─ Releases lock
   ├─ Calls render() with current filters
   └─ Updates tview.TextView with formatted table

7. USER INTERACTION
   ├─ Keyboard event captured in setupKeybindings()
   ├─ View switches or filters applied
   └─ Changes reflected on next draw cycle
```

**Concurrency Model:**
- Collector, aggregator, and TUI event loop run in separate goroutines
- No shared memory without mutex protection
- Minimal lock contention through read-lock usage in render functions
- UI updates batched where possible to reduce redraw frequency

## Testing Captures

The `internal/tui/capture_test.go` file contains tests that generate realistic UI output samples. These captures serve as documentation and regression tests.

### Generated Files

Output files are generated in `docs/ui/`:

| File | Content | Regeneration |
|------|---------|--------------|
| `realtime_view.txt` | Sample Real-time View with mixed metrics | `go test ./internal/tui -run TestCaptureTUIOutput` |
| `history_view.txt` | Sample History View with 1-minute window filter | `go test ./internal/tui -run TestCaptureTUIOutput` |
| `status_bar.txt` | Status bar in multiple states | `go test ./internal/tui -run TestCaptureTUIOutput` |

### Regenerating Captures

To regenerate UI captures after UI changes:

```bash
cd /Users/nineking/workspace/app/etlmon
go test ./internal/tui -run TestCaptureTUIOutput -v
```

**Test Output:**
```
=== RUN   TestCaptureTUIOutput
=== RUN   TestCaptureTUIOutput/RealtimeView
--- PASS: TestCaptureTUIOutput (0.003s)
    --- PASS: TestCaptureTUIOutput/RealtimeView (0.001s)
        capture_test.go:122: RealtimeView output written to docs/ui/realtime_view.txt
    --- PASS: TestCaptureTUIOutput/HistoryView (0.001s)
        capture_test.go:298: HistoryView output written to docs/ui/history_view.txt
    --- PASS: TestCaptureTUIOutput/StatusBar (0.002s)
        capture_test.go:340: StatusBar output written to docs/ui/status_bar.txt
    capture_test.go:343: All TUI captures generated successfully!
PASS
```

### Test Data Generation

The test creates realistic sample data:

**RealtimeView:** Creates 11 metrics across 3 resource types
- CPU: 3 metrics (2 cores + aggregate)
- Memory: 4 metrics (available, total, usage, used)
- Disk: 4 metrics (free, total, usage, used with mount points)

**HistoryView:** Creates 19 aggregation results across time windows
- 1-minute window: Latest data with all aggregation types
- 5-minute window: Aggregated data
- 1-hour window: Hourly aggregated data
- Multiple time points for time-series effect

**StatusBar:** Tests 5 different status states
- Initializing (no last update)
- Running
- Collecting metrics
- Aggregating data
- Persisting to database

## Implementation Notes

### Design Patterns

**View Pattern:**
Each view (RealtimeView, HistoryView, StatusBar) encapsulates:
- A `tview` component (TextView, etc.)
- Internal state (metrics, results, status)
- A mutex for thread safety
- Public `Update()` method for data changes
- Private `render()` method for formatting
- A `GetText()` method for testing

**Component Initialization:**
Views are created with:
- Dynamic colors enabled for markup support
- Scrollable for handling variable content
- No word wrapping for clean columnar output
- Borders with instructive titles

**Color Markup:**
- Inline color codes using tview syntax: `[colorname]text[white]`
- Always reset to white after colored text: `[white]`
- Supports nested markup within strings
- tcell handles terminal color translation

### Common Pitfalls

1. **Not Checking Running State:** Always verify `running` flag before calling `app.Draw()` to avoid panics during shutdown.

2. **Forgetting Read Locks:** Reading shared state requires locks to prevent data races detected by `-race` flag.

3. **Lock Ordering:** Always acquire locks in the same order across the codebase to prevent deadlocks (all components follow: operation → lock acquire → unlock).

4. **Blocking Operations in Render:** `render()` should only format text, not perform I/O or long operations.

5. **Color Syntax Errors:** Mismatched color markup can cause display issues:
   - Always close with `[white]` after each colored section
   - Avoid nested inconsistent colors
   - Test colors in target terminal emulator

### Performance Considerations

- **Memory:** History view limited to 1000 results to prevent unbounded growth
- **CPU:** Rendering is O(n) where n is result count, but display limited to 50 rows
- **I/O:** No I/O in render functions; only in update handlers
- **Mutex Contention:** Read locks preferred over write locks; most operations use RLock for render

## Architecture Integration

### Connection Points

The TUI integrates with other etlmon components:

**Collector Interface:**
```go
// Called from collector goroutine
func (a *App) OnMetricsCollected(metrics []collector.Metric) {
    a.realtimeView.Update(metrics)
    a.statusBar.SetLastUpdate(time.Now())
    if running {
        a.app.Draw()
    }
}
```

**Aggregator Interface:**
```go
// Called from aggregator goroutine
func (a *App) OnAggregationComplete(results []aggregator.AggregationResult) {
    a.historyView.Update(results)
    if running {
        a.app.Draw()
    }
}
```

**Storage Interface:**
```go
// Used to load historical data for History view
type Store interface {
    GetMetrics(opts storage.GetMetricsOptions) ([]*storage.AggregatedMetric, error)
    GetLatestMetrics(resourceType, windowSize string) ([]*storage.AggregatedMetric, error)
}
```

### Module Dependencies

```
internal/tui/
├── app.go              # Main App struct, layout, keybindings
├── realtime.go         # RealtimeView component
├── history.go          # HistoryView component
├── status.go           # StatusBar component
└── capture_test.go     # UI capture tests

Dependencies:
├── github.com/rivo/tview          # TUI framework
├── github.com/gdamore/tcell/v2    # Terminal control
├── github.com/etlmon/internal/collector    # Metric types
├── github.com/etlmon/internal/aggregator   # Aggregation result types
└── github.com/etlmon/internal/storage      # Data persistence interface
```

## Future Extensions

Potential enhancements to the TUI:

1. **Filtering UI:** Interactive resource/metric filter selection menu
2. **Export:** Keyboard shortcut to export visible data as CSV
3. **Alerts:** Visual indicators when thresholds are exceeded
4. **Configuration:** In-app settings menu for update intervals and thresholds
5. **Multiple Views:** Tab-based navigation for different resource dashboards
6. **Hotkeys Help:** Overlay showing available commands (? key)
7. **Themes:** User-selectable color schemes (dark, light, high-contrast)
8. **Mouse Support:** Click to sort History view columns or resize panes
9. **Search:** Find metrics or timestamps in History view
10. **Recording:** Capture session to file for later playback/analysis

## Troubleshooting

### Display Issues

**Colors not showing properly:**
- Verify terminal supports 256-color mode
- Check environment: `echo $TERM` (should be `xterm-256color` or similar)
- Try different terminal emulator (kitty, alacritty, iterm2 have best support)

**Text overlapping or misaligned:**
- Ensure terminal window is at least 80 characters wide
- Check for unsupported Unicode characters if using custom fonts
- Verify font metrics in terminal settings

**Updates not appearing:**
- Confirm collector goroutine is running and reporting metrics
- Check for panics in collector/aggregator logs
- Verify mutex isn't deadlocked (add debug logging)

### Performance Issues

**High CPU usage:**
- Reduce collection interval in config
- Check if render() is being called excessively
- Profile with `pprof`: add `-cpuprofile` flag

**Memory growth:**
- History view limited to 1000 results; should be stable
- Check for unclosed file descriptors in collector
- Monitor goroutine count: `debug/pprof`

**Slow updates:**
- Network latency in metrics collection? (shouldn't apply for local only)
- Large number of metrics causing slow format? (limit metrics in config)
- Terminal I/O bottleneck? (try different terminal)
