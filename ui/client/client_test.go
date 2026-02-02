package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Get_ParsesResponse(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"data": map[string]string{
				"message": "success",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	var result map[string]string
	err := client.get(context.Background(), "/test", &result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "success", result["message"])
}

func TestClient_Get_HandlesAPIError(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not found",
			"code":  "NOT_FOUND",
		})
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	var result map[string]string
	err := client.get(context.Background(), "/missing", &result)

	// Assert
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "error should be of type *APIError")
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "not found")
}

func TestClient_Get_HandlesTimeout(t *testing.T) {
	// Setup test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with short timeout
	client := NewClient(server.URL)
	client.SetTimeout(50 * time.Millisecond)

	var result map[string]string
	err := client.get(context.Background(), "/slow", &result)

	// Assert - timeout errors contain "deadline exceeded" or "Timeout exceeded"
	require.Error(t, err)
	errMsg := err.Error()
	hasTimeoutError := strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "Timeout exceeded") ||
		strings.Contains(errMsg, "timeout")
	assert.True(t, hasTimeoutError, "expected timeout error, got: %s", errMsg)
}

func TestClient_Post_SendsBody(t *testing.T) {
	// Setup test server
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read body
		json.NewDecoder(r.Body).Decode(&receivedBody)

		// Send response
		response := map[string]interface{}{
			"data": map[string]string{
				"status": "created",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test
	client := NewClient(server.URL)
	body := map[string]string{"name": "test"}
	var result map[string]string
	err := client.post(context.Background(), "/test", body, &result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "created", result["status"])
	assert.Equal(t, "test", receivedBody["name"])
}
