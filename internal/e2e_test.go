package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/storage"
)

// TestE2E_FullPipeline tests the complete etlmon pipeline without UI
// This simulates what the main application does:
// 1. Load config
// 2. Initialize storage
// 3. Start collectors
// 4. Aggregate metrics
// 5. Persist to database
// 6. Query results
func TestE2E_FullPipeline(t *testing.T) {
	// Create temp directory for test artifacts
	tmpDir := t.TempDir()

	// Create test config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
interval: 200ms
resources:
  - cpu
  - memory
windows:
  - 500ms
  - 1s
aggregations:
  - avg
  - max
  - min
  - last
database:
  path: ` + filepath.Join(tmpDir, "test.db") + `
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Step 1: Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}
	t.Logf("Config loaded: interval=%v, resources=%v, windows=%v", cfg.Interval, cfg.Resources, cfg.Windows)

	// Step 2: Initialize storage
	store, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	t.Log("Storage initialized")

	// Step 3: Setup aggregator
	windows, err := cfg.GetWindowDurations()
	if err != nil {
		t.Fatalf("Failed to get window durations: %v", err)
	}
	agg := aggregator.NewAggregator(windows, cfg.Aggregations)
	t.Logf("Aggregator created with %d windows and %d aggregation types", len(windows), len(cfg.Aggregations))

	// Step 4: Setup collectors
	collectorMgr := collector.NewManager(cfg.Interval)
	for _, res := range cfg.Resources {
		switch res {
		case "cpu":
			collectorMgr.Register(collector.NewCPUCollector())
		case "memory":
			collectorMgr.Register(collector.NewMemoryCollector())
		case "disk":
			collectorMgr.Register(collector.NewDiskCollector())
		}
	}
	t.Logf("Registered %d collectors", len(cfg.Resources))

	// Step 5: Run the pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	metricsChan := make(chan []collector.Metric, 100)
	var collectionCount int
	var persistedCount int

	// Start collector
	go func() {
		collectorMgr.Start(ctx, func(metrics []collector.Metric) {
			collectionCount++
			select {
			case metricsChan <- metrics:
			default:
			}
		})
	}()

	// Process metrics
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			goto done
		case metrics := <-metricsChan:
			t.Logf("Received %d metrics from collectors", len(metrics))
			// Add to aggregator
			for _, m := range metrics {
				agg.Add(m)
			}
		case <-ticker.C:
			// Check for completed windows
			results := agg.CheckWindows(time.Now())
			if len(results) > 0 {
				t.Logf("Window completed with %d aggregation results", len(results))
				// Convert and persist
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
				} else {
					persistedCount += len(results)
				}
			}
		}
	}

done:
	// Step 6: Verify results
	t.Logf("Pipeline stats: collections=%d, persisted=%d", collectionCount, persistedCount)

	if collectionCount == 0 {
		t.Error("No metrics were collected")
	}
	if persistedCount == 0 {
		t.Error("No metrics were persisted")
	}

	// Query all metrics
	allMetrics, err := store.GetMetrics(storage.GetMetricsOptions{})
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	t.Logf("Retrieved %d metrics from storage", len(allMetrics))

	if len(allMetrics) == 0 {
		t.Fatal("No metrics in database")
	}

	// Verify we have all expected resource types
	resourceCounts := make(map[string]int)
	aggTypeCounts := make(map[string]int)
	windowCounts := make(map[string]int)

	for _, m := range allMetrics {
		resourceCounts[m.ResourceType]++
		aggTypeCounts[m.AggregationType]++
		windowCounts[m.WindowSize]++
	}

	t.Log("Resource distribution:")
	for res, count := range resourceCounts {
		t.Logf("  %s: %d metrics", res, count)
	}

	t.Log("Aggregation type distribution:")
	for aggType, count := range aggTypeCounts {
		t.Logf("  %s: %d metrics", aggType, count)
	}

	t.Log("Window size distribution:")
	for window, count := range windowCounts {
		t.Logf("  %s: %d metrics", window, count)
	}

	// Verify each configured resource has metrics
	for _, res := range cfg.Resources {
		if resourceCounts[res] == 0 {
			t.Errorf("No metrics found for resource: %s", res)
		}
	}

	// Verify each configured aggregation type exists
	for _, aggType := range cfg.Aggregations {
		if aggTypeCounts[aggType] == 0 {
			t.Errorf("No metrics found for aggregation type: %s", aggType)
		}
	}

	// Test filtering by resource type
	cpuMetrics, err := store.GetMetrics(storage.GetMetricsOptions{
		ResourceType: "cpu",
		WindowSize:   "500ms",
	})
	if err != nil {
		t.Fatalf("Failed to get CPU metrics: %v", err)
	}
	t.Logf("CPU metrics (500ms window): %d", len(cpuMetrics))

	// Verify all retrieved metrics match the filter
	for _, m := range cpuMetrics {
		if m.ResourceType != "cpu" {
			t.Errorf("Expected CPU metric, got: %s", m.ResourceType)
		}
		if m.WindowSize != "500ms" {
			t.Errorf("Expected 500ms window, got: %s", m.WindowSize)
		}
	}

	// Test time range filtering
	now := time.Now().Unix()
	oneMinuteAgo := now - 60
	recentMetrics, err := store.GetMetrics(storage.GetMetricsOptions{
		StartTime: oneMinuteAgo,
		EndTime:   now,
	})
	if err != nil {
		t.Fatalf("Failed to get recent metrics: %v", err)
	}
	t.Logf("Recent metrics (last 60s): %d", len(recentMetrics))

	// Verify timestamp filtering
	for _, m := range recentMetrics {
		if m.Timestamp < oneMinuteAgo || m.Timestamp > now {
			t.Errorf("Metric timestamp %d outside range [%d, %d]", m.Timestamp, oneMinuteAgo, now)
		}
	}
}

// TestE2E_LongRunning tests the pipeline over a longer duration to ensure
// multiple windows complete and are properly handled
func TestE2E_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize storage
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Create aggregator with multiple windows
	windows := []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second}
	aggTypes := []string{"avg", "max", "min"}
	agg := aggregator.NewAggregator(windows, aggTypes)

	// Create collectors
	collectorMgr := collector.NewManager(200 * time.Millisecond)
	collectorMgr.Register(collector.NewCPUCollector())
	collectorMgr.Register(collector.NewMemoryCollector())

	// Run for 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	// Track window completions
	windowCompletions := make(map[string]int)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			goto done
		case metrics := <-metricsChan:
			for _, m := range metrics {
				agg.Add(m)
			}
		case <-ticker.C:
			results := agg.CheckWindows(time.Now())
			if len(results) > 0 {
				// Track completions by window
				for _, r := range results {
					windowKey := formatDuration(r.WindowSize)
					windowCompletions[windowKey]++
				}

				// Persist
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
			}
		}
	}

done:
	t.Log("Window completion counts:")
	for window, count := range windowCompletions {
		t.Logf("  %s: %d completions", window, count)
	}

	// Verify each window completed at least once
	for _, window := range windows {
		windowStr := formatDuration(window)
		if windowCompletions[windowStr] == 0 {
			t.Errorf("Window %s never completed", windowStr)
		}
	}

	// Verify metrics in database
	allMetrics, err := store.GetMetrics(storage.GetMetricsOptions{})
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	t.Logf("Total metrics in database: %d", len(allMetrics))

	if len(allMetrics) == 0 {
		t.Fatal("No metrics were persisted")
	}

	// Verify we have metrics for each window size
	windowMetricCounts := make(map[string]int)
	for _, m := range allMetrics {
		windowMetricCounts[m.WindowSize]++
	}

	for _, window := range windows {
		windowStr := formatDuration(window)
		if windowMetricCounts[windowStr] == 0 {
			t.Errorf("No metrics found for window: %s", windowStr)
		}
	}
}

// TestE2E_ConfigValidation tests configuration validation edge cases
func TestE2E_ConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid_minimal",
			config: `
interval: 10s
resources: [cpu]
windows: [1m]
aggregations: [avg]
`,
			wantError: false,
		},
		{
			name: "invalid_interval",
			config: `
interval: -10s
resources: [cpu]
windows: [1m]
aggregations: [avg]
`,
			wantError: true,
			errorMsg:  "interval",
		},
		{
			name: "invalid_resource",
			config: `
interval: 10s
resources: [invalid]
windows: [1m]
aggregations: [avg]
`,
			wantError: true,
			errorMsg:  "invalid resource",
		},
		{
			name: "invalid_window",
			config: `
interval: 10s
resources: [cpu]
windows: [invalid]
aggregations: [avg]
`,
			wantError: true,
			errorMsg:  "invalid window",
		},
		{
			name: "invalid_aggregation",
			config: `
interval: 10s
resources: [cpu]
windows: [1m]
aggregations: [invalid]
`,
			wantError: true,
			errorMsg:  "invalid aggregation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				if !tt.wantError {
					t.Fatalf("Unexpected error loading config: %v", err)
				}
				return
			}

			err = cfg.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}
