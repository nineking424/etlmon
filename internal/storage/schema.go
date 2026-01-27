package storage

// AggregatedMetric represents a single aggregated metric record
type AggregatedMetric struct {
	ID              int64   `json:"id,omitempty"`
	Timestamp       int64   `json:"timestamp"`        // Unix timestamp (window end time)
	ResourceType    string  `json:"resource_type"`    // cpu, memory, disk
	MetricName      string  `json:"metric_name"`      // usage_percent, etc.
	AggregatedValue float64 `json:"aggregated_value"` // The aggregated value
	WindowSize      string  `json:"window_size"`      // 1m, 5m, 1h
	AggregationType string  `json:"aggregation_type"` // avg, max, min, last
	Labels          string  `json:"labels,omitempty"` // JSON-encoded labels map
}

// GetMetricsOptions defines filtering options for querying metrics
type GetMetricsOptions struct {
	ResourceType    string // Filter by resource type
	MetricName      string // Filter by metric name
	WindowSize      string // Filter by window size
	AggregationType string // Filter by aggregation type
	StartTime       int64  // Filter by start time (Unix)
	EndTime         int64  // Filter by end time (Unix)
	Limit           int    // Maximum number of results
	Labels          string // Filter by labels (exact match)
}

// Schema SQL
const createTableSQL = `
CREATE TABLE IF NOT EXISTS aggregated_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    resource_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    aggregated_value REAL NOT NULL,
    window_size TEXT NOT NULL,
    aggregation_type TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON aggregated_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_metrics_resource ON aggregated_metrics(resource_type);
CREATE INDEX IF NOT EXISTS idx_metrics_window ON aggregated_metrics(window_size);
CREATE INDEX IF NOT EXISTS idx_metrics_composite ON aggregated_metrics(resource_type, window_size, timestamp);
`
