package aggregator

import (
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/collector"
)

// Test Aggregation Functions
func TestAvg_Empty(t *testing.T) {
	result := Avg([]float64{})
	if result != 0 {
		t.Errorf("Avg([]) = %v, want 0", result)
	}
}

func TestAvg_SingleValue(t *testing.T) {
	result := Avg([]float64{42.0})
	if result != 42.0 {
		t.Errorf("Avg([42.0]) = %v, want 42.0", result)
	}
}

func TestAvg_MultipleValues(t *testing.T) {
	result := Avg([]float64{10.0, 20.0, 30.0})
	if result != 20.0 {
		t.Errorf("Avg([10, 20, 30]) = %v, want 20.0", result)
	}
}

func TestMax_Empty(t *testing.T) {
	result := Max([]float64{})
	if result != 0 {
		t.Errorf("Max([]) = %v, want 0", result)
	}
}

func TestMax_SingleValue(t *testing.T) {
	result := Max([]float64{42.0})
	if result != 42.0 {
		t.Errorf("Max([42.0]) = %v, want 42.0", result)
	}
}

func TestMax_MultipleValues(t *testing.T) {
	result := Max([]float64{10.0, 50.0, 30.0})
	if result != 50.0 {
		t.Errorf("Max([10, 50, 30]) = %v, want 50.0", result)
	}
}

func TestMax_NegativeValues(t *testing.T) {
	result := Max([]float64{-10.0, -5.0, -20.0})
	if result != -5.0 {
		t.Errorf("Max([-10, -5, -20]) = %v, want -5.0", result)
	}
}

func TestMin_Empty(t *testing.T) {
	result := Min([]float64{})
	if result != 0 {
		t.Errorf("Min([]) = %v, want 0", result)
	}
}

func TestMin_SingleValue(t *testing.T) {
	result := Min([]float64{42.0})
	if result != 42.0 {
		t.Errorf("Min([42.0]) = %v, want 42.0", result)
	}
}

func TestMin_MultipleValues(t *testing.T) {
	result := Min([]float64{10.0, 5.0, 30.0})
	if result != 5.0 {
		t.Errorf("Min([10, 5, 30]) = %v, want 5.0", result)
	}
}

func TestLast_Empty(t *testing.T) {
	result := Last([]float64{})
	if result != 0 {
		t.Errorf("Last([]) = %v, want 0", result)
	}
}

func TestLast_SingleValue(t *testing.T) {
	result := Last([]float64{42.0})
	if result != 42.0 {
		t.Errorf("Last([42.0]) = %v, want 42.0", result)
	}
}

func TestLast_MultipleValues(t *testing.T) {
	result := Last([]float64{10.0, 20.0, 30.0})
	if result != 30.0 {
		t.Errorf("Last([10, 20, 30]) = %v, want 30.0", result)
	}
}

// Test MetricBuffer
func TestMetricBuffer_Add(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)

	m := collector.Metric{
		ResourceType: "cpu",
		Name:         "usage_percent",
		Value:        45.5,
		Timestamp:    time.Now(),
	}

	buf.Add(m)

	if buf.Len() != 1 {
		t.Errorf("Len() = %d, want 1", buf.Len())
	}
}

func TestMetricBuffer_AddMultiple(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)
	now := time.Now()

	for i := 0; i < 5; i++ {
		buf.Add(collector.Metric{
			ResourceType: "cpu",
			Name:         "usage_percent",
			Value:        float64(i * 10),
			Timestamp:    now.Add(time.Duration(i) * time.Second),
		})
	}

	if buf.Len() != 5 {
		t.Errorf("Len() = %d, want 5", buf.Len())
	}
}

func TestMetricBuffer_GetValues(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)
	now := time.Now()

	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 10.0, Timestamp: now})
	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 20.0, Timestamp: now})
	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 30.0, Timestamp: now})

	key := ResourceMetricKey{
		ResourceType: "cpu",
		MetricName:   "usage_percent",
		Labels:       "",
	}
	values := buf.GetValues(key)
	if len(values) != 3 {
		t.Fatalf("len(values) = %d, want 3", len(values))
	}

	expected := []float64{10.0, 20.0, 30.0}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestMetricBuffer_Clear(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)
	now := time.Now()

	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage", Value: 10.0, Timestamp: now})
	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage", Value: 20.0, Timestamp: now})

	if buf.Len() != 2 {
		t.Fatalf("Len() before clear = %d, want 2", buf.Len())
	}

	buf.Clear()

	if buf.Len() != 0 {
		t.Errorf("Len() after clear = %d, want 0", buf.Len())
	}
}

func TestMetricBuffer_WindowStart(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)

	// Window start should be aligned to the minute
	start := buf.WindowStart()

	// Check that it's truncated to minute boundary
	if start.Second() != 0 || start.Nanosecond() != 0 {
		t.Errorf("WindowStart not aligned to minute: %v", start)
	}
}

func TestMetricBuffer_IsWindowComplete_NotComplete(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)

	// Just created, window should not be complete
	if buf.IsWindowComplete(time.Now()) {
		t.Error("IsWindowComplete() = true for new buffer, want false")
	}
}

func TestMetricBuffer_IsWindowComplete_Complete(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)

	// Simulate time passing beyond window
	futureTime := buf.WindowStart().Add(time.Minute + time.Second)

	if !buf.IsWindowComplete(futureTime) {
		t.Error("IsWindowComplete() = false after window duration, want true")
	}
}

func TestMetricBuffer_ResetWindow(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)
	now := time.Now()

	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage", Value: 10.0, Timestamp: now})

	originalStart := buf.WindowStart()

	// Advance time past window
	newTime := originalStart.Add(time.Minute + time.Second)
	buf.ResetWindow(newTime)

	// Should have new window start
	newStart := buf.WindowStart()
	if !newStart.After(originalStart) {
		t.Errorf("WindowStart after reset should be later: got %v, original %v", newStart, originalStart)
	}

	// Should be cleared
	if buf.Len() != 0 {
		t.Errorf("Len() after reset = %d, want 0", buf.Len())
	}
}

// Test Aggregator
func TestNewAggregator(t *testing.T) {
	windows := []time.Duration{time.Minute, 5 * time.Minute}
	aggTypes := []string{"avg", "max"}

	agg := NewAggregator(windows, aggTypes)

	if agg == nil {
		t.Fatal("NewAggregator returned nil")
	}

	// Should have buffers for each window
	if len(agg.buffers) != 2 {
		t.Errorf("len(buffers) = %d, want 2", len(agg.buffers))
	}
}

func TestAggregator_Add(t *testing.T) {
	windows := []time.Duration{time.Minute}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	m := collector.Metric{
		ResourceType: "cpu",
		Name:         "usage_percent",
		Value:        45.5,
		Timestamp:    time.Now(),
	}

	agg.Add(m)

	// Metric should be added to all window buffers
	buf := agg.buffers[time.Minute]
	if buf.Len() != 1 {
		t.Errorf("Buffer len = %d, want 1", buf.Len())
	}
}

func TestAggregator_AddToAllWindows(t *testing.T) {
	windows := []time.Duration{time.Minute, 5 * time.Minute, time.Hour}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	m := collector.Metric{
		ResourceType: "cpu",
		Name:         "usage_percent",
		Value:        45.5,
		Timestamp:    time.Now(),
	}

	agg.Add(m)

	// Metric should be in ALL window buffers
	for _, window := range windows {
		buf := agg.buffers[window]
		if buf.Len() != 1 {
			t.Errorf("Buffer for %v has len = %d, want 1", window, buf.Len())
		}
	}
}

func TestAggregator_CheckWindows_NoComplete(t *testing.T) {
	windows := []time.Duration{time.Minute}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	m := collector.Metric{
		ResourceType: "cpu",
		Name:         "usage_percent",
		Value:        45.5,
		Timestamp:    time.Now(),
	}
	agg.Add(m)

	// Check windows immediately - should return nothing
	results := agg.CheckWindows(time.Now())
	if len(results) != 0 {
		t.Errorf("CheckWindows returned %d results, want 0", len(results))
	}
}

func TestAggregator_CheckWindows_Complete(t *testing.T) {
	windows := []time.Duration{time.Minute}
	aggTypes := []string{"avg", "max", "min", "last"}
	agg := NewAggregator(windows, aggTypes)

	// Add metrics
	now := time.Now().Truncate(time.Minute) // Align to minute
	for i := 0; i < 5; i++ {
		agg.Add(collector.Metric{
			ResourceType: "cpu",
			Name:         "usage_percent",
			Value:        float64(10 + i*10), // 10, 20, 30, 40, 50
			Timestamp:    now.Add(time.Duration(i) * 10 * time.Second),
		})
	}

	// Check after window completes
	futureTime := now.Add(time.Minute + time.Second)
	results := agg.CheckWindows(futureTime)

	// Should have results for each aggregation type
	if len(results) != 4 {
		t.Fatalf("CheckWindows returned %d results, want 4", len(results))
	}

	// Verify values
	for _, r := range results {
		switch r.AggregationType {
		case "avg":
			if r.Value != 30.0 { // (10+20+30+40+50)/5 = 30
				t.Errorf("avg = %v, want 30.0", r.Value)
			}
		case "max":
			if r.Value != 50.0 {
				t.Errorf("max = %v, want 50.0", r.Value)
			}
		case "min":
			if r.Value != 10.0 {
				t.Errorf("min = %v, want 10.0", r.Value)
			}
		case "last":
			if r.Value != 50.0 {
				t.Errorf("last = %v, want 50.0", r.Value)
			}
		}
	}
}

func TestAggregator_MultipleResourceTypes(t *testing.T) {
	windows := []time.Duration{time.Minute}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	now := time.Now().Truncate(time.Minute)

	// Add CPU metrics
	agg.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 20.0, Timestamp: now})
	agg.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 40.0, Timestamp: now})

	// Add Memory metrics
	agg.Add(collector.Metric{ResourceType: "memory", Name: "usage_percent", Value: 50.0, Timestamp: now})
	agg.Add(collector.Metric{ResourceType: "memory", Name: "usage_percent", Value: 70.0, Timestamp: now})

	// Check after window
	futureTime := now.Add(time.Minute + time.Second)
	results := agg.CheckWindows(futureTime)

	// Should have avg for both CPU and Memory
	cpuFound := false
	memFound := false

	for _, r := range results {
		if r.ResourceType == "cpu" && r.MetricName == "usage_percent" {
			cpuFound = true
			if r.Value != 30.0 { // (20+40)/2
				t.Errorf("cpu avg = %v, want 30.0", r.Value)
			}
		}
		if r.ResourceType == "memory" && r.MetricName == "usage_percent" {
			memFound = true
			if r.Value != 60.0 { // (50+70)/2
				t.Errorf("memory avg = %v, want 60.0", r.Value)
			}
		}
	}

	if !cpuFound {
		t.Error("CPU metrics not found in results")
	}
	if !memFound {
		t.Error("Memory metrics not found in results")
	}
}

func TestAggregator_GetResourceMetricKeys(t *testing.T) {
	buf := NewMetricBuffer(time.Minute)
	now := time.Now()

	buf.Add(collector.Metric{ResourceType: "cpu", Name: "usage_percent", Value: 10.0, Timestamp: now})
	buf.Add(collector.Metric{ResourceType: "cpu", Name: "idle_percent", Value: 90.0, Timestamp: now})
	buf.Add(collector.Metric{ResourceType: "memory", Name: "usage_percent", Value: 50.0, Timestamp: now})

	keys := buf.GetResourceMetricKeys()

	if len(keys) != 3 {
		t.Errorf("len(keys) = %d, want 3", len(keys))
	}
}

// Test AggregationResult
func TestAggregationResult_Fields(t *testing.T) {
	result := AggregationResult{
		ResourceType:    "cpu",
		MetricName:      "usage_percent",
		Value:           45.5,
		WindowSize:      time.Minute,
		AggregationType: "avg",
		Timestamp:       time.Now(),
	}

	if result.ResourceType != "cpu" {
		t.Errorf("ResourceType = %s, want cpu", result.ResourceType)
	}
	if result.Value != 45.5 {
		t.Errorf("Value = %v, want 45.5", result.Value)
	}
}

// Test edge cases
func TestAggregator_EmptyBuffer(t *testing.T) {
	windows := []time.Duration{time.Minute}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	// Don't add any metrics, just check windows after time passes
	now := time.Now().Truncate(time.Minute)
	futureTime := now.Add(time.Minute + time.Second)

	results := agg.CheckWindows(futureTime)

	// Should return empty results for empty buffer
	if len(results) != 0 {
		t.Errorf("CheckWindows on empty buffer returned %d results, want 0", len(results))
	}
}

func TestAggregator_GetWindowDurations(t *testing.T) {
	windows := []time.Duration{time.Minute, 5 * time.Minute, time.Hour}
	aggTypes := []string{"avg"}
	agg := NewAggregator(windows, aggTypes)

	durations := agg.GetWindowDurations()

	if len(durations) != 3 {
		t.Errorf("len(durations) = %d, want 3", len(durations))
	}

	// Verify all expected durations are present
	found := make(map[time.Duration]bool)
	for _, d := range durations {
		found[d] = true
	}

	for _, expected := range windows {
		if !found[expected] {
			t.Errorf("Expected duration %v not found in results", expected)
		}
	}
}

func TestMetricBuffer_Duration(t *testing.T) {
	buf := NewMetricBuffer(5 * time.Minute)

	if buf.Duration() != 5*time.Minute {
		t.Errorf("Duration() = %v, want 5m", buf.Duration())
	}
}

func TestAggregator_LabelAwareAggregation(t *testing.T) {
	windows := []time.Duration{100 * time.Millisecond}
	agg := NewAggregator(windows, []string{"avg"})

	// Add metrics with different labels (simulating different disk partitions)
	now := time.Now()

	// Metrics for partition /
	agg.Add(collector.Metric{
		ResourceType: "disk",
		Name:         "usage_percent",
		Value:        50.0,
		Timestamp:    now,
		Labels:       map[string]string{"mountpoint": "/"},
	})
	agg.Add(collector.Metric{
		ResourceType: "disk",
		Name:         "usage_percent",
		Value:        60.0,
		Timestamp:    now,
		Labels:       map[string]string{"mountpoint": "/"},
	})

	// Metrics for partition /home
	agg.Add(collector.Metric{
		ResourceType: "disk",
		Name:         "usage_percent",
		Value:        70.0,
		Timestamp:    now,
		Labels:       map[string]string{"mountpoint": "/home"},
	})
	agg.Add(collector.Metric{
		ResourceType: "disk",
		Name:         "usage_percent",
		Value:        80.0,
		Timestamp:    now,
		Labels:       map[string]string{"mountpoint": "/home"},
	})

	// Wait for window to complete
	time.Sleep(150 * time.Millisecond)

	results := agg.CheckWindows(time.Now())

	// Should have 2 separate aggregations (one for /, one for /home)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results (one per label), got %d", len(results))
	}

	// Check that labels are preserved
	labelsSeen := make(map[string]float64)
	for _, r := range results {
		labelsSeen[r.Labels] = r.Value
	}

	// Verify both label sets exist
	if len(labelsSeen) != 2 {
		t.Errorf("Expected 2 unique label sets, got %d", len(labelsSeen))
	}
}
