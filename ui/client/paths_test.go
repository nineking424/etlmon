package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetPathStats_ReturnsStats(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/paths", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		stats := []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      1500,
				DirCount:       25,
				ScanDurationMs: 250,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
			{
				Path:           "/data/archive",
				FileCount:      5000,
				DirCount:       100,
				ScanDurationMs: 850,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
		}

		response := map[string]interface{}{
			"data": stats,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	stats, err := client.GetPathStats(context.Background())

	// Assert
	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, "/data/logs", stats[0].Path)
	assert.Equal(t, int64(1500), stats[0].FileCount)
	assert.Equal(t, "OK", stats[0].Status)
	assert.Equal(t, "/data/archive", stats[1].Path)
	assert.Equal(t, int64(5000), stats[1].FileCount)
}

func TestClient_TriggerScan_Succeeds(t *testing.T) {
	// Setup test server
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/paths/scan", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read body
		json.NewDecoder(r.Body).Decode(&receivedBody)

		response := map[string]interface{}{
			"data": map[string]string{
				"status": "triggered",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	paths := []string{"/data/logs", "/data/archive"}
	err := client.TriggerScan(context.Background(), paths)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, receivedBody["paths"])
	pathsReceived := receivedBody["paths"].([]interface{})
	assert.Len(t, pathsReceived, 2)
	assert.Equal(t, "/data/logs", pathsReceived[0])
	assert.Equal(t, "/data/archive", pathsReceived[1])
}
