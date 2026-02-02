<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# schema

## Purpose

SQL migration files for database schema. Uses version-based migrations with a `schema_version` tracking table.

## Key Files

| File | Description |
|------|-------------|
| `001_initial.sql` | Initial schema (all tables) |
| `002_indexes.sql` | Performance indexes |
| `migrations.go` | Migration runner |

## For AI Agents

### Working In This Directory

- Each migration file is numbered sequentially
- Migrations are idempotent (can re-run safely)
- Use `schema_version` table to track applied migrations
- Never modify existing migrations, only add new ones

### Initial Schema (001_initial.sql)

```sql
-- Schema version tracking
CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT
);

INSERT OR IGNORE INTO meta (key, value) VALUES ('schema_version', '0');

-- Filesystem usage
CREATE TABLE IF NOT EXISTS filesystem_usage (
    mount_point TEXT NOT NULL,
    total_bytes INTEGER NOT NULL,
    used_bytes INTEGER NOT NULL,
    avail_bytes INTEGER NOT NULL,
    used_percent REAL NOT NULL,
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (mount_point)
);

-- Path statistics
CREATE TABLE IF NOT EXISTS path_stats (
    path TEXT PRIMARY KEY,
    file_count INTEGER NOT NULL DEFAULT 0,
    dir_count INTEGER NOT NULL DEFAULT 0,
    scan_duration_ms INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'OK',
    error_message TEXT,
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Log lines
CREATE TABLE IF NOT EXISTS log_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_name TEXT NOT NULL,
    log_path TEXT NOT NULL,
    line TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_inode INTEGER
);

-- Process statistics
CREATE TABLE IF NOT EXISTS process_stats (
    pid INTEGER PRIMARY KEY,
    process_name TEXT NOT NULL,
    cmdline TEXT,
    start_time DATETIME,
    runtime_sec INTEGER NOT NULL DEFAULT 0,
    cpu_percent REAL NOT NULL DEFAULT 0,
    mem_rss_bytes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'RUNNING',
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Cron jobs
CREATE TABLE IF NOT EXISTS cron_jobs (
    job_id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule TEXT NOT NULL,
    command TEXT NOT NULL,
    user TEXT,
    source TEXT NOT NULL DEFAULT 'system',
    next_run DATETIME,
    last_checked DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Xferlog entries
CREATE TABLE IF NOT EXISTS xferlog_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_time DATETIME NOT NULL,
    remote_host TEXT NOT NULL,
    username TEXT NOT NULL,
    filename TEXT NOT NULL,
    bytes INTEGER NOT NULL,
    transfer_time_sec INTEGER NOT NULL DEFAULT 0,
    direction TEXT NOT NULL,
    status TEXT NOT NULL
);
```

### Migration Runner

```go
func runMigrations(db *sql.DB) error {
    // Get current version
    var version int
    db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&version)

    // Apply pending migrations
    migrations := []string{
        schema001,
        schema002,
        // ...
    }

    for i := version; i < len(migrations); i++ {
        if _, err := db.Exec(migrations[i]); err != nil {
            return fmt.Errorf("migration %d failed: %w", i+1, err)
        }
        db.Exec("UPDATE meta SET value = ? WHERE key = 'schema_version'", i+1)
    }
    return nil
}
```

<!-- MANUAL: -->
