package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewDB_CreatesDatabase(t *testing.T) {
	// Setup: Create temp database path
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Execute: Create new DB
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Verify: Database file exists and can be queried
	var result int
	err = db.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected result = 1, got %d", result)
	}

	// Verify: Migrations have run (meta table exists)
	var schemaVersion string
	err = db.db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&schemaVersion)
	if err != nil {
		t.Errorf("Migrations did not run (meta table missing): %v", err)
	}
}

func TestNewDB_EnablesWALMode(t *testing.T) {
	// Setup: Create temp database path
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Execute: Create new DB with WAL mode
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Verify: WAL mode is enabled
	var journalMode string
	err = db.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Errorf("Failed to check journal_mode: %v", err)
	}
	if !strings.EqualFold(journalMode, "wal") {
		t.Errorf("Expected journal_mode = 'wal', got '%s'", journalMode)
	}

	// Verify: Synchronous mode is NORMAL for performance
	var synchronous string
	err = db.db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	if err != nil {
		t.Errorf("Failed to check synchronous: %v", err)
	}
	// NORMAL = 1
	if synchronous != "1" {
		t.Errorf("Expected synchronous = '1' (NORMAL), got '%s'", synchronous)
	}
}

func TestDB_WithTx_CommitsOnSuccess(t *testing.T) {
	// Setup: Create database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Execute: Run transaction that succeeds
	ctx := context.Background()
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent) VALUES ('/test', 1000, 500, 500, 50.0)`)
		return err
	})

	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	// Verify: Data was committed
	var mountPoint string
	err = db.db.QueryRow("SELECT mount_point FROM filesystem_usage WHERE mount_point = '/test'").Scan(&mountPoint)
	if err != nil {
		t.Errorf("Transaction was not committed: %v", err)
	}
	if mountPoint != "/test" {
		t.Errorf("Expected mount_point = '/test', got '%s'", mountPoint)
	}
}

func TestDB_WithTx_RollbacksOnError(t *testing.T) {
	// Setup: Create database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Execute: Run transaction that fails
	ctx := context.Background()
	expectedErr := errors.New("forced error")
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent) VALUES ('/test', 1000, 500, 500, 50.0)`)
		if err != nil {
			return err
		}
		return expectedErr // Force error to trigger rollback
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify: Data was NOT committed (rolled back)
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM filesystem_usage WHERE mount_point = '/test'").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if count != 0 {
		t.Errorf("Transaction was not rolled back, found %d rows", count)
	}
}

func TestDB_WithTx_RespectsContext(t *testing.T) {
	// Setup: Create database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Execute: Run transaction with cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond) // Ensure context is cancelled

	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		time.Sleep(10 * time.Millisecond)
		_, err := tx.Exec(`INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent) VALUES ('/test', 1000, 500, 500, 50.0)`)
		return err
	})

	// Verify: Should get context error
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestDB_Compact_Succeeds(t *testing.T) {
	// Setup: Create database with some data
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Insert and delete data to create fragmentation
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		mountPoint := filepath.Join("/test", fmt.Sprintf("%d", i))
		_, insertErr := db.db.Exec(`INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent) VALUES (?, 1000, 500, 500, 50.0)`, mountPoint)
		if insertErr != nil {
			t.Fatalf("Failed to insert test data: %v", insertErr)
		}
	}

	_, err = db.db.Exec(`DELETE FROM filesystem_usage`)
	if err != nil {
		t.Fatalf("Failed to delete test data: %v", err)
	}

	// Execute: Compact database
	err = db.Compact(ctx)

	// Verify: Should succeed without error
	if err != nil {
		t.Errorf("Compact failed: %v", err)
	}

	// Verify: Database is still functional
	var result int
	err = db.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Database not functional after compact: %v", err)
	}
}

func TestDB_Close_Succeeds(t *testing.T) {
	// Setup: Create database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}

	// Execute: Close database
	err = db.Close()

	// Verify: Should succeed
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify: Database is no longer usable
	err = db.db.Ping()
	if err == nil {
		t.Error("Expected error after Close, database still usable")
	}
}

func TestDB_GetDB_ReturnsUnderlyingDB(t *testing.T) {
	// Setup: Create database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Execute: Get underlying *sql.DB
	sqlDB := db.GetDB()

	// Verify: Should return non-nil database
	if sqlDB == nil {
		t.Error("GetDB returned nil")
	}

	// Verify: Can use returned database
	var result int
	err = sqlDB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Returned database is not functional: %v", err)
	}
}
