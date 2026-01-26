package aggregator

import (
	"sync"
	"time"

	"github.com/etlmon/etlmon/internal/collector"
)

// AggregationResult holds the result of an aggregation
type AggregationResult struct {
	ResourceType    string
	MetricName      string
	Value           float64
	WindowSize      time.Duration
	AggregationType string
	Timestamp       time.Time // Window end time
}

// Aggregator manages time-window aggregation of metrics
type Aggregator struct {
	buffers  map[time.Duration]*MetricBuffer
	aggTypes []string
	mu       sync.RWMutex
}

// NewAggregator creates a new aggregator with the specified windows and aggregation types
func NewAggregator(windows []time.Duration, aggTypes []string) *Aggregator {
	buffers := make(map[time.Duration]*MetricBuffer)
	for _, w := range windows {
		buffers[w] = NewMetricBuffer(w)
	}

	return &Aggregator{
		buffers:  buffers,
		aggTypes: aggTypes,
	}
}

// Add adds a metric to all window buffers
func (a *Aggregator) Add(m collector.Metric) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, buf := range a.buffers {
		buf.Add(m)
	}
}

// CheckWindows checks if any windows are complete and returns aggregation results
func (a *Aggregator) CheckWindows(now time.Time) []AggregationResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	var results []AggregationResult

	for duration, buf := range a.buffers {
		if buf.IsWindowComplete(now) {
			// Get all unique resource/metric keys
			keys := buf.GetResourceMetricKeys()

			for _, key := range keys {
				values := buf.GetValues(key.ResourceType, key.MetricName)
				if len(values) == 0 {
					continue
				}

				// Apply each aggregation function
				for _, aggType := range a.aggTypes {
					fn, ok := AggregationFuncs[aggType]
					if !ok {
						continue
					}

					results = append(results, AggregationResult{
						ResourceType:    key.ResourceType,
						MetricName:      key.MetricName,
						Value:           fn(values),
						WindowSize:      duration,
						AggregationType: aggType,
						Timestamp:       buf.WindowEnd(),
					})
				}
			}

			// Reset window for next period
			buf.ResetWindow(now)
		}
	}

	return results
}

// GetWindowDurations returns all configured window durations
func (a *Aggregator) GetWindowDurations() []time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	durations := make([]time.Duration, 0, len(a.buffers))
	for d := range a.buffers {
		durations = append(durations, d)
	}
	return durations
}
