package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/etlmon/etlmon/internal/db/repository"
)

// PathScanner interface for triggering path scans
type PathScanner interface {
	ScanPaths(paths []string) error
}

// ScannerProxy wraps a PathScanner and allows hot-swapping the underlying scanner
type ScannerProxy struct {
	mu      sync.RWMutex
	scanner PathScanner
}

// NewScannerProxy creates a new scanner proxy
func NewScannerProxy() *ScannerProxy {
	return &ScannerProxy{}
}

// ScanPaths delegates to the underlying scanner
func (p *ScannerProxy) ScanPaths(paths []string) error {
	p.mu.RLock()
	s := p.scanner
	p.mu.RUnlock()
	if s == nil {
		return fmt.Errorf("scanner not available")
	}
	return s.ScanPaths(paths)
}

// Update replaces the underlying scanner
func (p *ScannerProxy) Update(scanner PathScanner) {
	p.mu.Lock()
	p.scanner = scanner
	p.mu.Unlock()
}

// Server represents the HTTP API server
type Server struct {
	addr           string
	repo           *repository.Repository
	nodeName       string
	configPath     string
	httpServer     *http.Server
	scannerProxy   *ScannerProxy
	onConfigReload func()
	listener       net.Listener
	mu             sync.RWMutex
}

// NewServer creates a new API server
func NewServer(addr string, repo *repository.Repository, nodeName string, configPath string) *Server {
	return &Server{
		addr:         addr,
		repo:         repo,
		nodeName:     nodeName,
		configPath:   configPath,
		scannerProxy: NewScannerProxy(),
	}
}

// SetPathScanner sets the path scanner for triggering scans
func (s *Server) SetPathScanner(scanner PathScanner) {
	s.scannerProxy.Update(scanner)
}

// SetConfigReloadCallback sets the callback to invoke when config is updated via API
func (s *Server) SetConfigReloadCallback(cb func()) {
	s.onConfigReload = cb
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
