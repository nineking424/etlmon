package tui

import (
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
	"github.com/etlmon/etlmon/internal/storage"
)

// Test App Creation
func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
}

func TestNewApp_HasRealtimeView(t *testing.T) {
	app := NewApp()
	if app.realtimeView == nil {
		t.Error("App has no realtimeView")
	}
}

func TestNewApp_HasHistoryView(t *testing.T) {
	app := NewApp()
	if app.historyView == nil {
		t.Error("App has no historyView")
	}
}

func TestNewApp_HasStatusBar(t *testing.T) {
	app := NewApp()
	if app.statusBar == nil {
		t.Error("App has no statusBar")
	}
}

// Test RealtimeView
func TestRealtimeView_Update(t *testing.T) {
	view := NewRealtimeView()

	metrics := []collector.Metric{
		{ResourceType: "cpu", Name: "usage_percent", Value: 45.5, Timestamp: time.Now()},
		{ResourceType: "memory", Name: "usage_percent", Value: 60.0, Timestamp: time.Now()},
	}

	view.Update(metrics)

	// View should have the metrics
	if len(view.currentMetrics) != 2 {
		t.Errorf("currentMetrics len = %d, want 2", len(view.currentMetrics))
	}
}

func TestRealtimeView_UpdateThreadSafe(t *testing.T) {
	view := NewRealtimeView()

	done := make(chan bool)

	// Concurrent updates
	go func() {
		for i := 0; i < 100; i++ {
			view.Update([]collector.Metric{
				{ResourceType: "cpu", Name: "usage_percent", Value: float64(i), Timestamp: time.Now()},
			})
		}
		done <- true
	}()

	// Concurrent reads (simulating render)
	go func() {
		for i := 0; i < 100; i++ {
			_ = view.GetText()
		}
		done <- true
	}()

	<-done
	<-done
}

func TestRealtimeView_GetText(t *testing.T) {
	view := NewRealtimeView()

	metrics := []collector.Metric{
		{ResourceType: "cpu", Name: "usage_percent", Value: 45.5, Timestamp: time.Now()},
	}

	view.Update(metrics)
	text := view.GetText()

	if text == "" {
		t.Error("GetText() returned empty string")
	}
}

// Test HistoryView
func TestHistoryView_Update(t *testing.T) {
	view := NewHistoryView()

	results := []aggregator.AggregationResult{
		{
			ResourceType:    "cpu",
			MetricName:      "usage_percent",
			Value:           45.5,
			WindowSize:      time.Minute,
			AggregationType: "avg",
			Timestamp:       time.Now(),
		},
	}

	view.Update(results)

	if len(view.results) != 1 {
		t.Errorf("results len = %d, want 1", len(view.results))
	}
}

func TestHistoryView_SetWindowFilter(t *testing.T) {
	view := NewHistoryView()

	view.SetWindowFilter("1m")

	if view.windowFilter != "1m" {
		t.Errorf("windowFilter = %s, want 1m", view.windowFilter)
	}
}

func TestHistoryView_SetResourceFilter(t *testing.T) {
	view := NewHistoryView()

	view.SetResourceFilter("cpu")

	if view.resourceFilter != "cpu" {
		t.Errorf("resourceFilter = %s, want cpu", view.resourceFilter)
	}
}

func TestHistoryView_GetText(t *testing.T) {
	view := NewHistoryView()

	results := []aggregator.AggregationResult{
		{
			ResourceType:    "cpu",
			MetricName:      "usage_percent",
			Value:           45.5,
			WindowSize:      time.Minute,
			AggregationType: "avg",
			Timestamp:       time.Now(),
		},
	}

	view.Update(results)
	text := view.GetText()

	if text == "" {
		t.Error("GetText() returned empty string")
	}
}

// Test StatusBar
func TestStatusBar_SetStatus(t *testing.T) {
	bar := NewStatusBar()

	bar.SetStatus("Collecting...")

	if bar.status != "Collecting..." {
		t.Errorf("status = %s, want 'Collecting...'", bar.status)
	}
}

func TestStatusBar_SetLastUpdate(t *testing.T) {
	bar := NewStatusBar()
	now := time.Now()

	bar.SetLastUpdate(now)

	if !bar.lastUpdate.Equal(now) {
		t.Errorf("lastUpdate = %v, want %v", bar.lastUpdate, now)
	}
}

func TestStatusBar_GetText(t *testing.T) {
	bar := NewStatusBar()
	bar.SetStatus("Running")
	bar.SetLastUpdate(time.Now())

	text := bar.GetText()

	if text == "" {
		t.Error("GetText() returned empty string")
	}
}

// Test App Methods
func TestApp_SetStore(t *testing.T) {
	app := NewApp()
	// nil store should be acceptable (used before initialization)
	app.SetStore(nil)
}

func TestApp_OnMetricsCollected(t *testing.T) {
	app := NewApp()

	metrics := []collector.Metric{
		{ResourceType: "cpu", Name: "usage_percent", Value: 45.5, Timestamp: time.Now()},
	}

	// Should not panic
	app.OnMetricsCollected(metrics)

	// Check realtime view updated
	if len(app.realtimeView.currentMetrics) != 1 {
		t.Error("Metrics not passed to realtime view")
	}
}

func TestApp_OnAggregationComplete(t *testing.T) {
	app := NewApp()

	results := []aggregator.AggregationResult{
		{
			ResourceType:    "cpu",
			MetricName:      "usage_percent",
			Value:           45.5,
			WindowSize:      time.Minute,
			AggregationType: "avg",
			Timestamp:       time.Now(),
		},
	}

	// Should not panic
	app.OnAggregationComplete(results)
}

// Test View Mode switching
func TestApp_SwitchView(t *testing.T) {
	app := NewApp()

	// Default should be realtime
	if app.currentView != ViewRealtime {
		t.Errorf("Default view = %v, want ViewRealtime", app.currentView)
	}

	app.SwitchView(ViewHistory)
	if app.currentView != ViewHistory {
		t.Errorf("After switch, view = %v, want ViewHistory", app.currentView)
	}

	app.SwitchView(ViewRealtime)
	if app.currentView != ViewRealtime {
		t.Errorf("After switch back, view = %v, want ViewRealtime", app.currentView)
	}
}

// Test formatting helpers
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("FormatBytes(%v) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{time.Minute, "1m"},
		{5 * time.Minute, "5m"},
		{time.Hour, "1h"},
		{30 * time.Second, "30s"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

// Test Store interface
func TestStore_Interface(t *testing.T) {
	// Verify storage.SQLiteStore implements our Store interface
	var _ Store = (*storage.SQLiteStore)(nil)
}
