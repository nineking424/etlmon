package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

// PathScanner interface for triggering path scans
type PathScanner interface {
	ScanPaths(paths []string) error
}

// PathsHandler handles path statistics API requests
type PathsHandler struct {
	repo    *repository.PathsRepository
	scanner PathScanner // Optional scanner for triggering scans
}

// NewPathsHandler creates a new paths handler
func NewPathsHandler(repo *repository.PathsRepository) *PathsHandler {
	return &PathsHandler{repo: repo}
}

// SetScanner sets the path scanner (optional)
func (h *PathsHandler) SetScanner(scanner PathScanner) {
	h.scanner = scanner
}

// List handles GET /api/v1/paths
func (h *PathsHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 0
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeError(w, http.StatusBadRequest, errors.New("invalid limit parameter"))
			return
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeError(w, http.StatusBadRequest, errors.New("invalid offset parameter"))
			return
		}
	}

	var stats []models.PathStats
	var err error
	var total int

	// If pagination is requested
	if limit > 0 {
		stats, err = h.repo.ListWithPagination(limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		total, err = h.repo.Count()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		stats, err = h.repo.ListAll()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		total = len(stats)
	}

	// Ensure we always return an array
	if stats == nil {
		stats = []models.PathStats{}
	}

	resp := models.Response{
		Data: stats,
	}

	// Add metadata if pagination is used
	if limit > 0 {
		resp.Meta = &models.Meta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// TriggerScan handles POST /api/v1/paths/scan
func (h *PathsHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	if h.scanner == nil {
		writeError(w, http.StatusNotImplemented, errors.New("path scanner not configured"))
		return
	}

	var req struct {
		Paths []string `json:"paths"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if len(req.Paths) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("no paths provided"))
		return
	}

	if err := h.scanner.ScanPaths(req.Paths); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	resp := models.Response{
		Data: map[string]interface{}{
			"status":  "scan triggered",
			"paths":   req.Paths,
			"message": "scan initiated successfully",
		},
	}

	writeJSON(w, http.StatusAccepted, resp)
}
