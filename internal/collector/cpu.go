package collector

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUCollector collects CPU metrics
type CPUCollector struct {
	// Use 200ms measurement interval for accurate readings
	measurementInterval time.Duration
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		measurementInterval: 200 * time.Millisecond,
	}
}

// Type returns the resource type
func (c *CPUCollector) Type() string {
	return "cpu"
}

// Collect gathers CPU metrics
func (c *CPUCollector) Collect(ctx context.Context) ([]Metric, error) {
	// Get CPU usage percentage (use interval > 0 for accurate readings)
	percentages, err := cpu.PercentWithContext(ctx, c.measurementInterval, false)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	metrics := make([]Metric, 0, len(percentages)+1)

	// Overall CPU usage (average across all cores)
	if len(percentages) > 0 {
		var total float64
		for _, p := range percentages {
			total += p
		}
		avg := total / float64(len(percentages))

		metrics = append(metrics, Metric{
			ResourceType: "cpu",
			Name:         "usage_percent",
			Value:        avg,
			Timestamp:    now,
		})
	}

	return metrics, nil
}
