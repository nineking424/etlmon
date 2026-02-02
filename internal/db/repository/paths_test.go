package repository

import (
	"context"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

func TestPathsRepository_Save_InsertsNewRecord(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())
	defer repo.Close()

	// Execute
	ctx := context.Background()
	stats := &models.PathStats{
		Path:           "/data/logs",
		FileCount:      1500,
		DirCount:       25,
		ScanDurationMs: 340,
		Status:         "OK",
		CollectedAt:    time.Now(),
	}

	err := repo.Save(ctx, stats)

	// Verify
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Query directly to verify
	var path string
	var fileCount int64
	var status string
	err = database.GetDB().QueryRow("SELECT path, file_count, status FROM path_stats WHERE path = '/data/logs'").Scan(&path, &fileCount, &status)
	if err != nil {
		t.Errorf("Failed to query saved record: %v", err)
	}
	if path != "/data/logs" {
		t.Errorf("Expected path = '/data/logs', got '%s'", path)
	}
	if fileCount != 1500 {
		t.Errorf("Expected file_count = 1500, got %d", fileCount)
	}
	if status != "OK" {
		t.Errorf("Expected status = 'OK', got '%s'", status)
	}
}

func TestPathsRepository_Save_UpdatesExistingRecord(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())
	defer repo.Close()

	ctx := context.Background()

	// Insert initial record
	stats1 := &models.PathStats{
		Path:           "/data/logs",
		FileCount:      1500,
		DirCount:       25,
		ScanDurationMs: 340,
		Status:         "OK",
		CollectedAt:    time.Now(),
	}
	repo.Save(ctx, stats1)

	// Execute: Update with new values
	stats2 := &models.PathStats{
		Path:           "/data/logs",
		FileCount:      1600,
		DirCount:       26,
		ScanDurationMs: 350,
		Status:         "OK",
		CollectedAt:    time.Now(),
	}
	err := repo.Save(ctx, stats2)

	// Verify
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Verify only one record exists with updated value
	var count int
	var fileCount int64
	database.GetDB().QueryRow("SELECT COUNT(*), MAX(file_count) FROM path_stats WHERE path = '/data/logs'").Scan(&count, &fileCount)

	if count != 1 {
		t.Errorf("Expected 1 record, got %d (should replace, not insert new)", count)
	}
	if fileCount != 1600 {
		t.Errorf("Expected file_count = 1600, got %d", fileCount)
	}
}

func TestPathsRepository_Save_WithError_StoresErrorMessage(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())
	defer repo.Close()

	// Execute
	ctx := context.Background()
	stats := &models.PathStats{
		Path:           "/data/logs",
		FileCount:      0,
		DirCount:       0,
		ScanDurationMs: 50,
		Status:         "ERROR",
		ErrorMessage:   "permission denied",
		CollectedAt:    time.Now(),
	}

	err := repo.Save(ctx, stats)

	// Verify
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Query directly to verify
	var status string
	var errorMessage string
	err = database.GetDB().QueryRow("SELECT status, error_message FROM path_stats WHERE path = '/data/logs'").Scan(&status, &errorMessage)
	if err != nil {
		t.Errorf("Failed to query saved record: %v", err)
	}
	if status != "ERROR" {
		t.Errorf("Expected status = 'ERROR', got '%s'", status)
	}
	if errorMessage != "permission denied" {
		t.Errorf("Expected error_message = 'permission denied', got '%s'", errorMessage)
	}
}

func TestPathsRepository_GetAll_ReturnsAllRecords(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())
	defer repo.Close()

	ctx := context.Background()

	// Insert multiple records
	repo.Save(ctx, &models.PathStats{
		Path:           "/data/logs",
		FileCount:      1500,
		DirCount:       25,
		ScanDurationMs: 340,
		Status:         "OK",
		CollectedAt:    time.Now(),
	})
	repo.Save(ctx, &models.PathStats{
		Path:           "/var/log",
		FileCount:      800,
		DirCount:       10,
		ScanDurationMs: 150,
		Status:         "OK",
		CollectedAt:    time.Now(),
	})

	// Execute
	results, err := repo.GetAll(ctx)

	// Verify
	if err != nil {
		t.Errorf("GetAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify sorting by path
	if len(results) == 2 {
		if results[0].Path != "/data/logs" {
			t.Errorf("Expected first result path = '/data/logs', got '%s'", results[0].Path)
		}
		if results[1].Path != "/var/log" {
			t.Errorf("Expected second result path = '/var/log', got '%s'", results[1].Path)
		}
	}
}

func TestPathsRepository_GetAll_EmptyDatabase_ReturnsEmptySlice(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())
	defer repo.Close()

	// Execute
	ctx := context.Background()
	results, err := repo.GetAll(ctx)

	// Verify
	if err != nil {
		t.Errorf("GetAll failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestPathsRepository_Close_ClosesStatements(t *testing.T) {
	// Setup
	database := setupTestDB(t)
	defer database.Close()

	repo := NewPathsRepository(database.GetDB())

	// Execute
	err := repo.Close()

	// Verify
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify statements are closed
	ctx := context.Background()
	stats := &models.PathStats{
		Path:           "/data/logs",
		FileCount:      1500,
		DirCount:       25,
		ScanDurationMs: 340,
		Status:         "OK",
		CollectedAt:    time.Now(),
	}

	err = repo.Save(ctx, stats)
	if err == nil {
		t.Error("Expected error when using closed statement, got nil")
	}
}
