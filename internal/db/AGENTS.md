<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# db

## Purpose

SQLite database layer. Handles schema migrations, provides repository interfaces for data access, manages connection and transactions.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `schema/` | SQL migrations (see `schema/AGENTS.md`) |
| `repository/` | Data access interfaces (see `repository/AGENTS.md`) |

## Key Files

| File | Description |
|------|-------------|
| `db.go` | Database connection and initialization |
| `db_test.go` | Database tests |

## For AI Agents

### Working In This Directory

- Use `github.com/mattn/go-sqlite3` driver
- Enable WAL mode for concurrent reads
- Single writer pattern: collectors write, API reads
- Use prepared statements for repeated queries
- Keep transactions short

### Database Configuration

```go
// Connection string with WAL mode
dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000", dbPath)

db, err := sql.Open("sqlite3", dsn)
if err != nil {
    return nil, err
}

// Set connection pool (single writer)
db.SetMaxOpenConns(1)  // For writes
db.SetMaxIdleConns(1)
```

### Initialization

```go
func NewDB(dbPath string) (*DB, error) {
    db, _ := sql.Open("sqlite3", dsn)

    // Run migrations
    if err := runMigrations(db); err != nil {
        return nil, fmt.Errorf("migrations failed: %w", err)
    }

    return &DB{db: db}, nil
}
```

### Transaction Pattern

```go
// For writes (collectors)
func (d *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
    tx, err := d.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit()
}

// For reads (API)
func (d *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return d.db.QueryContext(ctx, query, args...)
}
```

### Maintenance

```go
// VACUUM and checkpoint (call periodically or on admin endpoint)
func (d *DB) Compact(ctx context.Context) error {
    _, err := d.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
    if err != nil {
        return err
    }
    _, err = d.db.ExecContext(ctx, "VACUUM")
    return err
}
```

<!-- MANUAL: -->
