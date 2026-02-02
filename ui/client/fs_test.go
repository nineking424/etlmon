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

func TestClient_GetFilesystemUsage_ReturnsUsage(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/fs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		usage := []*models.FilesystemUsage{
			{
				MountPoint:  "/data",
				TotalBytes:  1000000000,
				UsedBytes:   600000000,
				AvailBytes:  400000000,
				UsedPercent: 60.0,
				CollectedAt: time.Now(),
			},
			{
				MountPoint:  "/logs",
				TotalBytes:  500000000,
				UsedBytes:   450000000,
				AvailBytes:  50000000,
				UsedPercent: 90.0,
				CollectedAt: time.Now(),
			},
		}

		response := map[string]interface{}{
			"data": usage,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	usage, err := client.GetFilesystemUsage(context.Background())

	// Assert
	require.NoError(t, err)
	require.Len(t, usage, 2)
	assert.Equal(t, "/data", usage[0].MountPoint)
	assert.Equal(t, uint64(1000000000), usage[0].TotalBytes)
	assert.Equal(t, 60.0, usage[0].UsedPercent)
	assert.Equal(t, "/logs", usage[1].MountPoint)
	assert.Equal(t, 90.0, usage[1].UsedPercent)
}

func TestClient_GetFilesystemUsage_HandlesEmpty(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	usage, err := client.GetFilesystemUsage(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Empty(t, usage)
}
