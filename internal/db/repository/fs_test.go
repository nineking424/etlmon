package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/db"
	"github.com/etlmon/etlmon/pkg/models"
)

func setupTestDB(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return database
}

func TestFSRepository_Save_InsertsNewRecord(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewFSRepository(database.GetDB())
	defer repo.Close()

	// Execute
	ctx := context.Background()
	usage := &models.FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000,
		UsedBytes:   600000,
		AvailBytes:  400000,
		UsedPercent: 60.0,
		CollectedAt: time.Now(),
	}

	err := repo.Save(ctx, usage)

	// Verify
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Query directly to verify
	var mountPoint string
	var usedPercent float64
	err = database.GetDB().QueryRow("SELECT mount_point, used_percent FROM filesystem_usage WHERE mount_point = '/data'").Scan(&mountPoint, &usedPercent)
	if err != nil {
		t.Errorf("Failed to query saved record: %v", err)
	}
	if mountPoint != "/data" {
		t.Errorf("Expected mount_point = '/data', got '%s'", mountPoint)
	}
	if usedPercent != 60.0 {
		t.Errorf("Expected used_percent = 60.0, got %f", usedPercent)
	}
}

func TestFSRepository_Save_UpdatesExistingRecord(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewFSRepository(database.GetDB())
	defer repo.Close()

	ctx := context.Background()

	// Insert initial record
	usage1 := &models.FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000,
		UsedBytes:   600000,
		AvailBytes:  400000,
		UsedPercent: 60.0,
		CollectedAt: time.Now(),
	}
	repo.Save(ctx, usage1)

	// Execute: Update with new values
	usage2 := &models.FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000,
		UsedBytes:   700000,
		AvailBytes:  300000,
		UsedPercent: 70.0,
		CollectedAt: time.Now(),
	}
	err := repo.Save(ctx, usage2)

	// Verify
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Verify only one record exists with updated value
	var count int
	var usedPercent float64
	database.GetDB().QueryRow("SELECT COUNT(*), MAX(used_percent) FROM filesystem_usage WHERE mount_point = '/data'").Scan(&count, &usedPercent)

	if count != 1 {
		t.Errorf("Expected 1 record, got %d (should replace, not insert new)", count)
	}
	if usedPercent != 70.0 {
		t.Errorf("Expected used_percent = 70.0, got %f", usedPercent)
	}
}

func TestFSRepository_GetLatest_ReturnsAllRecords(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewFSRepository(database.GetDB())
	defer repo.Close()

	ctx := context.Background()

	// Insert multiple records
	repo.Save(ctx, &models.FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000,
		UsedBytes:   600000,
		AvailBytes:  400000,
		UsedPercent: 60.0,
		CollectedAt: time.Now(),
	})
	repo.Save(ctx, &models.FilesystemUsage{
		MountPoint:  "/home",
		TotalBytes:  2000000,
		UsedBytes:   1200000,
		AvailBytes:  800000,
		UsedPercent: 60.0,
		CollectedAt: time.Now(),
	})

	// Execute
	results, err := repo.GetLatest(ctx)

	// Verify
	if err != nil {
		t.Errorf("GetLatest failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify sorting by mount_point
	if len(results) == 2 {
		if results[0].MountPoint != "/data" {
			t.Errorf("Expected first result mount_point = '/data', got '%s'", results[0].MountPoint)
		}
		if results[1].MountPoint != "/home" {
			t.Errorf("Expected second result mount_point = '/home', got '%s'", results[1].MountPoint)
		}
	}
}

func TestFSRepository_GetLatest_EmptyDatabase_ReturnsEmptySlice(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewFSRepository(database.GetDB())
	defer repo.Close()

	// Execute
	ctx := context.Background()
	results, err := repo.GetLatest(ctx)

	// Verify
	if err != nil {
		t.Errorf("GetLatest failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestFSRepository_Close_ClosesStatements(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewFSRepository(database.GetDB())

	// Execute
	err := repo.Close()

	// Verify
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify statements are closed (they should return error)
	ctx := context.Background()
	usage := &models.FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000,
		UsedBytes:   600000,
		AvailBytes:  400000,
		UsedPercent: 60.0,
		CollectedAt: time.Now(),
	}

	err = repo.Save(ctx, usage)
	if err == nil {
		t.Error("Expected error when using closed statement, got nil")
	}
}
