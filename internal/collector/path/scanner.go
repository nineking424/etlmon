package path

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// PathsRepository defines the interface for storing path statistics
type PathsRepository interface {
	SavePathStats(ctx context.Context, stats *models.PathStats) error
	GetLatestPathStats(ctx context.Context) ([]*models.PathStats, error)
	GetPathStats(ctx context.Context, path string) (*models.PathStats, error)
}

// PathConfig represents configuration for a monitored path
type PathConfig struct {
	Path         string
	ScanInterval time.Duration
	MaxDepth     int
	Exclude      []string
	Timeout      time.Duration
}

// PathScanner monitors filesystem paths and collects statistics
type PathScanner struct {
	repo      PathsRepository
	paths     []PathConfig
	scanning  map[string]bool // Track which paths are currently being scanned
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
}

// NewPathScanner creates a new path scanner
func NewPathScanner(repo PathsRepository, paths []PathConfig) *PathScanner {
	return &PathScanner{
		repo:     repo,
		paths:    paths,
		scanning: make(map[string]bool),
	}
}

// Start begins periodic path scanning
func (s *PathScanner) Start(ctx context.Context) {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	// Start a goroutine for each configured path
	for _, cfg := range s.paths {
		s.wg.Add(1)
		go func(pathCfg PathConfig) {
			defer s.wg.Done()
			s.scanLoop(ctx, pathCfg)
		}(cfg)
	}
}

// Stop stops all path scanning
func (s *PathScanner) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.mu.Unlock()

	s.wg.Wait()
}

// ScanPath performs a single scan of a path
func (s *PathScanner) ScanPath(ctx context.Context, cfg PathConfig) (*models.PathStats, error) {
	// Check if already scanning this path
	s.mu.Lock()
	if s.scanning[cfg.Path] {
		s.mu.Unlock()
		return nil, fmt.Errorf("path %s is already being scanned", cfg.Path)
	}
	s.scanning[cfg.Path] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.scanning[cfg.Path] = false
		s.mu.Unlock()
	}()

	// Create context with timeout
	scanCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	startTime := time.Now()
	stats := &models.PathStats{
		Path:        cfg.Path,
		Status:      "SCANNING",
		CollectedAt: time.Now(),
	}

	// Perform the scan
	fileCount, dirCount, err := s.walkPath(scanCtx, cfg)
	duration := time.Since(startTime)

	stats.FileCount = fileCount
	stats.DirCount = dirCount
	stats.ScanDurationMs = duration.Milliseconds()

	if err != nil {
		if scanCtx.Err() == context.DeadlineExceeded {
			stats.Status = "ERROR"
			stats.ErrorMessage = "scan timeout exceeded"
		} else {
			stats.Status = "ERROR"
			stats.ErrorMessage = err.Error()
		}
	} else {
		stats.Status = "OK"
	}

	stats.CollectedAt = time.Now()

	return stats, nil
}

// TriggerScan manually triggers a scan for specific paths
func (s *PathScanner) TriggerScan(ctx context.Context, paths []string) error {
	for _, path := range paths {
		// Find the config for this path
		var cfg *PathConfig
		for i := range s.paths {
			if s.paths[i].Path == path {
				cfg = &s.paths[i]
				break
			}
		}

		if cfg == nil {
			return fmt.Errorf("no configuration found for path: %s", path)
		}

		stats, err := s.ScanPath(ctx, *cfg)
		if err != nil {
			return err
		}

		if err := s.repo.SavePathStats(ctx, stats); err != nil {
			return fmt.Errorf("failed to save path stats: %w", err)
		}
	}

	return nil
}

// scanLoop runs the periodic scanning for a single path
func (s *PathScanner) scanLoop(ctx context.Context, cfg PathConfig) {
	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()

	// Scan immediately on start
	stats, _ := s.ScanPath(ctx, cfg)
	if stats != nil {
		_ = s.repo.SavePathStats(ctx, stats)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, _ := s.ScanPath(ctx, cfg)
			if stats != nil {
				_ = s.repo.SavePathStats(ctx, stats)
			}
		}
	}
}

// walkPath walks the directory tree and counts files and directories
func (s *PathScanner) walkPath(ctx context.Context, cfg PathConfig) (fileCount, dirCount int64, err error) {
	var mu sync.Mutex
	baseDepth := strings.Count(cfg.Path, string(filepath.Separator))

	err = filepath.WalkDir(cfg.Path, func(path string, d fs.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Skip paths we can't access
			return nil
		}

		// Skip the root path itself
		if path == cfg.Path {
			return nil
		}

		// Check max depth
		if cfg.MaxDepth > 0 {
			currentDepth := strings.Count(path, string(filepath.Separator))
			relativeDepth := currentDepth - baseDepth
			if relativeDepth > cfg.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check exclude patterns
		if s.shouldExclude(path, cfg.Exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Count files and directories
		mu.Lock()
		if d.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		mu.Unlock()

		return nil
	})

	return fileCount, dirCount, err
}

// shouldExclude checks if a path matches any exclude pattern
func (s *PathScanner) shouldExclude(path string, excludePatterns []string) bool {
	base := filepath.Base(path)

	for _, pattern := range excludePatterns {
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return true
		}
	}

	return false
}
