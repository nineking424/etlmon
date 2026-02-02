package handler

import (
	"encoding/json"
	"net/http"

	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

// FSHandler handles filesystem usage API requests
type FSHandler struct {
	repo *repository.FSRepository
}

// NewFSHandler creates a new filesystem handler
func NewFSHandler(repo *repository.FSRepository) *FSHandler {
	return &FSHandler{repo: repo}
}

// List handles GET /api/v1/fs
func (h *FSHandler) List(w http.ResponseWriter, r *http.Request) {
	usages, err := h.repo.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Ensure we always return an array, even if empty
	if usages == nil {
		usages = []models.FilesystemUsage{}
	}

	resp := models.Response{
		Data: usages,
	}

	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, err error) {
	errResp := models.ErrorResponse{
		Error: err.Error(),
	}
	writeJSON(w, status, errResp)
}
