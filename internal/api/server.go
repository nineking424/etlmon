package api

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/etlmon/etlmon/internal/db/repository"
)

// PathScanner interface for triggering path scans
type PathScanner interface {
	ScanPaths(paths []string) error
}

// Server represents the HTTP API server
type Server struct {
	addr       string
	repo       *repository.Repository
	nodeName   string
	httpServer *http.Server
	scanner    PathScanner
	listener   net.Listener
	mu         sync.RWMutex
}

// NewServer creates a new API server
func NewServer(addr string, repo *repository.Repository, nodeName string) *Server {
	return &Server{
		addr:     addr,
		repo:     repo,
		nodeName: nodeName,
	}
}

// SetPathScanner sets the path scanner for triggering scans
func (s *Server) SetPathScanner(scanner PathScanner) {
	s.scanner = scanner
}

// Start starts the HTTP server (blocking)
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.httpServer = &http.Server{
		Handler: s.setupRoutes(),
	}
	s.mu.Unlock()

	return s.httpServer.Serve(listener)
}

// Addr returns the address the server is listening on
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	srv := s.httpServer
	s.mu.RUnlock()

	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}
