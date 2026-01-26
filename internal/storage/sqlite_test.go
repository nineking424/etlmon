package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSQLiteStore_CreateDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer store.Close()

	// Check that DB file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNewSQLiteStore_InvalidPath(t *testing.T) {
	// Try to create DB in non-existent directory without permission
	_, err := NewSQLiteStore("/nonexistent/dir/test.db")
	if err == nil {
		t.Error("NewSQLiteStore() expected error for invalid path")
	}
}

func TestInitialize_CreatesSchema(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	err := store.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify table exists by querying it
	_, err = store.db.Exec("SELECT * FROM aggregated_metrics LIMIT 1")
	if err != nil {
		t.Errorf("Table aggregated_metrics not created: %v", err)
	}
}

func TestInitialize_Idempotent(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Initialize twice should not error
	if err := store.Initialize(); err != nil {
		t.Fatalf("First Initialize() error = %v", err)
	}
	if err := store.Initialize(); err != nil {
		t.Fatalf("Second Initialize() error = %v", err)
	}
}

func TestSaveAggregatedMetric(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	metric := &AggregatedMetric{
		Timestamp:       time.Now().Unix(),
		ResourceType:    "cpu",
		MetricName:      "usage_percent",
		AggregatedValue: 45.5,
		WindowSize:      "1m",
		AggregationType: "avg",
	}

	err := store.SaveAggregatedMetric(metric)
	if err != nil {
		t.Fatalf("SaveAggregatedMetric() error = %v", err)
	}

	// Verify it was saved
	metrics, err := store.GetMetrics(GetMetricsOptions{})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].AggregatedValue != 45.5 {
		t.Errorf("AggregatedValue = %v, want 45.5", metrics[0].AggregatedValue)
	}
}

func TestSaveBatch(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	metrics := []*AggregatedMetric{
		{Timestamp: time.Now().Unix(), ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"},
		{Timestamp: time.Now().Unix(), ResourceType: "memory", MetricName: "usage", AggregatedValue: 20.0, WindowSize: "1m", AggregationType: "avg"},
		{Timestamp: time.Now().Unix(), ResourceType: "disk", MetricName: "usage", AggregatedValue: 30.0, WindowSize: "1m", AggregationType: "avg"},
	}

	err := store.SaveBatch(metrics)
	if err != nil {
		t.Fatalf("SaveBatch() error = %v", err)
	}

	// Verify all were saved
	result, err := store.GetMetrics(GetMetricsOptions{})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(result))
	}
}

func TestGetMetrics_FilterByResourceType(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"})
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "memory", MetricName: "usage", AggregatedValue: 20.0, WindowSize: "1m", AggregationType: "avg"})

	metrics, err := store.GetMetrics(GetMetricsOptions{ResourceType: "cpu"})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].ResourceType != "cpu" {
		t.Errorf("ResourceType = %s, want cpu", metrics[0].ResourceType)
	}
}

func TestGetMetrics_FilterByWindowSize(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"})
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 20.0, WindowSize: "5m", AggregationType: "avg"})

	metrics, err := store.GetMetrics(GetMetricsOptions{WindowSize: "5m"})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].WindowSize != "5m" {
		t.Errorf("WindowSize = %s, want 5m", metrics[0].WindowSize)
	}
}

func TestGetMetrics_FilterByTimeRange(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now()
	old := now.Add(-2 * time.Hour)

	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: old.Unix(), ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"})
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now.Unix(), ResourceType: "cpu", MetricName: "usage", AggregatedValue: 20.0, WindowSize: "1m", AggregationType: "avg"})

	startTime := now.Add(-1 * time.Hour).Unix()
	metrics, err := store.GetMetrics(GetMetricsOptions{StartTime: startTime})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].AggregatedValue != 20.0 {
		t.Errorf("AggregatedValue = %v, want 20.0", metrics[0].AggregatedValue)
	}
}

func TestGetMetrics_FilterByAggregationType(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"})
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 50.0, WindowSize: "1m", AggregationType: "max"})

	metrics, err := store.GetMetrics(GetMetricsOptions{AggregationType: "max"})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].AggregationType != "max" {
		t.Errorf("AggregationType = %s, want max", metrics[0].AggregationType)
	}
}

func TestGetMetrics_Limit(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		store.SaveAggregatedMetric(&AggregatedMetric{
			Timestamp: now + int64(i), ResourceType: "cpu", MetricName: "usage",
			AggregatedValue: float64(i), WindowSize: "1m", AggregationType: "avg",
		})
	}

	metrics, err := store.GetMetrics(GetMetricsOptions{Limit: 5})
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(metrics))
	}
}

func TestGetLatestMetrics(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	store.Initialize()

	now := time.Now().Unix()
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now - 60, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 10.0, WindowSize: "1m", AggregationType: "avg"})
	store.SaveAggregatedMetric(&AggregatedMetric{Timestamp: now, ResourceType: "cpu", MetricName: "usage", AggregatedValue: 20.0, WindowSize: "1m", AggregationType: "avg"})

	metrics, err := store.GetLatestMetrics("cpu", "1m")
	if err != nil {
		t.Fatalf("GetLatestMetrics() error = %v", err)
	}
	if len(metrics) == 0 {
		t.Fatal("Expected at least 1 metric")
	}
	// Should return the latest one
	if metrics[0].AggregatedValue != 20.0 {
		t.Errorf("AggregatedValue = %v, want 20.0", metrics[0].AggregatedValue)
	}
}

func TestClose(t *testing.T) {
	store := newTestStore(t)

	err := store.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should not error (idempotent)
	err = store.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

// Helper function
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	return store
}
