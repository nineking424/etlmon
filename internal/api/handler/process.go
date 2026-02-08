package handler

import (
	"net/http"

	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

// ProcessHandler handles process info API requests
type ProcessHandler struct {
	repo *repository.ProcessRepository
}

// NewProcessHandler creates a new process handler
func NewProcessHandler(repo *repository.ProcessRepository) *ProcessHandler {
	return &ProcessHandler{repo: repo}
}

// List handles GET /api/v1/processes
func (h *ProcessHandler) List(w http.ResponseWriter, r *http.Request) {
	procs, err := h.repo.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if procs == nil {
		procs = []models.ProcessInfo{}
	}
	resp := models.Response{Data: procs}
	writeJSON(w, http.StatusOK, resp)
}
