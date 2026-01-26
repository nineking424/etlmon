package collector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Metric represents a single collected metric
type Metric struct {
	ResourceType string            // cpu, memory, disk
	Name         string            // metric name (e.g., usage_percent)
	Value        float64           // metric value
	Timestamp    time.Time         // when the metric was collected
	Labels       map[string]string // optional labels (e.g., disk mount point)
}

// Collector interface for all metric collectors
type Collector interface {
	// Type returns the resource type (cpu, memory, disk)
	Type() string
	// Collect gathers metrics and returns them
	Collect(ctx context.Context) ([]Metric, error)
}

// Manager manages multiple collectors
type Manager struct {
	collectors []Collector
	interval   time.Duration
	mu         sync.RWMutex
}

// NewManager creates a new collector manager
func NewManager(interval time.Duration) *Manager {
	return &Manager{
		collectors: make([]Collector, 0),
		interval:   interval,
	}
}

// Register adds a collector to the manager
func (m *Manager) Register(c Collector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectors = append(m.collectors, c)
}

// Start begins collecting metrics at the configured interval
// It calls the handler function with collected metrics
func (m *Manager) Start(ctx context.Context, handler func([]Metric)) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Collect immediately on start
	metrics, err := m.CollectOnce(ctx)
	if err == nil && len(metrics) > 0 {
		handler(metrics)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := m.CollectOnce(ctx)
			if err != nil {
				continue // Log error in production
			}
			if len(metrics) > 0 {
				handler(metrics)
			}
		}
	}
}

// CollectOnce collects metrics from all collectors once
func (m *Manager) CollectOnce(ctx context.Context) ([]Metric, error) {
	m.mu.RLock()
	collectors := make([]Collector, len(m.collectors))
	copy(collectors, m.collectors)
	m.mu.RUnlock()

	var allMetrics []Metric
	var errs []error

	for _, c := range collectors {
		metrics, err := c.Collect(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s collector: %w", c.Type(), err))
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}

	if len(errs) > 0 && len(allMetrics) == 0 {
		return nil, fmt.Errorf("all collectors failed: %v", errs)
	}

	return allMetrics, nil
}

// Pseudo-filesystem types to filter out for disk collector
var pseudoFSTypes = map[string]bool{
	"tmpfs":      true,
	"devtmpfs":   true,
	"sysfs":      true,
	"proc":       true,
	"overlay":    true,
	"squashfs":   true,
	"aufs":       true,
	"debugfs":    true,
	"cgroup":     true,
	"cgroup2":    true,
	"pstore":     true,
	"bpf":        true,
	"tracefs":    true,
	"hugetlbfs":  true,
	"mqueue":     true,
	"fusectl":    true,
	"configfs":   true,
	"efivarfs":   true,
	"nsfs":       true,
	"devpts":     true,
	"securityfs": true,
	"autofs":     true,
}

// isPseudoFS checks if a filesystem type is a pseudo-filesystem
func isPseudoFS(fstype string) bool {
	return pseudoFSTypes[fstype]
}
