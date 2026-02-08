package process

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// ProcessRepository defines the interface for storing process data
type ProcessRepository interface {
	SaveProcessInfo(ctx context.Context, info *models.ProcessInfo) error
	GetLatestProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error)
}

// Config holds process monitoring configuration
type Config struct {
	Patterns []string // process name patterns to monitor (empty = top N by CPU)
	TopN     int      // max processes to keep (default: 50)
}

// Collector collects process statistics
type Collector struct {
	repo     ProcessRepository
	interval time.Duration
	config   Config
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// NewCollector creates a new process collector
func NewCollector(repo ProcessRepository, interval time.Duration, cfg Config) *Collector {
	if cfg.TopN <= 0 {
		cfg.TopN = 50
	}
	return &Collector{
		repo:     repo,
		interval: interval,
		config:   cfg,
	}
}

// Start begins periodic process collection
func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		return fmt.Errorf("collector already started")
	}
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.collectLoop(ctx)
	}()
	return nil
}

// Stop stops the process collection
func (c *Collector) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.mu.Unlock()
	c.wg.Wait()
}

func (c *Collector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	_ = c.CollectOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.CollectOnce(ctx)
		}
	}
}

// CollectOnce performs a single collection of process info
func (c *Collector) CollectOnce(ctx context.Context) error {
	procs, err := c.getProcesses()
	if err != nil {
		return fmt.Errorf("failed to get processes: %w", err)
	}

	// Filter by patterns if configured
	if len(c.config.Patterns) > 0 {
		procs = c.filterByPatterns(procs)
	}

	// Limit to TopN
	if len(procs) > c.config.TopN {
		procs = procs[:c.config.TopN]
	}

	for _, proc := range procs {
		if err := c.repo.SaveProcessInfo(ctx, proc); err != nil {
			return fmt.Errorf("failed to save process info for PID %d: %w", proc.PID, err)
		}
	}

	return nil
}

// getProcesses runs ps command and parses output
func (c *Collector) getProcesses() ([]*models.ProcessInfo, error) {
	// ps aux output is sorted by default (varies by OS)
	// Using ps with custom format for consistent cross-platform parsing
	out, err := exec.Command("ps", "-eo", "pid,user,pcpu,rss,state,etime,comm").Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var procs []*models.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	// Skip header line
	if scanner.Scan() {
		// header consumed
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		proc, err := parsePsLine(line)
		if err != nil {
			continue // skip unparseable lines
		}
		procs = append(procs, proc)
	}

	return procs, nil
}

// parsePsLine parses a single line from ps -eo output
func parsePsLine(line string) (*models.ProcessInfo, error) {
	// Fields: PID USER %CPU RSS STATE ELAPSED COMMAND
	// The command may contain spaces, so we split carefully
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return nil, fmt.Errorf("insufficient fields: %s", line)
	}

	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, fmt.Errorf("invalid PID: %s", fields[0])
	}

	cpu, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		cpu = 0
	}

	rss, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		rss = 0
	}

	// Convert state code to human-readable
	status := parseState(fields[4])

	return &models.ProcessInfo{
		PID:         pid,
		User:        fields[1],
		CPUPercent:  cpu,
		MemRSS:      rss * 1024, // ps reports RSS in KB, convert to bytes
		Status:      status,
		Elapsed:     fields[5],
		Name:        strings.Join(fields[6:], " "),
		CollectedAt: time.Now(),
	}, nil
}

// parseState converts ps state codes to human-readable status
func parseState(state string) string {
	if len(state) == 0 {
		return "unknown"
	}
	switch state[0] {
	case 'R':
		return "running"
	case 'S', 'I':
		return "sleeping"
	case 'D':
		return "disk-wait"
	case 'Z':
		return "zombie"
	case 'T':
		return "stopped"
	default:
		return "other"
	}
}

// filterByPatterns filters processes by configured name patterns
func (c *Collector) filterByPatterns(procs []*models.ProcessInfo) []*models.ProcessInfo {
	var filtered []*models.ProcessInfo
	for _, proc := range procs {
		for _, pattern := range c.config.Patterns {
			matched, err := regexp.MatchString(pattern, proc.Name)
			if err == nil && matched {
				filtered = append(filtered, proc)
				break
			}
		}
	}
	return filtered
}
