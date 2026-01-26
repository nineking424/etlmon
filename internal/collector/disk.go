package collector

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskCollector collects disk metrics
type DiskCollector struct{}

// NewDiskCollector creates a new disk collector
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

// Type returns the resource type
func (c *DiskCollector) Type() string {
	return "disk"
}

// Collect gathers disk metrics
func (c *DiskCollector) Collect(ctx context.Context) ([]Metric, error) {
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
			continue // Skip partitions we can't read
		}

		// Use mount point as label
		labels := map[string]string{
			"mountpoint": p.Mountpoint,
			"device":     p.Device,
			"fstype":     p.Fstype,
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
