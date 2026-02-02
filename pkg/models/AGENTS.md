<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# models

## Purpose

Shared data structures used by both Node (API responses) and UI (API client). These models represent the JSON contract between components.

## Key Files

| File | Description |
|------|-------------|
| `fs.go` | Filesystem usage models |
| `paths.go` | Path statistics models |
| `logs.go` | Log line models |
| `processes.go` | Process statistics models |
| `cron.go` | Cron job models |
| `xferlog.go` | Xferlog entry models |
| `api.go` | Generic API response wrappers |

## For AI Agents

### Working In This Directory

- Models should match database schema closely
- Use JSON tags for API serialization
- Keep models simple (data only, no methods)
- Document field meanings in comments

### Model Definitions

```go
// fs.go
type FilesystemUsage struct {
    MountPoint  string    `json:"mount_point"`
    TotalBytes  uint64    `json:"total_bytes"`
    UsedBytes   uint64    `json:"used_bytes"`
    AvailBytes  uint64    `json:"avail_bytes"`
    UsedPercent float64   `json:"used_percent"`
    CollectedAt time.Time `json:"collected_at"`
}

// paths.go
type PathStats struct {
    Path           string    `json:"path"`
    FileCount      int64     `json:"file_count"`
    DirCount       int64     `json:"dir_count"`
    ScanDurationMs int64     `json:"scan_duration_ms"`
    Status         string    `json:"status"` // OK, SCANNING, ERROR
    ErrorMessage   string    `json:"error_message,omitempty"`
    CollectedAt    time.Time `json:"collected_at"`
}

// logs.go
type LogLine struct {
    ID          int64     `json:"id"`
    LogName     string    `json:"log_name"`
    LogPath     string    `json:"log_path"`
    Line        string    `json:"line"`
    CreatedAt   time.Time `json:"created_at"`
    SourceInode uint64    `json:"source_inode,omitempty"`
}

// processes.go
type ProcessStats struct {
    PID         int       `json:"pid"`
    ProcessName string    `json:"process_name"`
    Cmdline     string    `json:"cmdline"`
    StartTime   time.Time `json:"start_time"`
    RuntimeSec  int64     `json:"runtime_sec"`
    CPUPercent  float64   `json:"cpu_percent"`
    MemRSSBytes int64     `json:"mem_rss_bytes"`
    Status      string    `json:"status"` // RUNNING, STOPPED
    CollectedAt time.Time `json:"collected_at"`
}

// cron.go
type CronJob struct {
    JobID       int64     `json:"job_id"`
    Schedule    string    `json:"schedule"`
    Command     string    `json:"command"`
    User        string    `json:"user"`
    Source      string    `json:"source"` // system, user
    NextRun     time.Time `json:"next_run"`
    LastChecked time.Time `json:"last_checked"`
}

// xferlog.go
type XferlogEntry struct {
    ID              int64     `json:"id"`
    LogTime         time.Time `json:"log_time"`
    RemoteHost      string    `json:"remote_host"`
    Username        string    `json:"username"`
    Filename        string    `json:"filename"`
    Bytes           int64     `json:"bytes"`
    TransferTimeSec int       `json:"transfer_time_sec"`
    Direction       string    `json:"direction"` // upload, download
    Status          string    `json:"status"`
}
```

### API Response Wrappers

```go
// api.go

// Response is the standard API response wrapper
type Response struct {
    Data interface{} `json:"data"`
    Meta *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination info
type Meta struct {
    Total  int `json:"total,omitempty"`
    Limit  int `json:"limit,omitempty"`
    Offset int `json:"offset,omitempty"`
}

// ErrorResponse is returned on API errors
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}
```

<!-- MANUAL: -->
