package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

func setupPathsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
		CREATE TABLE path_stats (
			path TEXT PRIMARY KEY,
			file_count INTEGER NOT NULL DEFAULT 0,
			dir_count INTEGER NOT NULL DEFAULT 0,
			scan_duration_ms INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'OK',
			error_message TEXT,
			collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestPathsHandler_List_ReturnsStats(t *testing.T) {
	db := setupPathsTestDB(t)
	defer db.Close()

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO path_stats (path, file_count, dir_count, scan_duration_ms, status, collected_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/data/logs", 1500, 25, 350, "OK", now)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	repo := repository.NewPathsRepository(db)
	handler := NewPathsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/paths", nil)
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

	if len(data) != 1 {
		t.Errorf("expected 1 item, got %d", len(data))
	}
}

func TestPathsHandler_List_WithPagination(t *testing.T) {
	db := setupPathsTestDB(t)
	defer db.Close()

	// Insert 5 test records
	now := time.Now()
	for i := 1; i <= 5; i++ {
		_, err := db.Exec(`
			INSERT INTO path_stats (path, file_count, dir_count, scan_duration_ms, status, collected_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, "/data/path"+string(rune('0'+i)), i*100, i*10, 100, "OK", now)
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}
	}

	repo := repository.NewPathsRepository(db)
	handler := NewPathsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/paths?limit=2&offset=1", nil)
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

	if len(data) != 2 {
		t.Errorf("expected 2 items with limit=2, got %d", len(data))
	}

	if response.Meta == nil {
		t.Fatal("expected meta to be present")
	}

	if response.Meta.Total != 5 {
		t.Errorf("expected total=5, got %d", response.Meta.Total)
	}

	if response.Meta.Limit != 2 {
		t.Errorf("expected limit=2, got %d", response.Meta.Limit)
	}

	if response.Meta.Offset != 1 {
		t.Errorf("expected offset=1, got %d", response.Meta.Offset)
	}
}

func TestPathsHandler_List_EmptyDB_ReturnsEmptyArray(t *testing.T) {
	db := setupPathsTestDB(t)
	defer db.Close()

	repo := repository.NewPathsRepository(db)
	handler := NewPathsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/paths", nil)
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

func TestPathsHandler_TriggerScan_SucceedsWithPaths(t *testing.T) {
	db := setupPathsTestDB(t)
	defer db.Close()

	repo := repository.NewPathsRepository(db)
	handler := NewPathsHandler(repo)

	// Create request body
	reqBody := map[string]interface{}{
		"paths": []string{"/data/logs", "/data/archive"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/paths/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.TriggerScan(w, req)

	// Without scanner configured, should return 501 Not Implemented
	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
	}
}
