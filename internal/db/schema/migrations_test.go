package schema

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRunMigrations_FreshDB_CreatesAllTables(t *testing.T) {
	// Setup: Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Execute: Run migrations
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify: Check meta table exists and has schema_version
	var schemaVersion string
	err = db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&schemaVersion)
	if err != nil {
		t.Errorf("Failed to query schema_version: %v", err)
	}
	if schemaVersion != "1" {
		t.Errorf("Expected schema_version = '1', got '%s'", schemaVersion)
	}

	// Verify: Check filesystem_usage table exists and has correct columns
	rows, err := db.Query("SELECT mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at FROM filesystem_usage LIMIT 0")
	if err != nil {
		t.Errorf("filesystem_usage table missing or invalid: %v", err)
	}
	rows.Close()

	// Verify: Check path_stats table exists and has correct columns
	rows, err = db.Query("SELECT path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at FROM path_stats LIMIT 0")
	if err != nil {
		t.Errorf("path_stats table missing or invalid: %v", err)
	}
	rows.Close()
}

func TestRunMigrations_AlreadyMigrated_SkipsCompleted(t *testing.T) {
	// Setup: Create database and run migrations once
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// First migration
	if err := RunMigrations(db); err != nil {
		t.Fatalf("First RunMigrations failed: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent) VALUES ('/test', 1000, 500, 500, 50.0)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Execute: Run migrations again
	if err := RunMigrations(db); err != nil {
		t.Fatalf("Second RunMigrations failed: %v", err)
	}

	// Verify: Test data still exists (migrations didn't recreate tables)
	var mountPoint string
	err = db.QueryRow("SELECT mount_point FROM filesystem_usage WHERE mount_point = '/test'").Scan(&mountPoint)
	if err != nil {
		t.Errorf("Test data lost after second migration: %v", err)
	}
	if mountPoint != "/test" {
		t.Errorf("Expected mount_point = '/test', got '%s'", mountPoint)
	}

	// Verify: Schema version is still 1
	var schemaVersion string
	err = db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&schemaVersion)
	if err != nil {
		t.Errorf("Failed to query schema_version: %v", err)
	}
	if schemaVersion != "1" {
		t.Errorf("Expected schema_version = '1', got '%s'", schemaVersion)
	}
}

func TestRunMigrations_InvalidDB_ReturnsError(t *testing.T) {
	// Setup: Create database and close it
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	db.Close()

	// Remove database file to force error
	os.Remove(dbPath)

	// Execute: Run migrations on closed database
	err = RunMigrations(db)

	// Verify: Should return error
	if err == nil {
		t.Error("Expected error when running migrations on closed database, got nil")
	}
}
