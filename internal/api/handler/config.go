package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/pkg/models"
)

// ConfigHandler handles configuration API requests
type ConfigHandler struct {
	configPath string
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configPath string) *ConfigHandler {
	return &ConfigHandler{configPath: configPath}
}

// Get handles GET /api/v1/config
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadNodeConfig(h.configPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	resp := models.Response{Data: cfg}
	writeJSON(w, http.StatusOK, resp)
}

// Update handles PUT /api/v1/config
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var cfg config.NodeConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := config.SaveNodeConfig(h.configPath, &cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	resp := models.Response{Data: map[string]string{"status": "saved"}}
	writeJSON(w, http.StatusOK, resp)
}
