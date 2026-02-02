<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# process

## Purpose

Monitors configured processes (by name pattern or PID). Collects CPU%, memory RSS, runtime. Writes to `process_stats` table. Supports kill command via controller.

## Key Files

| File | Description |
|------|-------------|
| `collector.go` | ProcessCollector implementation |
| `collector_test.go` | Unit tests |
| `proc.go` | /proc filesystem parsing |
| `cpu.go` | CPU percentage calculation |

## For AI Agents

### Working In This Directory

- Read from `/proc/[pid]/stat`, `/proc/[pid]/status`, `/proc/[pid]/cmdline`
- CPU% requires delta calculation between samples
- Track process existence (restart detection)
- Default collection interval: 5 seconds

### Data Model

```go
type ProcessStats struct {
    PID         int
    ProcessName string
    Cmdline     string
    StartTime   time.Time
    RuntimeSec  int64
    CPUPercent  float64
    MemRSSBytes int64
    Status      string // "RUNNING", "STOPPED"
    CollectedAt time.Time
}

type ProcessWatch struct {
    Name  string // Process name to match
    Match string // Cmdline pattern (regex)
    PID   int    // Specific PID (optional)
}
```

### CPU Calculation

```go
// CPU% = (process_cpu_time_delta / total_cpu_time_delta) * 100 * num_cpus

type cpuSample struct {
    processTime uint64 // utime + stime from /proc/[pid]/stat
    totalTime   uint64 // sum of all fields from /proc/stat
    timestamp   time.Time
}

func calculateCPUPercent(prev, curr cpuSample) float64 {
    processDelta := curr.processTime - prev.processTime
    totalDelta := curr.totalTime - prev.totalTime
    if totalDelta == 0 {
        return 0
    }
    return float64(processDelta) / float64(totalDelta) * 100 * float64(numCPU)
}
```

### Process Discovery

```go
// Find processes matching config
func (c *Collector) findProcesses() ([]int, error) {
    entries, _ := os.ReadDir("/proc")
    var pids []int
    for _, entry := range entries {
        if pid, err := strconv.Atoi(entry.Name()); err == nil {
            if c.matchesConfig(pid) {
                pids = append(pids, pid)
            }
        }
    }
    return pids, nil
}
```

### Kill Support

Kill is handled by `internal/controller`, not this collector. This collector only monitors and reports.

<!-- MANUAL: -->
