package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create schema
	schema := `
		CREATE TABLE filesystem_usage (
			mount_point TEXT PRIMARY KEY,
			total_bytes INTEGER NOT NULL,
			used_bytes INTEGER NOT NULL,
			avail_bytes INTEGER NOT NULL,
			used_percent REAL NOT NULL,
			collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestFSHandler_List_ReturnsUsage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO filesystem_usage (mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/data", 1000000, 400000, 600000, 40.0, now)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	repo := repository.NewFSRepository(db)
	handler := NewFSHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fs", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Response.Data should be a slice
	data, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", response.Data)
	}

	if len(data) != 1 {
		t.Errorf("expected 1 item, got %d", len(data))
	}
}

func TestFSHandler_List_EmptyDB_ReturnsEmptyArray(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewFSRepository(db)
	handler := NewFSHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fs", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", response.Data)
	}

	if len(data) != 0 {
		t.Errorf("expected empty array, got %d items", len(data))
	}
}

func TestFSHandler_List_RepoError_Returns500(t *testing.T) {
	db := setupTestDB(t)

	// Create repository first (before closing DB)
	repo := repository.NewFSRepository(db)
	handler := NewFSHandler(repo)

	// Close DB after repository is created to cause error on query
	db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fs", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error == "" {
		t.Error("expected error message, got empty string")
	}
}
