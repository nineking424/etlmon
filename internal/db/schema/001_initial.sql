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
