package aggregator

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/etlmon/etlmon/internal/collector"
)

// MetricBuffer holds metrics for a specific time window
type MetricBuffer struct {
	duration    time.Duration
	windowStart time.Time
	metrics     []collector.Metric
	mu          sync.RWMutex
}

// NewMetricBuffer creates a new buffer for the given window duration
func NewMetricBuffer(duration time.Duration) *MetricBuffer {
	return &MetricBuffer{
		duration:    duration,
		windowStart: time.Now().Truncate(duration),
		metrics:     make([]collector.Metric, 0),
	}
}

// Add adds a metric to the buffer
func (b *MetricBuffer) Add(m collector.Metric) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metrics = append(b.metrics, m)
}

// Len returns the number of metrics in the buffer
func (b *MetricBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.metrics)
}

// GetValues returns all values for a specific resource/metric/labels key
func (b *MetricBuffer) GetValues(key ResourceMetricKey) []float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var values []float64
	for _, m := range b.metrics {
		if m.ResourceType == key.ResourceType && m.Name == key.MetricName {
			// Match labels (empty string matches empty/nil labels)
			metricLabels := LabelsToString(m.Labels)
			if metricLabels == key.Labels {
				values = append(values, m.Value)
			}
		}
	}
	return values
}

// GetResourceMetricKeys returns all unique (resourceType, metricName, labels) tuples
func (b *MetricBuffer) GetResourceMetricKeys() []ResourceMetricKey {
	b.mu.RLock()
	defer b.mu.RUnlock()

	seen := make(map[ResourceMetricKey]bool)
	var keys []ResourceMetricKey

	for _, m := range b.metrics {
		key := ResourceMetricKey{
			ResourceType: m.ResourceType,
			MetricName:   m.Name,
			Labels:       LabelsToString(m.Labels),
		}
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	return keys
}

// ResourceMetricKey identifies a unique resource/metric/labels combination
type ResourceMetricKey struct {
	ResourceType string
	MetricName   string
	Labels       string // JSON-encoded map, sorted keys for determinism
}

// LabelsToString converts labels map to deterministic JSON string
func LabelsToString(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	// Sort keys for deterministic output
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create ordered map for JSON
	orderedMap := make(map[string]string)
	for _, k := range keys {
		orderedMap[k] = labels[k]
	}

	b, _ := json.Marshal(orderedMap)
	return string(b)
}

// Clear removes all metrics from the buffer
func (b *MetricBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metrics = b.metrics[:0] // Reuse underlying array
}

// WindowStart returns the start time of the current window
func (b *MetricBuffer) WindowStart() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.windowStart
}

// WindowEnd returns the end time of the current window
func (b *MetricBuffer) WindowEnd() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.windowStart.Add(b.duration)
}

// IsWindowComplete returns true if the window has ended
func (b *MetricBuffer) IsWindowComplete(now time.Time) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return now.After(b.windowStart.Add(b.duration)) || now.Equal(b.windowStart.Add(b.duration))
}

// ResetWindow resets the buffer for a new window
func (b *MetricBuffer) ResetWindow(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.windowStart = now.Truncate(b.duration)
	b.metrics = b.metrics[:0]
}

// Duration returns the window duration
func (b *MetricBuffer) Duration() time.Duration {
	return b.duration
}
