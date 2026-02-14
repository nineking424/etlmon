package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/db/repository"
	"github.com/etlmon/etlmon/pkg/models"
)

// LogHandler handles log entry API requests
type LogHandler struct {
	repo       *repository.LogRepository
	configPath string
}

// NewLogHandler creates a new log handler
func NewLogHandler(repo *repository.LogRepository, configPath string) *LogHandler {
	return &LogHandler{repo: repo, configPath: configPath}
}

// List handles GET /api/v1/logs
func (h *LogHandler) List(w http.ResponseWriter, r *http.Request) {
	// Optional query params: name (filter by log name), limit
	logName := r.URL.Query().Get("name")
	limitStr := r.URL.Query().Get("limit")

	limit := 200
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var entries []models.LogEntry
	var err error

	if logName != "" {
		results, e := h.repo.GetLogEntries(r.Context(), logName, limit)
		if e != nil {
			writeError(w, http.StatusInternalServerError, e)
			return
		}
		for _, item := range results {
			entries = append(entries, *item)
		}
	} else {
		entries, err = h.repo.ListAll()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	if entries == nil {
		entries = []models.LogEntry{}
	}

	resp := models.Response{Data: entries}
	writeJSON(w, http.StatusOK, resp)
}

// ListFiles handles GET /api/v1/logs/files
func (h *LogHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadNodeConfig(h.configPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	files := make([]models.LogFileInfo, 0, len(cfg.Logs))
	for _, l := range cfg.Logs {
		info := models.LogFileInfo{
			Name:     l.Name,
			Path:     l.Path,
			MaxLines: l.MaxLines,
		}
		if stat, err := os.Stat(l.Path); err == nil {
			info.Size = stat.Size()
			info.ModTime = stat.ModTime()
		}
		files = append(files, info)
	}

	resp := models.Response{Data: files}
	writeJSON(w, http.StatusOK, resp)
}
