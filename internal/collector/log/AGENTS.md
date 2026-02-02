<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# log

## Purpose

Tails log files in real-time using fsnotify. Handles logrotate (inode changes, truncation). Writes new lines to `log_lines` table.

## Key Files

| File | Description |
|------|-------------|
| `tailer.go` | LogTailer implementation |
| `tailer_test.go` | Unit tests |
| `rotate.go` | Logrotate detection and handling |

## For AI Agents

### Working In This Directory

- Use `fsnotify` for efficient file watching
- Track file offset for resuming after restart
- Detect logrotate via inode change or truncation
- Batch insert new lines for efficiency
- Configurable: start from end vs. beginning

### Data Model

```go
type LogLine struct {
    ID          int64
    LogName     string
    LogPath     string
    Line        string
    CreatedAt   time.Time
    SourceInode uint64 // For rotate tracking
}

type LogConfig struct {
    Name        string
    Path        string
    Follow      bool
    BufferLines int // Max lines to buffer before insert
}

type TailerState struct {
    Path   string
    Offset int64
    Inode  uint64
}
```

### Logrotate Handling

```go
// Detect rotation scenarios:
// 1. Inode changed -> reopen new file
// 2. File size decreased -> truncation, reset offset
// 3. File deleted -> wait for recreation

func (t *Tailer) checkRotation() (rotated bool, err error) {
    stat, err := os.Stat(t.path)
    if err != nil {
        return false, err
    }

    currentInode := getInode(stat)
    if currentInode != t.lastInode {
        return true, nil // File rotated
    }

    if stat.Size() < t.offset {
        t.offset = 0 // File truncated
    }
    return false, nil
}
```

### Implementation Pattern

```go
func (t *Tailer) Start(ctx context.Context) error {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(filepath.Dir(t.path)) // Watch directory

    for {
        select {
        case <-ctx.Done():
            return nil
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write != 0 {
                t.readNewLines()
            }
        }
    }
}
```

<!-- MANUAL: -->
