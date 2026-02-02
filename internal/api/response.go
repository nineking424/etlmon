package api

import (
	"encoding/json"
	"net/http"

	"github.com/etlmon/etlmon/pkg/models"
)

// writeJSON writes a JSON response with the given status code and data
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, we can't send an error response since headers are already sent
		// Log this error in production
		return
	}
}

// writeError writes an error response in JSON format
func writeError(w http.ResponseWriter, status int, err error) {
	errResp := models.ErrorResponse{
		Error: err.Error(),
	}

	writeJSON(w, status, errResp)
}
