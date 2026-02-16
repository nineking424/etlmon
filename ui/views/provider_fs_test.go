package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/pkg/models"
)

// mockAPIClient implements ui.APIClient for testing
type mockAPIClient struct {
	fsUsage       []*models.FilesystemUsage
	pathStats     []*models.PathStats
	procInfo      []*models.ProcessInfo
	logFiles      []models.LogFileInfo
	logEntries    []*models.LogEntry
	cfg           *config.NodeConfig
	fsErr         error
	pathErr       error
	procErr       error
	logErr        error
	logEntriesErr error
	scanErr       error
	cfgErr        error
	saveErr       error
}

func (m *mockAPIClient) GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return m.fsUsage, m.fsErr
}

func (m *mockAPIClient) GetPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return m.pathStats, m.pathErr
}

func (m *mockAPIClient) TriggerScan(ctx context.Context, paths []string) error {
	return m.scanErr
}

func (m *mockAPIClient) GetProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error) {
	return m.procInfo, m.procErr
}

func (m *mockAPIClient) GetLogFiles(ctx context.Context) ([]models.LogFileInfo, error) {
	return m.logFiles, m.logErr
}

func (m *mockAPIClient) GetLogEntriesByName(ctx context.Context, name string, limit int) ([]*models.LogEntry, error) {
	return m.logEntries, m.logEntriesErr
}

func (m *mockAPIClient) GetConfig(ctx context.Context) (*config.NodeConfig, error) {
	return m.cfg, m.cfgErr
}

func (m *mockAPIClient) SaveConfig(ctx context.Context, cfg *config.NodeConfig) error {
	return m.saveErr
}

func TestFSProvider_Tabs(t *testing.T) {
	provider := NewFSDetailProvider()
	tabs := provider.Tabs()

	expected := []string{"Summary", "Usage"}
	if len(tabs) != len(expected) {
		t.Fatalf("expected %d tabs, got %d", len(expected), len(tabs))
	}

	for i, tab := range tabs {
		if tab != expected[i] {
			t.Errorf("tab %d: expected %q, got %q", i, expected[i], tab)
		}
	}
}

func TestFSProvider_Refresh_Success(t *testing.T) {
	mock := &mockAPIClient{
		fsUsage: []*models.FilesystemUsage{
			{
				MountPoint:  "/",
				TotalBytes:  100 * 1024 * 1024 * 1024, // 100 GB
				UsedBytes:   60 * 1024 * 1024 * 1024,  // 60 GB
				AvailBytes:  40 * 1024 * 1024 * 1024,  // 40 GB
				UsedPercent: 60.0,
				CollectedAt: time.Now(),
			},
			{
				MountPoint:  "/data",
				TotalBytes:  500 * 1024 * 1024 * 1024, // 500 GB
				UsedBytes:   450 * 1024 * 1024 * 1024, // 450 GB
				AvailBytes:  50 * 1024 * 1024 * 1024,  // 50 GB
				UsedPercent: 90.0,
				CollectedAt: time.Now(),
			},
		},
	}

	provider := NewFSDetailProvider()
	err := provider.Refresh(context.Background(), mock)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Verify data was stored
	if len(provider.data) != 2 {
		t.Errorf("expected 2 filesystem entries, got %d", len(provider.data))
	}
}

func TestFSProvider_Refresh_Error(t *testing.T) {
	mock := &mockAPIClient{
		fsErr: context.DeadlineExceeded,
	}

	provider := NewFSDetailProvider()
	err := provider.Refresh(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestFSProvider_SummaryTab(t *testing.T) {
	mock := &mockAPIClient{
		fsUsage: []*models.FilesystemUsage{
			{
				MountPoint:  "/",
				TotalBytes:  100 * 1024 * 1024 * 1024,
				UsedBytes:   60 * 1024 * 1024 * 1024,
				AvailBytes:  40 * 1024 * 1024 * 1024,
				UsedPercent: 60.0,
			},
			{
				MountPoint:  "/data",
				TotalBytes:  200 * 1024 * 1024 * 1024,
				UsedBytes:   180 * 1024 * 1024 * 1024,
				AvailBytes:  20 * 1024 * 1024 * 1024,
				UsedPercent: 90.0,
			},
		},
	}

	provider := NewFSDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Get summary tab content
	primitive := provider.TabContent(0)
	if primitive == nil {
		t.Fatal("Summary tab returned nil primitive")
	}

	// Verify it's a TextView
	if provider.summaryBox == nil {
		t.Fatal("summaryBox is nil")
	}

	text := provider.summaryBox.GetText(false)

	// Check for key summary information
	if !strings.Contains(text, "2") { // mount count
		t.Error("Summary should contain mount count")
	}
	if !strings.Contains(text, "Total") {
		t.Error("Summary should contain 'Total' label")
	}
	if !strings.Contains(text, "Used") {
		t.Error("Summary should contain 'Used' label")
	}
	if !strings.Contains(text, "Available") {
		t.Error("Summary should contain 'Available' label")
	}
}

func TestFSProvider_UsageTab(t *testing.T) {
	mock := &mockAPIClient{
		fsUsage: []*models.FilesystemUsage{
			{
				MountPoint:  "/",
				TotalBytes:  100 * 1024 * 1024 * 1024,
				UsedBytes:   60 * 1024 * 1024 * 1024,
				AvailBytes:  40 * 1024 * 1024 * 1024,
				UsedPercent: 60.0,
			},
		},
	}

	provider := NewFSDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Get usage tab content
	primitive := provider.TabContent(1)
	if primitive == nil {
		t.Fatal("Usage tab returned nil primitive")
	}

	// Verify it's a Table
	if provider.usageTable == nil {
		t.Fatal("usageTable is nil")
	}

	// Check table has header + 1 data row
	rowCount := provider.usageTable.GetRowCount()
	if rowCount != 2 { // header + 1 data row
		t.Errorf("expected 2 rows (header + 1 data), got %d", rowCount)
	}

	// Check header cells
	headers := []string{"Mount", "Total", "Used", "Avail", "Use%", "Usage"}
	for col, expectedHeader := range headers {
		cell := provider.usageTable.GetCell(0, col)
		if cell == nil {
			t.Errorf("header cell [0,%d] is nil", col)
			continue
		}
		if cell.Text != expectedHeader {
			t.Errorf("header[%d]: expected %q, got %q", col, expectedHeader, cell.Text)
		}
	}

	// Check first data row contains mount point
	mountCell := provider.usageTable.GetCell(1, 0)
	if mountCell == nil {
		t.Fatal("mount cell is nil")
	}
	if mountCell.Text != "/" {
		t.Errorf("expected mount '/', got %q", mountCell.Text)
	}

	// Check gauge column exists
	gaugeCell := provider.usageTable.GetCell(1, 5)
	if gaugeCell == nil {
		t.Fatal("gauge cell is nil")
	}
	if !strings.Contains(gaugeCell.Text, "█") && !strings.Contains(gaugeCell.Text, "░") {
		t.Errorf("gauge cell should contain gauge characters, got %q", gaugeCell.Text)
	}
}
