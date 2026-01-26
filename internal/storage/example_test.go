package storage_test

import (
	"fmt"
	"log"
	"time"

	"github.com/etlmon/etlmon/internal/storage"
)

// Example demonstrates basic usage of the storage package
func Example() {
	// Create a new SQLite store
	store, err := storage.NewSQLiteStore("etlmon_example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Initialize the database schema
	if err := store.Initialize(); err != nil {
		log.Fatal(err)
	}

	// Save a single metric
	metric := &storage.AggregatedMetric{
		Timestamp:       time.Now().Unix(),
		ResourceType:    "cpu",
		MetricName:      "usage_percent",
		AggregatedValue: 45.5,
		WindowSize:      "1m",
		AggregationType: "avg",
	}

	if err := store.SaveAggregatedMetric(metric); err != nil {
		log.Fatal(err)
	}

	// Query metrics with filters
	metrics, err := store.GetMetrics(storage.GetMetricsOptions{
		ResourceType: "cpu",
		WindowSize:   "1m",
		Limit:        10,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range metrics {
		fmt.Printf("Resource: %s, Value: %.2f, Window: %s, Type: %s\n",
			m.ResourceType, m.AggregatedValue, m.WindowSize, m.AggregationType)
	}
}

// ExampleSQLiteStore_SaveBatch demonstrates batch insertion
func ExampleSQLiteStore_SaveBatch() {
	tmpDir := fmt.Sprintf("/tmp/etlmon_batch_%d.db", time.Now().UnixNano())
	store, _ := storage.NewSQLiteStore(tmpDir)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	metrics := []*storage.AggregatedMetric{
		{
			Timestamp:       now,
			ResourceType:    "cpu",
			MetricName:      "usage_percent",
			AggregatedValue: 45.5,
			WindowSize:      "1m",
			AggregationType: "avg",
		},
		{
			Timestamp:       now,
			ResourceType:    "memory",
			MetricName:      "usage_percent",
			AggregatedValue: 62.3,
			WindowSize:      "1m",
			AggregationType: "avg",
		},
		{
			Timestamp:       now,
			ResourceType:    "disk",
			MetricName:      "usage_percent",
			AggregatedValue: 78.1,
			WindowSize:      "1m",
			AggregationType: "avg",
		},
	}

	if err := store.SaveBatch(metrics); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Saved 3 metrics in batch")
	// Output: Saved 3 metrics in batch
}

// ExampleSQLiteStore_GetMetrics demonstrates filtering
func ExampleSQLiteStore_GetMetrics() {
	// Use temp directory to avoid file conflicts
	tmpDir := fmt.Sprintf("/tmp/etlmon_filter_%d.db", time.Now().UnixNano())
	store, _ := storage.NewSQLiteStore(tmpDir)
	defer store.Close()
	store.Initialize()

	// Save some test data
	now := time.Now()
	store.SaveAggregatedMetric(&storage.AggregatedMetric{
		Timestamp:       now.Add(-2 * time.Hour).Unix(),
		ResourceType:    "cpu",
		MetricName:      "usage_percent",
		AggregatedValue: 30.0,
		WindowSize:      "1m",
		AggregationType: "avg",
	})
	store.SaveAggregatedMetric(&storage.AggregatedMetric{
		Timestamp:       now.Unix(),
		ResourceType:    "cpu",
		MetricName:      "usage_percent",
		AggregatedValue: 45.5,
		WindowSize:      "1m",
		AggregationType: "avg",
	})

	// Query metrics from last hour only
	startTime := now.Add(-1 * time.Hour).Unix()
	metrics, _ := store.GetMetrics(storage.GetMetricsOptions{
		ResourceType: "cpu",
		StartTime:    startTime,
	})

	fmt.Printf("Found %d metrics from the last hour\n", len(metrics))
	// Output: Found 1 metrics from the last hour
}

// ExampleSQLiteStore_GetLatestMetrics demonstrates getting recent metrics
func ExampleSQLiteStore_GetLatestMetrics() {
	tmpDir := fmt.Sprintf("/tmp/etlmon_latest_%d.db", time.Now().UnixNano())
	store, _ := storage.NewSQLiteStore(tmpDir)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	// Save metrics at different times
	for i := 0; i < 5; i++ {
		store.SaveAggregatedMetric(&storage.AggregatedMetric{
			Timestamp:       now + int64(i*60),
			ResourceType:    "cpu",
			MetricName:      "usage_percent",
			AggregatedValue: float64(40 + i),
			WindowSize:      "1m",
			AggregationType: "avg",
		})
	}

	// Get latest metrics
	metrics, _ := store.GetLatestMetrics("cpu", "1m")

	if len(metrics) > 0 {
		fmt.Printf("Latest CPU usage: %.1f%%\n", metrics[0].AggregatedValue)
		// Output: Latest CPU usage: 44.0%
	}
}
