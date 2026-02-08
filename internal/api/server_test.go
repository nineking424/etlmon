package api

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/etlmon/etlmon/internal/db/repository"
)

func setupServerTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
		CREATE TABLE filesystem_usage (
			mount_point TEXT PRIMARY KEY,
			total_bytes INTEGER NOT NULL,
			used_bytes INTEGER NOT NULL,
			avail_bytes INTEGER NOT NULL,
			used_percent REAL NOT NULL,
			collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE path_stats (
			path TEXT PRIMARY KEY,
			file_count INTEGER NOT NULL DEFAULT 0,
			dir_count INTEGER NOT NULL DEFAULT 0,
			scan_duration_ms INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'OK',
			error_message TEXT,
			collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE process_stats (
			pid INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			user TEXT NOT NULL,
			cpu_percent REAL NOT NULL DEFAULT 0,
			mem_rss INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'unknown',
			elapsed TEXT NOT NULL DEFAULT '',
			collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE log_lines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			log_name TEXT NOT NULL,
			log_path TEXT NOT NULL,
			line TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_log_lines_name ON log_lines(log_name, id DESC);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestServer_Start_ListensOnConfiguredAddress(t *testing.T) {
	db := setupServerTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	server := NewServer("127.0.0.1:0", repo, "test-node", "") // Port 0 = random available port

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Check that server is running by making a request
	addr := server.Addr()
	resp, err := http.Get("http://" + addr + "/api/v1/health")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}

	// Check that Start() returned (or error)
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("server did not shutdown within timeout")
	}
}

func TestServer_Shutdown_GracefullyStops(t *testing.T) {
	db := setupServerTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	server := NewServer("127.0.0.1:0", repo, "test-node", "")

	go server.Start()
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestServer_Routes_FSEndpoint(t *testing.T) {
	db := setupServerTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	server := NewServer("127.0.0.1:0", repo, "test-node", "")

	handler := server.setupRoutes()

	// Verify that handler is not nil
	if handler == nil {
		t.Fatal("setupRoutes returned nil handler")
	}
}

func TestServer_Routes_PathsEndpoint(t *testing.T) {
	db := setupServerTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	server := NewServer("127.0.0.1:0", repo, "test-node", "")

	handler := server.setupRoutes()

	if handler == nil {
		t.Fatal("setupRoutes returned nil handler")
	}
}

func TestServer_Routes_HealthEndpoint(t *testing.T) {
	db := setupServerTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)
	server := NewServer("127.0.0.1:0", repo, "test-node", "")

	handler := server.setupRoutes()

	if handler == nil {
		t.Fatal("setupRoutes returned nil handler")
	}
}
