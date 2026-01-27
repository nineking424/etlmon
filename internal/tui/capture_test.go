package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
)

// TestCaptureTUIOutput generates ASCII art captures of TUI components
func TestCaptureTUIOutput(t *testing.T) {
	// Create output directory
	outputDir := filepath.Join("..", "..", "docs", "ui")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Sample metrics for RealtimeView
	cpuMetrics := []collector.Metric{
		{
			ResourceType: "cpu",
			Name:         "core0_usage_percent",
			Value:        30.2,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
		{
			ResourceType: "cpu",
			Name:         "core1_usage_percent",
			Value:        55.8,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
		{
			ResourceType: "cpu",
			Name:         "usage_percent",
			Value:        45.5,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
	}

	memoryMetrics := []collector.Metric{
		{
			ResourceType: "memory",
			Name:         "available_bytes",
			Value:        6481819648,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
		{
			ResourceType: "memory",
			Name:         "total_bytes",
			Value:        17179869184,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
		{
			ResourceType: "memory",
			Name:         "usage_percent",
			Value:        62.3,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
		{
			ResourceType: "memory",
			Name:         "used_bytes",
			Value:        10698049536,
			Timestamp:    time.Now(),
			Labels:       map[string]string{},
		},
	}

	diskMetrics := []collector.Metric{
		{
			ResourceType: "disk",
			Name:         "free_bytes",
			Value:        107942625280,
			Timestamp:    time.Now(),
			Labels:       map[string]string{"mountpoint": "/"},
		},
		{
			ResourceType: "disk",
			Name:         "total_bytes",
			Value:        499963174912,
			Timestamp:    time.Now(),
			Labels:       map[string]string{"mountpoint": "/"},
		},
		{
			ResourceType: "disk",
			Name:         "usage_percent",
			Value:        78.5,
			Timestamp:    time.Now(),
			Labels:       map[string]string{"mountpoint": "/"},
		},
		{
			ResourceType: "disk",
			Name:         "used_bytes",
			Value:        392020549632,
			Timestamp:    time.Now(),
			Labels:       map[string]string{"mountpoint": "/"},
		},
	}

	allMetrics := append(cpuMetrics, memoryMetrics...)
	allMetrics = append(allMetrics, diskMetrics...)

	// Test RealtimeView
	t.Run("RealtimeView", func(t *testing.T) {
		view := NewRealtimeView()
		view.Update(allMetrics)
		output := view.GetText()

		outputPath := filepath.Join(outputDir, "realtime_view.txt")
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			t.Fatalf("Failed to write realtime view output: %v", err)
		}
		t.Logf("RealtimeView output written to %s", outputPath)
	})

	// Test HistoryView with sample aggregation results
	t.Run("HistoryView", func(t *testing.T) {
		view := NewHistoryView()

		// Create sample aggregation results at different times
		baseTime := time.Now()
		results := []aggregator.AggregationResult{
			// CPU metrics - 1m window
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           45.5,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           58.2,
				WindowSize:      1 * time.Minute,
				AggregationType: "max",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           32.1,
				WindowSize:      1 * time.Minute,
				AggregationType: "min",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           47.3,
				WindowSize:      1 * time.Minute,
				AggregationType: "last",
				Timestamp:       baseTime,
			},
			// Memory metrics - 1m window
			{
				ResourceType:    "memory",
				MetricName:      "usage_percent",
				Value:           62.3,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "memory",
				MetricName:      "usage_percent",
				Value:           65.8,
				WindowSize:      1 * time.Minute,
				AggregationType: "max",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "memory",
				MetricName:      "used_bytes",
				Value:           10698049536,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			// Disk metrics - 1m window
			{
				ResourceType:    "disk",
				MetricName:      "usage_percent",
				Value:           78.5,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "disk",
				MetricName:      "usage_percent",
				Value:           79.2,
				WindowSize:      1 * time.Minute,
				AggregationType: "max",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "disk",
				MetricName:      "free_bytes",
				Value:           107942625280,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			// Previous minute data
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           42.1,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime.Add(-1 * time.Minute),
			},
			{
				ResourceType:    "memory",
				MetricName:      "usage_percent",
				Value:           61.5,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime.Add(-1 * time.Minute),
			},
			{
				ResourceType:    "disk",
				MetricName:      "usage_percent",
				Value:           78.3,
				WindowSize:      1 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime.Add(-1 * time.Minute),
			},
			// 5 minute window data
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           43.8,
				WindowSize:      5 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "memory",
				MetricName:      "usage_percent",
				Value:           61.9,
				WindowSize:      5 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "disk",
				MetricName:      "usage_percent",
				Value:           78.4,
				WindowSize:      5 * time.Minute,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			// 1 hour window data
			{
				ResourceType:    "cpu",
				MetricName:      "usage_percent",
				Value:           44.2,
				WindowSize:      1 * time.Hour,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "memory",
				MetricName:      "usage_percent",
				Value:           62.0,
				WindowSize:      1 * time.Hour,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
			{
				ResourceType:    "disk",
				MetricName:      "usage_percent",
				Value:           78.2,
				WindowSize:      1 * time.Hour,
				AggregationType: "avg",
				Timestamp:       baseTime,
			},
		}

		view.Update(results)
		output := view.GetText()

		outputPath := filepath.Join(outputDir, "history_view.txt")
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			t.Fatalf("Failed to write history view output: %v", err)
		}
		t.Logf("HistoryView output written to %s", outputPath)
	})

	// Test StatusBar in various states
	t.Run("StatusBar", func(t *testing.T) {
		var outputs []string

		// State 1: Initializing
		bar := NewStatusBar()
		outputs = append(outputs, "=== Initializing ===\n"+bar.GetText()+"\n")

		// State 2: Running with last update
		bar.SetStatus("Running")
		bar.SetLastUpdate(time.Now())
		outputs = append(outputs, "\n=== Running ===\n"+bar.GetText()+"\n")

		// State 3: Collecting
		bar.SetStatus("Collecting metrics")
		bar.SetLastUpdate(time.Now())
		outputs = append(outputs, "\n=== Collecting ===\n"+bar.GetText()+"\n")

		// State 4: Aggregating
		bar.SetStatus("Aggregating data")
		bar.SetLastUpdate(time.Now())
		outputs = append(outputs, "\n=== Aggregating ===\n"+bar.GetText()+"\n")

		// State 5: Persisting
		bar.SetStatus("Persisting to database")
		bar.SetLastUpdate(time.Now())
		outputs = append(outputs, "\n=== Persisting ===\n"+bar.GetText()+"\n")

		// Combine all states
		combinedOutput := ""
		for _, out := range outputs {
			combinedOutput += out
		}

		outputPath := filepath.Join(outputDir, "status_bar.txt")
		if err := os.WriteFile(outputPath, []byte(combinedOutput), 0644); err != nil {
			t.Fatalf("Failed to write status bar output: %v", err)
		}
		t.Logf("StatusBar output written to %s", outputPath)
	})

	t.Log("All TUI captures generated successfully!")
}
