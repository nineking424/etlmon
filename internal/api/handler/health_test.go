package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Returns200(t *testing.T) {
	handler := NewHealthHandler("test-node")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", response["status"])
	}
}

func TestHealthHandler_IncludesNodeName(t *testing.T) {
	nodeName := "my-etlmon-node"
	handler := NewHealthHandler(nodeName)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["node_name"] != nodeName {
		t.Errorf("expected node_name '%s', got '%v'", nodeName, response["node_name"])
	}
}

func TestHealthHandler_IncludesUptime(t *testing.T) {
	handler := NewHealthHandler("test-node")

	// Wait a bit to ensure uptime is > 0
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	uptimeSeconds, ok := response["uptime_seconds"].(float64)
	if !ok {
		t.Fatalf("expected uptime_seconds to be a number, got %T", response["uptime_seconds"])
	}

	if uptimeSeconds <= 0 {
		t.Errorf("expected uptime > 0, got %v", uptimeSeconds)
	}
}
