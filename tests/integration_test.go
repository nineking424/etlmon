//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/api"
	"github.com/etlmon/etlmon/internal/db"
	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/ui/client"
)

// testEnv holds the test environment
type testEnv struct {
	db     *db.DB
	repo   *repository.Repository
	server *api.Server
	client *client.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Create temp database
	dbPath := t.TempDir() + "/test.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	repo := repository.NewRepository(database.GetDB())

	ctx, cancel := context.WithCancel(context.Background())

	// Start server on random port (port 0 makes the OS assign a free port)
	server := api.NewServer("127.0.0.1:0", repo, "test-node", "")

	// Start server in background
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// Get actual port from server
	addr := server.Addr()

	httpClient := client.NewClient("http://" + addr)

	return &testEnv{
		db:     database,
		repo:   repo,
		server: server,
		client: httpClient,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (e *testEnv) teardown(t *testing.T) {
	t.Helper()
	e.cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.server.Shutdown(shutdownCtx); err != nil {
		t.Logf("shutdown error: %v", err)
	}
	e.db.Close()
}

func TestIntegration_NodeStartup(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown(t)

	// Verify server is responding
	resp, err := http.Get("http://" + env.server.Addr() + "/api/v1/health")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	t.Logf("Server started successfully at %s", env.server.Addr())
}

func TestIntegration_FSEndpoint(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown(t)

	usage, err := env.client.GetFilesystemUsage(env.ctx)
	if err != nil {
		t.Fatalf("failed to get filesystem usage: %v", err)
	}

	// Empty is OK - just verify the endpoint works
	t.Logf("got %d filesystem entries", len(usage))
}

func TestIntegration_PathsEndpoint(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown(t)

	paths, err := env.client.GetPathStats(env.ctx)
	if err != nil {
		t.Fatalf("failed to get path stats: %v", err)
	}

	// Empty is OK - just verify the endpoint works
	t.Logf("got %d path entries", len(paths))
}

func TestIntegration_TriggerScan(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown(t)

	// Trigger scan with nil paths (scan all configured paths)
	// In integration tests, scanner is not configured, so we expect an error
	err := env.client.TriggerScan(env.ctx, nil)
	if err != nil {
		// Expected - scanner not configured in test environment
		apiErr, ok := err.(*client.APIError)
		if !ok {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", apiErr.StatusCode)
		}
		t.Log("Scan endpoint correctly returns 501 when scanner not configured")
	} else {
		t.Log("Scan triggered successfully")
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown(t)

	resp, err := http.Get("http://" + env.server.Addr() + "/api/v1/health")
	if err != nil {
		t.Fatalf("failed to get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var healthResp struct {
		NodeName string  `json:"node_name"`
		Status   string  `json:"status"`
		Uptime   float64 `json:"uptime_seconds"`
	}

	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if healthResp.NodeName != "test-node" {
		t.Errorf("expected node name 'test-node', got '%s'", healthResp.NodeName)
	}

	if healthResp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", healthResp.Status)
	}

	if healthResp.Uptime < 0 {
		t.Errorf("expected positive uptime, got %f", healthResp.Uptime)
	}

	t.Logf("Health check passed: node=%s, status=%s, uptime=%.2fs",
		healthResp.NodeName, healthResp.Status, healthResp.Uptime)
}
