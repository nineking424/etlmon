package api

import (
	"net/http"

	"github.com/etlmon/etlmon/internal/api/handler"
)

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	fsHandler := handler.NewFSHandler(s.repo.FS)
	pathsHandler := handler.NewPathsHandler(s.repo.Paths)
	healthHandler := handler.NewHealthHandler(s.nodeName)
	processHandler := handler.NewProcessHandler(s.repo.Process)
	logHandler := handler.NewLogHandler(s.repo.Log)

	// Set scanner proxy (supports hot-swap on config reload)
	pathsHandler.SetScanner(s.scannerProxy)

	// Config handler with reload callback
	configHandler := handler.NewConfigHandler(s.configPath, s.onConfigReload)
	mux.HandleFunc("/api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configHandler.Get(w, r)
		case http.MethodPut:
			configHandler.Update(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Register routes
	mux.HandleFunc("/api/v1/fs", fsHandler.List)
	mux.HandleFunc("/api/v1/paths", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			pathsHandler.List(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/paths/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			pathsHandler.TriggerScan(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/health", healthHandler.Health)
	mux.HandleFunc("/api/v1/processes", processHandler.List)
	mux.HandleFunc("/api/v1/logs", logHandler.List)

	return mux
}
