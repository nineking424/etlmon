package handler

import (
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	nodeName  string
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(nodeName string) *HealthHandler {
	return &HealthHandler{
		nodeName:  nodeName,
		startTime: time.Now(),
	}
}

// Health handles GET /api/v1/health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime).Seconds()

	response := map[string]interface{}{
		"status":         "ok",
		"node_name":      h.nodeName,
		"uptime_seconds": uptime,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, response)
}
