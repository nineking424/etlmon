package log

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// LogRepository defines the interface for storing log data
type LogRepository interface {
	SaveLogEntry(ctx context.Context, entry *models.LogEntry) error
	GetLogEntries(ctx context.Context, logName string, limit int) ([]*models.LogEntry, error)
	TrimOldEntries(ctx context.Context, logName string, maxLines int) error
}

// TailerConfig holds configuration for a single log tailer
type TailerConfig struct {
	Name     string
	Path     string
	MaxLines int // max lines to keep in DB (default: 1000)
}

// LogTailer manages tailing multiple log files
type LogTailer struct {
	repo     LogRepository
	configs  []TailerConfig
	interval time.Duration
	states   map[string]*tailState
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// tailState tracks the current position in a log file
type tailState struct {
	offset int64
	size   int64
}

// NewLogTailer creates a new log tailer
func NewLogTailer(repo LogRepository, configs []TailerConfig, interval time.Duration) *LogTailer {
	for i := range configs {
		if configs[i].MaxLines <= 0 {
			configs[i].MaxLines = 1000
		}
	}
	return &LogTailer{
		repo:     repo,
		configs:  configs,
		interval: interval,
		states:   make(map[string]*tailState),
	}
}

// Start begins periodic log tailing
func (t *LogTailer) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.cancel != nil {
		t.mu.Unlock()
		return fmt.Errorf("tailer already started")
	}
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel
	t.mu.Unlock()

	// Initialize states - start from beginning to read existing content
	for _, cfg := range t.configs {
		info, err := os.Stat(cfg.Path)
		if err != nil {
			// File doesn't exist yet, start from 0
			t.states[cfg.Name] = &tailState{offset: 0, size: 0}
			continue
		}
		t.states[cfg.Name] = &tailState{
			offset: 0,
			size:   info.Size(),
		}
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.tailLoop(ctx)
	}()
	return nil
}

// Stop stops the log tailer
func (t *LogTailer) Stop() {
	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.mu.Unlock()
	t.wg.Wait()
}

func (t *LogTailer) tailLoop(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.tailAll(ctx)
		}
	}
}

func (t *LogTailer) tailAll(ctx context.Context) {
	for _, cfg := range t.configs {
		state, ok := t.states[cfg.Name]
		if !ok {
			state = &tailState{}
			t.states[cfg.Name] = state
		}
		_ = t.tailFile(ctx, cfg, state)
	}
}

func (t *LogTailer) tailFile(ctx context.Context, cfg TailerConfig, state *tailState) error {
	info, err := os.Stat(cfg.Path)
	if err != nil {
		return nil // file not available yet
	}

	currentSize := info.Size()

	// Detect rotation: file got smaller
	if currentSize < state.size {
		state.offset = 0
	}
	state.size = currentSize

	// Nothing new to read
	if state.offset >= currentSize {
		return nil
	}

	f, err := os.Open(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", cfg.Path, err)
	}
	defer f.Close()

	// Seek to last known offset
	if _, err := f.Seek(state.offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek in %s: %w", cfg.Path, err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // up to 1MB lines

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry := &models.LogEntry{
			LogName:   cfg.Name,
			LogPath:   cfg.Path,
			Line:      line,
			CreatedAt: time.Now(),
		}

		if err := t.repo.SaveLogEntry(ctx, entry); err != nil {
			return fmt.Errorf("failed to save log entry: %w", err)
		}
	}

	// Update offset to current position
	newOffset, _ := f.Seek(0, io.SeekCurrent)
	state.offset = newOffset

	// Trim old entries
	_ = t.repo.TrimOldEntries(ctx, cfg.Name, cfg.MaxLines)

	return nil
}
