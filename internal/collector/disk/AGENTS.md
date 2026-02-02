<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# disk

## Purpose

Collects filesystem usage metrics for all mounted filesystems. Writes to `filesystem_usage` table.

## Key Files

| File | Description |
|------|-------------|
| `collector.go` | DiskCollector implementation |
| `collector_test.go` | Unit tests |

## For AI Agents

### Working In This Directory

- Use `syscall.Statfs()` for getting filesystem stats
- Filter out pseudo-filesystems (proc, sys, dev, etc.)
- Calculate used_percent as `(total - avail) / total * 100`
- Default collection interval: 15 seconds

### Data Model

```go
type FilesystemUsage struct {
    MountPoint  string
    TotalBytes  uint64
    UsedBytes   uint64
    AvailBytes  uint64
    UsedPercent float64
    CollectedAt time.Time
}
```

### Implementation Notes

```go
// Get mount points from /proc/mounts
func getMountPoints() ([]string, error)

// Get stats for a single mount point
func getFilesystemStats(mountPoint string) (*FilesystemUsage, error) {
    var stat syscall.Statfs_t
    if err := syscall.Statfs(mountPoint, &stat); err != nil {
        return nil, err
    }
    // Calculate bytes from blocks
    total := stat.Blocks * uint64(stat.Bsize)
    avail := stat.Bavail * uint64(stat.Bsize)
    used := total - avail
    // ...
}
```

### Excluded Filesystems

Skip these filesystem types:
- `proc`, `sysfs`, `devtmpfs`, `devpts`
- `tmpfs` (optional, configurable)
- `squashfs`, `overlay` (container-related)

<!-- MANUAL: -->
