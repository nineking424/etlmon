-- Schema version tracking
CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT
);

INSERT OR IGNORE INTO meta (key, value) VALUES ('schema_version', '1');

-- Filesystem usage
CREATE TABLE IF NOT EXISTS filesystem_usage (
    mount_point TEXT PRIMARY KEY,
    total_bytes INTEGER NOT NULL,
    used_bytes INTEGER NOT NULL,
    avail_bytes INTEGER NOT NULL,
    used_percent REAL NOT NULL,
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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

-- Process statistics
CREATE TABLE IF NOT EXISTS process_stats (
    pid INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    user TEXT NOT NULL,
    cpu_percent REAL NOT NULL DEFAULT 0,
    mem_rss INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'unknown',
    elapsed TEXT NOT NULL DEFAULT '',
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Log lines
CREATE TABLE IF NOT EXISTS log_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_name TEXT NOT NULL,
    log_path TEXT NOT NULL,
    line TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_log_lines_name ON log_lines(log_name, id DESC);
