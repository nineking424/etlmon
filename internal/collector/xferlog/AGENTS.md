<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# xferlog

## Purpose

Parses vsftpd xferlog files to extract FTP transfer records. Writes to `xferlog_entries` table. Supports tailing for new entries.

## Key Files

| File | Description |
|------|-------------|
| `collector.go` | XferlogCollector implementation |
| `collector_test.go` | Unit tests |
| `parser.go` | Xferlog line parser |

## For AI Agents

### Working In This Directory

- Parse vsftpd xferlog format (space-delimited fields)
- Tail log file for new entries (similar to log collector)
- Apply retention/max_rows limit
- Configurable parse start time

### Data Model

```go
type XferlogEntry struct {
    ID              int64
    LogTime         time.Time
    RemoteHost      string
    Username        string
    Filename        string
    Bytes           int64
    TransferTimeSec int
    Direction       string // "upload" or "download"
    Status          string // "c" complete, "i" incomplete
}
```

### Xferlog Format

```
current-time  transfer-time  remote-host  file-size  filename  transfer-type  special-action-flag  direction  access-mode  username  service-name  authentication-method  authenticated-user-id  completion-status

Example:
Mon Jan 15 10:23:45 2024 5 192.168.1.100 1234567 /home/user/file.txt b _ i r user ftp 0 * c
```

### Parser Implementation

```go
func parseXferlogLine(line string) (*XferlogEntry, error) {
    // Fields are space-separated, but datetime has spaces
    // Format: "Www Mmm DD HH:MM:SS YYYY"

    // Extract timestamp (first 24 chars)
    timeStr := line[:24]
    logTime, _ := time.Parse("Mon Jan 2 15:04:05 2006", timeStr)

    // Split remaining fields
    rest := strings.Fields(line[24:])
    // rest[0] = transfer-time
    // rest[1] = remote-host
    // rest[2] = file-size
    // rest[3] = filename
    // rest[4] = transfer-type (a=ascii, b=binary)
    // rest[5] = special-action-flag
    // rest[6] = direction (i=incoming/upload, o=outgoing/download)
    // rest[7] = access-mode
    // rest[8] = username
    // ...
    // rest[13] = completion-status (c=complete, i=incomplete)

    direction := "download"
    if rest[6] == "i" {
        direction = "upload"
    }

    return &XferlogEntry{
        LogTime:         logTime,
        TransferTimeSec: parseInt(rest[0]),
        RemoteHost:      rest[1],
        Bytes:           parseInt64(rest[2]),
        Filename:        rest[3],
        Direction:       direction,
        Username:        rest[8],
        Status:          rest[13],
    }, nil
}
```

### Configuration

```yaml
xferlog:
  path: /var/log/xferlog
  parse_start: "2026-01-01T00:00:00Z"  # Only parse entries after this
  retention_days: 7
  max_rows: 100000
```

<!-- MANUAL: -->
