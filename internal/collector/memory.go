package collector

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCollector collects memory metrics
type MemoryCollector struct{}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

// Type returns the resource type
func (c *MemoryCollector) Type() string {
	return "memory"
}

// Collect gathers memory metrics
func (c *MemoryCollector) Collect(ctx context.Context) ([]Metric, error) {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	metrics := []Metric{
		{
			ResourceType: "memory",
			Name:         "usage_percent",
			Value:        v.UsedPercent,
			Timestamp:    now,
		},
		{
			ResourceType: "memory",
			Name:         "used_bytes",
			Value:        float64(v.Used),
			Timestamp:    now,
		},
		{
			ResourceType: "memory",
			Name:         "available_bytes",
			Value:        float64(v.Available),
			Timestamp:    now,
		},
		{
			ResourceType: "memory",
			Name:         "total_bytes",
			Value:        float64(v.Total),
			Timestamp:    now,
		},
	}

	return metrics, nil
}
