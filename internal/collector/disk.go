package collector

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/disk"

	"github.com/etlmon/etlmon/internal/config"
)

// DiskCollector collects disk metrics
type DiskCollector struct {
	config     config.DiskConfig
	collectors map[config.DiskMethod]DiskMethodCollector
}

// NewDiskCollector creates a disk collector with auto-discover behavior (backward compatible)
func NewDiskCollector() *DiskCollector {
	return NewDiskCollectorWithConfig(config.DiskConfig{})
}

// NewDiskCollectorWithConfig creates a disk collector with explicit configuration
func NewDiskCollectorWithConfig(cfg config.DiskConfig) *DiskCollector {
	collectors := map[config.DiskMethod]DiskMethodCollector{
		config.DiskMethodStats: NewStatsCollector(),
		config.DiskMethodDF:    NewDFCollector(),
		config.DiskMethodDU:    NewDUCollector(),
	}

	return &DiskCollector{
		config:     cfg,
		collectors: collectors,
	}
}

// Type returns the resource type
func (c *DiskCollector) Type() string {
	return "disk"
}

// Collect gathers disk metrics
func (c *DiskCollector) Collect(ctx context.Context) ([]Metric, error) {
	// If no paths configured, use auto-discover (backward compatible)
	if len(c.config.Paths) == 0 {
		return c.collectAutoDiscover(ctx)
	}

	return c.collectConfiguredPaths(ctx)
}

// collectAutoDiscover collects from all discovered partitions (original behavior)
func (c *DiskCollector) collectAutoDiscover(ctx context.Context) ([]Metric, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var metrics []Metric

	for _, p := range partitions {
		// Filter out pseudo-filesystems
		if isPseudoFS(p.Fstype) {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			log.Printf("Warning: failed to get usage for %s: %v", p.Mountpoint, err)
			continue // Skip partitions we can't read
		}

		// Use mount point as label
		labels := map[string]string{
			"mountpoint": p.Mountpoint,
			"device":     p.Device,
			"fstype":     p.Fstype,
			"method":     "stats", // Auto-discover always uses stats
		}

		metrics = append(metrics, Metric{
			ResourceType: "disk",
			Name:         "usage_percent",
			Value:        usage.UsedPercent,
			Timestamp:    now,
			Labels:       labels,
		})

		metrics = append(metrics, Metric{
			ResourceType: "disk",
			Name:         "used_bytes",
			Value:        float64(usage.Used),
			Timestamp:    now,
			Labels:       labels,
		})

		metrics = append(metrics, Metric{
			ResourceType: "disk",
			Name:         "total_bytes",
			Value:        float64(usage.Total),
			Timestamp:    now,
			Labels:       labels,
		})
	}

	return metrics, nil
}

// collectConfiguredPaths collects from configured paths only
func (c *DiskCollector) collectConfiguredPaths(ctx context.Context) ([]Metric, error) {
	now := time.Now()
	var metrics []Metric

	for _, pathCfg := range c.config.Paths {
		method := c.config.GetMethod(pathCfg)
		collector, ok := c.collectors[method]
		if !ok {
			log.Printf("Warning: unknown method %s for path %s, using stats", method, pathCfg.Path)
			collector = c.collectors[config.DiskMethodStats]
		}

		usage, err := collector.Collect(ctx, pathCfg.Path)
		if err != nil {
			log.Printf("Warning: failed to collect disk metrics for %s (method=%s): %v", pathCfg.Path, method, err)
			continue // Skip paths that fail, continue with others
		}

		// Build labels
		labels := map[string]string{
			"path":       pathCfg.Path,
			"mountpoint": usage.Path,
			"method":     string(method),
		}
		if usage.Device != "" {
			labels["device"] = usage.Device
		}
		if usage.FSType != "" {
			labels["fstype"] = usage.FSType
		}

		// Always add used_bytes (available for all methods)
		metrics = append(metrics, Metric{
			ResourceType: "disk",
			Name:         "used_bytes",
			Value:        float64(usage.Used),
			Timestamp:    now,
			Labels:       labels,
		})

		// Only add total_bytes and usage_percent if we have total (not available for du)
		if usage.Total > 0 {
			metrics = append(metrics, Metric{
				ResourceType: "disk",
				Name:         "total_bytes",
				Value:        float64(usage.Total),
				Timestamp:    now,
				Labels:       labels,
			})

			metrics = append(metrics, Metric{
				ResourceType: "disk",
				Name:         "usage_percent",
				Value:        usage.UsedPercent,
				Timestamp:    now,
				Labels:       labels,
			})
		}

		// Add free_bytes if available
		if usage.Free > 0 {
			metrics = append(metrics, Metric{
				ResourceType: "disk",
				Name:         "free_bytes",
				Value:        float64(usage.Free),
				Timestamp:    now,
				Labels:       labels,
			})
		}
	}

	return metrics, nil
}
