package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

func TestPathsProvider_Tabs(t *testing.T) {
	provider := NewPathsDetailProvider(nil, nil)
	tabs := provider.Tabs()

	expected := []string{"Stats", "Scan"}
	if len(tabs) != len(expected) {
		t.Fatalf("expected %d tabs, got %d", len(expected), len(tabs))
	}

	for i, tab := range tabs {
		if tab != expected[i] {
			t.Errorf("tab %d: expected %q, got %q", i, expected[i], tab)
		}
	}
}

func TestPathsProvider_Refresh_Success(t *testing.T) {
	mock := &mockAPIClient{
		pathStats: []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      12345,
				DirCount:       678,
				ScanDurationMs: 2500,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
			{
				Path:           "/var/log",
				FileCount:      9876,
				DirCount:       432,
				ScanDurationMs: 1800,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
		},
	}

	provider := NewPathsDetailProvider(mock, nil)
	err := provider.Refresh(context.Background(), mock)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Verify data was stored
	if len(provider.data) != 2 {
		t.Errorf("expected 2 path stats entries, got %d", len(provider.data))
	}
}

func TestPathsProvider_Refresh_Error(t *testing.T) {
	mock := &mockAPIClient{
		pathErr: context.Canceled,
	}

	provider := NewPathsDetailProvider(mock, nil)
	err := provider.Refresh(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected Canceled, got %v", err)
	}
}

func TestPathsProvider_StatsTab(t *testing.T) {
	mock := &mockAPIClient{
		pathStats: []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      12345,
				DirCount:       678,
				ScanDurationMs: 2500,
				Status:         "OK",
			},
		},
	}

	provider := NewPathsDetailProvider(mock, nil)
	_ = provider.Refresh(context.Background(), mock)

	// Get stats tab content
	primitive := provider.TabContent(0)
	if primitive == nil {
		t.Fatal("Stats tab returned nil primitive")
	}

	// Verify it's a Table
	if provider.statsTable == nil {
		t.Fatal("statsTable is nil")
	}

	// Check table has header + 1 data row
	rowCount := provider.statsTable.GetRowCount()
	if rowCount != 2 { // header + 1 data row
		t.Errorf("expected 2 rows (header + 1 data), got %d", rowCount)
	}

	// Check header cells
	headers := []string{"Path", "Files", "Dirs", "Duration", "Status"}
	for col, expectedHeader := range headers {
		cell := provider.statsTable.GetCell(0, col)
		if cell == nil {
			t.Errorf("header cell [0,%d] is nil", col)
			continue
		}
		if cell.Text != expectedHeader {
			t.Errorf("header[%d]: expected %q, got %q", col, expectedHeader, cell.Text)
		}
	}

	// Check first data row contains path
	pathCell := provider.statsTable.GetCell(1, 0)
	if pathCell == nil {
		t.Fatal("path cell is nil")
	}
	if pathCell.Text != "/data/logs" {
		t.Errorf("expected path '/data/logs', got %q", pathCell.Text)
	}

	// Check file count is formatted with commas
	filesCell := provider.statsTable.GetCell(1, 1)
	if filesCell == nil {
		t.Fatal("files cell is nil")
	}
	if !strings.Contains(filesCell.Text, ",") {
		t.Errorf("file count should be formatted with commas, got %q", filesCell.Text)
	}
}

func TestPathsProvider_ScanTab(t *testing.T) {
	mock := &mockAPIClient{
		pathStats: []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      100,
				DirCount:       10,
				ScanDurationMs: 500,
				Status:         "OK",
			},
		},
	}

	provider := NewPathsDetailProvider(mock, nil)
	_ = provider.Refresh(context.Background(), mock)

	// Get scan tab content
	primitive := provider.TabContent(1)
	if primitive == nil {
		t.Fatal("Scan tab returned nil primitive")
	}

	// Verify scanFlex exists
	if provider.scanFlex == nil {
		t.Fatal("scanFlex is nil")
	}
}
