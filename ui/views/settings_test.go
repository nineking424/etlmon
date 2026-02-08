package views

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/rivo/tview"
)

// mockSettingsClient implements settingsAPIClient for testing
type mockSettingsClient struct {
	cfg      *config.NodeConfig
	savedCfg *config.NodeConfig
	getErr   error
	saveErr  error
}

func (m *mockSettingsClient) GetConfig(ctx context.Context) (*config.NodeConfig, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cfg, nil
}

func (m *mockSettingsClient) SaveConfig(ctx context.Context, cfg *config.NodeConfig) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCfg = cfg
	return nil
}

// testNodeConfig creates a test configuration
func testNodeConfig() *config.NodeConfig {
	return &config.NodeConfig{
		Node: config.NodeSettings{
			Listen:   "127.0.0.1:8080",
			NodeName: "test",
			DBPath:   "/tmp/test.db",
		},
		Process: config.ProcessConfig{
			Patterns: []string{"java*", "*nifi*"},
			TopN:     50,
		},
		Logs: []config.LogMonitorConfig{
			{Name: "system", Path: "/var/log/system.log", MaxLines: 1000},
		},
		Paths: []config.PathConfig{
			{Path: "/tmp", ScanInterval: 60 * time.Second, MaxDepth: 3, Timeout: 30 * time.Second},
		},
	}
}

func TestNewSettingsView(t *testing.T) {
	v := NewSettingsView()
	if v.Name() != "settings" {
		t.Errorf("expected name 'settings', got %q", v.Name())
	}
	if v.Primitive() == nil {
		t.Error("Primitive() should not be nil")
	}
}

func TestSettingsView_RefreshLoadsConfig(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{cfg: testNodeConfig()}
	ctx := context.Background()

	err := v.refresh(ctx, mock)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if v.cfg == nil {
		t.Fatal("cfg should not be nil after refresh")
	}
	if len(v.cfg.Process.Patterns) != 2 {
		t.Errorf("expected 2 process patterns, got %d", len(v.cfg.Process.Patterns))
	}
	if v.cfg.Process.Patterns[0] != "java*" {
		t.Errorf("expected first pattern 'java*', got %q", v.cfg.Process.Patterns[0])
	}
}

func TestSettingsView_DirtyPreventsRefresh(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{cfg: testNodeConfig()}
	ctx := context.Background()

	// First refresh
	v.refresh(ctx, mock)

	// Modify and set dirty
	v.cfg.Process.Patterns = append(v.cfg.Process.Patterns, "new-pattern")
	v.dirty = true

	// Create new mock with different data
	mock2 := &mockSettingsClient{cfg: &config.NodeConfig{
		Process: config.ProcessConfig{Patterns: []string{"replaced"}},
	}}

	// Refresh should be skipped because dirty
	err := v.refresh(ctx, mock2)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	// Should still have our modified data
	if len(v.cfg.Process.Patterns) != 3 {
		t.Errorf("expected 3 patterns (dirty=true should skip refresh), got %d", len(v.cfg.Process.Patterns))
	}
}

func TestSettingsView_SaveSendsConfig(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{cfg: testNodeConfig()}
	ctx := context.Background()

	// Load config
	v.refresh(ctx, mock)
	v.apiClient = mock

	// Modify
	v.cfg.Process.Patterns = append(v.cfg.Process.Patterns, "test-pattern")
	v.dirty = true

	// Save
	v.save()

	if mock.savedCfg == nil {
		t.Fatal("save should have sent config to client")
	}
	if len(mock.savedCfg.Process.Patterns) != 3 {
		t.Errorf("expected 3 patterns in saved config, got %d", len(mock.savedCfg.Process.Patterns))
	}
	if mock.savedCfg.Process.Patterns[2] != "test-pattern" {
		t.Errorf("expected 'test-pattern', got %q", mock.savedCfg.Process.Patterns[2])
	}
	if v.dirty {
		t.Error("dirty should be false after save")
	}
}

func TestSettingsView_SaveWithNoClient(t *testing.T) {
	v := NewSettingsView()
	// No apiClient set
	var gotError bool
	v.onStatusChange = func(msg string, isError bool) {
		if isError {
			gotError = true
		}
	}
	v.save()
	if !gotError {
		t.Error("save with no client should report error")
	}
}

func TestSettingsView_SaveWithNoConfig(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{}
	v.apiClient = mock
	// No config loaded
	var gotError bool
	v.onStatusChange = func(msg string, isError bool) {
		if isError {
			gotError = true
		}
	}
	v.save()
	if !gotError {
		t.Error("save with no config should report error")
	}
}

func TestSettingsView_RefreshError(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{getErr: fmt.Errorf("connection refused")}
	ctx := context.Background()

	err := v.refresh(ctx, mock)
	if err == nil {
		t.Error("expected error from refresh")
	}
}

func TestSettingsView_DeleteProcessPattern(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	// Select row 1 (first pattern - "java*")
	v.processTable.Select(1, 0)

	v.deleteProcessPattern()

	if len(v.cfg.Process.Patterns) != 1 {
		t.Errorf("expected 1 pattern after delete, got %d", len(v.cfg.Process.Patterns))
	}
	if v.cfg.Process.Patterns[0] != "*nifi*" {
		t.Errorf("expected '*nifi*', got %q", v.cfg.Process.Patterns[0])
	}
	if !v.dirty {
		t.Error("dirty should be true after delete")
	}
}

func TestSettingsView_DeleteLogEntry(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	// Select row 1 (first log)
	v.logTable.Select(1, 0)

	v.deleteLogEntry()

	if len(v.cfg.Logs) != 0 {
		t.Errorf("expected 0 logs after delete, got %d", len(v.cfg.Logs))
	}
	if !v.dirty {
		t.Error("dirty should be true after delete")
	}
}

func TestSettingsView_DeletePathEntry(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	// Select row 1 (first path)
	v.pathTable.Select(1, 0)

	v.deletePathEntry()

	if len(v.cfg.Paths) != 0 {
		t.Errorf("expected 0 paths after delete, got %d", len(v.cfg.Paths))
	}
	if !v.dirty {
		t.Error("dirty should be true after delete")
	}
}

func TestSettingsView_IsEditing(t *testing.T) {
	v := NewSettingsView()

	if v.IsEditing() {
		t.Error("should not be editing initially")
	}

	// Simulate modal
	v.pages.AddPage("modal", tview.NewBox(), true, true)
	if !v.IsEditing() {
		t.Error("should be editing when modal page exists")
	}

	v.pages.RemovePage("modal")
	if v.IsEditing() {
		t.Error("should not be editing after modal removed")
	}
}

func TestSettingsView_RefreshProcessTable(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	v.refreshProcessTable()

	// Should have header + 2 patterns + 1 blank + 1 TopN = 5 rows
	rowCount := v.processTable.GetRowCount()
	if rowCount != 5 {
		t.Errorf("expected 5 rows in process table, got %d", rowCount)
	}
}

func TestSettingsView_RefreshLogTable(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	v.refreshLogTable()

	// Should have header + 1 log = 2 rows
	rowCount := v.logTable.GetRowCount()
	if rowCount != 2 {
		t.Errorf("expected 2 rows in log table, got %d", rowCount)
	}
}

func TestSettingsView_RefreshPathTable(t *testing.T) {
	v := NewSettingsView()
	v.cfg = testNodeConfig()

	v.refreshPathTable()

	// Should have header + 1 path = 2 rows
	rowCount := v.pathTable.GetRowCount()
	if rowCount != 2 {
		t.Errorf("expected 2 rows in path table, got %d", rowCount)
	}
}

func TestSettingsView_RefreshWithNilConfig(t *testing.T) {
	v := NewSettingsView()
	v.cfg = nil

	// Should not panic
	v.refreshProcessTable()
	v.refreshLogTable()
	v.refreshPathTable()
}

func TestSettingsView_SaveError(t *testing.T) {
	v := NewSettingsView()
	mock := &mockSettingsClient{
		cfg:     testNodeConfig(),
		saveErr: fmt.Errorf("save failed"),
	}
	ctx := context.Background()

	// Load config
	v.refresh(ctx, mock)
	v.apiClient = mock

	var gotError bool
	var errorMsg string
	v.onStatusChange = func(msg string, isError bool) {
		if isError {
			gotError = true
			errorMsg = msg
		}
	}

	// Save should fail
	v.save()

	if !gotError {
		t.Error("expected error status")
	}
	if errorMsg == "" {
		t.Error("expected error message")
	}
}
