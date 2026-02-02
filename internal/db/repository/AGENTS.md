<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# repository

## Purpose

Data access layer with typed interfaces for each domain. Separates SQL queries from business logic.

## Key Files

| File | Description |
|------|-------------|
| `repository.go` | Main repository aggregating all sub-repositories |
| `fs.go` | Filesystem usage queries |
| `paths.go` | Path statistics queries |
| `logs.go` | Log lines queries |
| `processes.go` | Process statistics queries |
| `cron.go` | Cron jobs queries |
| `xferlog.go` | Xferlog entries queries |

## For AI Agents

### Working In This Directory

- Each domain has its own repository file
- Use prepared statements for performance
- Queries use `context.Context` for cancellation
- Write methods for collectors, read methods for API

### Repository Structure

```go
// Main repository aggregates all sub-repositories
type Repository struct {
    FS       *FSRepository
    Paths    *PathsRepository
    Logs     *LogsRepository
    Procs    *ProcessRepository
    Cron     *CronRepository
    Xferlog  *XferlogRepository
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{
        FS:      NewFSRepository(db),
        Paths:   NewPathsRepository(db),
        Logs:    NewLogsRepository(db),
        Procs:   NewProcessRepository(db),
        Cron:    NewCronRepository(db),
        Xferlog: NewXferlogRepository(db),
    }
}
```

### Repository Pattern

```go
type FSRepository struct {
    db          *sql.DB
    stmtInsert  *sql.Stmt
    stmtGetAll  *sql.Stmt
}

func NewFSRepository(db *sql.DB) *FSRepository {
    r := &FSRepository{db: db}
    r.stmtInsert, _ = db.Prepare(`
        INSERT OR REPLACE INTO filesystem_usage
        (mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `)
    r.stmtGetAll, _ = db.Prepare(`
        SELECT mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at
        FROM filesystem_usage
        ORDER BY mount_point
    `)
    return r
}

// Write (for collectors)
func (r *FSRepository) Save(ctx context.Context, usage *FilesystemUsage) error {
    _, err := r.stmtInsert.ExecContext(ctx,
        usage.MountPoint, usage.TotalBytes, usage.UsedBytes,
        usage.AvailBytes, usage.UsedPercent, usage.CollectedAt)
    return err
}

// Read (for API)
func (r *FSRepository) GetAll(ctx context.Context) ([]*FilesystemUsage, error) {
    rows, err := r.stmtGetAll.QueryContext(ctx)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []*FilesystemUsage
    for rows.Next() {
        u := &FilesystemUsage{}
        rows.Scan(&u.MountPoint, &u.TotalBytes, &u.UsedBytes,
            &u.AvailBytes, &u.UsedPercent, &u.CollectedAt)
        result = append(result, u)
    }
    return result, rows.Err()
}
```

### Pagination Pattern

```go
func (r *LogsRepository) GetLines(ctx context.Context, logName string, limit, offset int) ([]*LogLine, int, error) {
    // Get total count
    var total int
    r.db.QueryRowContext(ctx,
        "SELECT COUNT(*) FROM log_lines WHERE log_name = ?", logName).Scan(&total)

    // Get paginated results
    rows, _ := r.db.QueryContext(ctx, `
        SELECT id, log_name, log_path, line, created_at
        FROM log_lines
        WHERE log_name = ?
        ORDER BY id DESC
        LIMIT ? OFFSET ?
    `, logName, limit, offset)

    // ... scan rows
    return lines, total, nil
}
```

<!-- MANUAL: -->
