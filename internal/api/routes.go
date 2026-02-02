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

	// Set scanner if available
	if s.scanner != nil {
		pathsHandler.SetScanner(s.scanner)
	}

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

	return mux
}
