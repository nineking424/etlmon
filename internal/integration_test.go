package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/storage"
)

// TestCollectAggregatePersist tests the full pipeline:
// collect → aggregate → persist
func TestCollectAggregatePersist(t *testing.T) {
	// Create temp directory for database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create storage
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Create aggregator with short windows for testing
	windows := []time.Duration{500 * time.Millisecond, time.Second}
	aggTypes := []string{"avg", "max", "min"}
	agg := aggregator.NewAggregator(windows, aggTypes)

	// Create collector manager
	collectorMgr := collector.NewManager(100 * time.Millisecond)
	collectorMgr.Register(collector.NewCPUCollector())
	collectorMgr.Register(collector.NewMemoryCollector())

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Metrics channel
	metricsChan := make(chan []collector.Metric, 100)

	// Start collector
	go func() {
		collectorMgr.Start(ctx, func(metrics []collector.Metric) {
			select {
			case metricsChan <- metrics:
			default:
			}
		})
	}()

	// Process metrics and check windows
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var totalResults int

	for {
		select {
		case <-ctx.Done():
			goto done
		case metrics := <-metricsChan:
			// Add to aggregator
			for _, m := range metrics {
				agg.Add(m)
			}
		case <-ticker.C:
			// Check for completed windows
			results := agg.CheckWindows(time.Now())
			if len(results) > 0 {
				// Persist to storage
				batch := make([]*storage.AggregatedMetric, len(results))
				for i, r := range results {
					batch[i] = &storage.AggregatedMetric{
						Timestamp:       r.Timestamp.Unix(),
						ResourceType:    r.ResourceType,
						MetricName:      r.MetricName,
						AggregatedValue: r.Value,
						WindowSize:      formatDuration(r.WindowSize),
						AggregationType: r.AggregationType,
					}
				}
				if err := store.SaveBatch(batch); err != nil {
					t.Errorf("Failed to save batch: %v", err)
				}
				totalResults += len(results)
			}
		}
	}

done:
	// Verify we collected and persisted some metrics
	if totalResults == 0 {
		t.Error("No aggregation results were generated")
	}

	// Query stored metrics
	metrics, err := store.GetMetrics(storage.GetMetricsOptions{})
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("No metrics were persisted to storage")
	}

	t.Logf("Collected and persisted %d aggregation results", len(metrics))

	// Verify we have different resource types
	resourceTypes := make(map[string]bool)
	for _, m := range metrics {
		resourceTypes[m.ResourceType] = true
	}

	if !resourceTypes["cpu"] {
		t.Error("No CPU metrics found")
	}
	if !resourceTypes["memory"] {
		t.Error("No memory metrics found")
	}
}

// TestConfigLoadAndValidate tests configuration loading
func TestConfigLoadAndValidate(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
interval: 10s
resources:
  - cpu
  - memory
  - disk
windows:
  - 1m
  - 5m
aggregations:
  - avg
  - max
database:
  path: ./test.db
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	if cfg.Interval != 10*time.Second {
		t.Errorf("Interval = %v, want 10s", cfg.Interval)
	}

	if len(cfg.Resources) != 3 {
		t.Errorf("len(Resources) = %d, want 3", len(cfg.Resources))
	}
}

// TestStorageRoundTrip tests save and retrieve
func TestStorageRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Save metrics
	now := time.Now()
	metrics := []*storage.AggregatedMetric{
		{Timestamp: now.Unix(), ResourceType: "cpu", MetricName: "usage_percent", AggregatedValue: 45.5, WindowSize: "1m", AggregationType: "avg"},
		{Timestamp: now.Unix(), ResourceType: "memory", MetricName: "usage_percent", AggregatedValue: 60.0, WindowSize: "1m", AggregationType: "avg"},
	}

	if err := store.SaveBatch(metrics); err != nil {
		t.Fatalf("Failed to save batch: %v", err)
	}

	// Retrieve and verify
	retrieved, err := store.GetMetrics(storage.GetMetricsOptions{})
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(retrieved) != 2 {
		t.Fatalf("Retrieved %d metrics, want 2", len(retrieved))
	}

	// Verify filtering
	cpuMetrics, err := store.GetMetrics(storage.GetMetricsOptions{ResourceType: "cpu"})
	if err != nil {
		t.Fatalf("Failed to get CPU metrics: %v", err)
	}

	if len(cpuMetrics) != 1 {
		t.Errorf("CPU metrics count = %d, want 1", len(cpuMetrics))
	}
}

// TestAggregationAccuracy tests that aggregation produces correct values
func TestAggregationAccuracy(t *testing.T) {
	windows := []time.Duration{100 * time.Millisecond}
	aggTypes := []string{"avg", "max", "min", "last"}
	agg := aggregator.NewAggregator(windows, aggTypes)

	// Add known values
	baseTime := time.Now().Truncate(100 * time.Millisecond)
	values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}

	for i, v := range values {
		agg.Add(collector.Metric{
			ResourceType: "test",
			Name:         "value",
			Value:        v,
			Timestamp:    baseTime.Add(time.Duration(i) * 10 * time.Millisecond),
		})
	}

	// Wait for window to complete
	time.Sleep(150 * time.Millisecond)
	results := agg.CheckWindows(time.Now())

	// Verify results
	resultMap := make(map[string]float64)
	for _, r := range results {
		resultMap[r.AggregationType] = r.Value
	}

	// Check expected values
	if resultMap["avg"] != 30.0 {
		t.Errorf("avg = %v, want 30.0", resultMap["avg"])
	}
	if resultMap["max"] != 50.0 {
		t.Errorf("max = %v, want 50.0", resultMap["max"])
	}
	if resultMap["min"] != 10.0 {
		t.Errorf("min = %v, want 10.0", resultMap["min"])
	}
	if resultMap["last"] != 50.0 {
		t.Errorf("last = %v, want 50.0", resultMap["last"])
	}
}

// Helper function
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d >= time.Second {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}
