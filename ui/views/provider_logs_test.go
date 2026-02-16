package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/rivo/tview"
)

func TestLogsProvider_Tabs(t *testing.T) {
	app := tview.NewApplication()
	provider := NewLogsDetailProvider(nil, app)
	tabs := provider.Tabs()

	expected := []string{"Files", "Viewer"}
	if len(tabs) != len(expected) {
		t.Fatalf("expected %d tabs, got %d", len(expected), len(tabs))
	}

	for i, tab := range tabs {
		if tab != expected[i] {
			t.Errorf("tab %d: expected %q, got %q", i, expected[i], tab)
		}
	}
}

func TestLogsProvider_Refresh_Success(t *testing.T) {
	mock := &mockAPIClient{
		logFiles: []models.LogFileInfo{
			{
				Name:     "app.log",
				Path:     "/var/log/app.log",
				MaxLines: 1000,
				Size:     1024 * 512, // 512 KB
				ModTime:  time.Now(),
			},
			{
				Name:     "error.log",
				Path:     "/var/log/error.log",
				MaxLines: 500,
				Size:     1024 * 1024 * 2, // 2 MB
				ModTime:  time.Now(),
			},
		},
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)
	err := provider.Refresh(context.Background(), mock)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Verify data was stored
	if len(provider.logFiles) != 2 {
		t.Errorf("expected 2 log files, got %d", len(provider.logFiles))
	}
}

func TestLogsProvider_Refresh_Error(t *testing.T) {
	mock := &mockAPIClient{
		logErr: context.DeadlineExceeded,
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)
	err := provider.Refresh(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestLogsProvider_FilesTab(t *testing.T) {
	mock := &mockAPIClient{
		logFiles: []models.LogFileInfo{
			{
				Name:     "app.log",
				Path:     "/var/log/app.log",
				MaxLines: 1000,
				Size:     1024 * 512,
				ModTime:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			{
				Name:     "error.log",
				Path:     "/var/log/error.log",
				MaxLines: 500,
				Size:     1024 * 1024 * 2,
				ModTime:  time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC),
			},
		},
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)
	_ = provider.Refresh(context.Background(), mock)

	// Get Files tab content
	primitive := provider.TabContent(0)
	if primitive == nil {
		t.Fatal("Files tab returned nil primitive")
	}

	// Verify it's the filesTable
	if provider.filesTable == nil {
		t.Fatal("filesTable is nil")
	}

	// Check table has header + 2 data rows
	rowCount := provider.filesTable.GetRowCount()
	if rowCount != 3 { // header + 2 data rows
		t.Errorf("expected 3 rows (header + 2 data), got %d", rowCount)
	}

	// Check header cells
	headers := []string{"Name", "Path", "Size", "Modified"}
	for col, expectedHeader := range headers {
		cell := provider.filesTable.GetCell(0, col)
		if cell == nil {
			t.Errorf("header cell [0,%d] is nil", col)
			continue
		}
		if cell.Text != expectedHeader {
			t.Errorf("header[%d]: expected %q, got %q", col, expectedHeader, cell.Text)
		}
	}

	// Check first data row
	nameCell := provider.filesTable.GetCell(1, 0)
	if nameCell == nil {
		t.Fatal("Name cell is nil")
	}
	if nameCell.Text != "app.log" {
		t.Errorf("expected Name 'app.log', got %q", nameCell.Text)
	}

	pathCell := provider.filesTable.GetCell(1, 1)
	if pathCell == nil {
		t.Fatal("Path cell is nil")
	}
	if pathCell.Text != "/var/log/app.log" {
		t.Errorf("expected Path '/var/log/app.log', got %q", pathCell.Text)
	}

	sizeCell := provider.filesTable.GetCell(1, 2)
	if sizeCell == nil {
		t.Fatal("Size cell is nil")
	}
	// Should contain "512" and "KB"
	if !strings.Contains(sizeCell.Text, "512") || !strings.Contains(sizeCell.Text, "KB") {
		t.Errorf("expected Size to contain '512 KB', got %q", sizeCell.Text)
	}

	modCell := provider.filesTable.GetCell(1, 3)
	if modCell == nil {
		t.Fatal("Modified cell is nil")
	}
	if !strings.Contains(modCell.Text, "2024-01-15") {
		t.Errorf("expected Modified to contain date, got %q", modCell.Text)
	}
}

func TestLogsProvider_FilesTab_EmptySize(t *testing.T) {
	mock := &mockAPIClient{
		logFiles: []models.LogFileInfo{
			{
				Name:     "empty.log",
				Path:     "/var/log/empty.log",
				MaxLines: 100,
				Size:     0, // Should display as "-"
				ModTime:  time.Now(),
			},
		},
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)
	_ = provider.Refresh(context.Background(), mock)

	// Check size cell for 0 bytes
	sizeCell := provider.filesTable.GetCell(1, 2)
	if sizeCell == nil {
		t.Fatal("Size cell is nil")
	}
	if sizeCell.Text != "-" {
		t.Errorf("expected Size '-' for 0 bytes, got %q", sizeCell.Text)
	}
}

func TestLogsProvider_ViewerTab_Empty(t *testing.T) {
	app := tview.NewApplication()
	provider := NewLogsDetailProvider(nil, app)

	// Get Viewer tab content without selecting a file
	primitive := provider.TabContent(1)
	if primitive == nil {
		t.Fatal("Viewer tab returned nil primitive")
	}

	// Verify it's the viewer TextView
	if provider.viewer == nil {
		t.Fatal("viewer is nil")
	}

	// Check for placeholder text
	text := provider.viewer.GetText(false)
	if !strings.Contains(text, "Select a log file") {
		t.Errorf("expected placeholder text, got %q", text)
	}
}

func TestLogsProvider_ViewerTab_Content(t *testing.T) {
	mock := &mockAPIClient{
		logFiles: []models.LogFileInfo{
			{Name: "app.log", Path: "/var/log/app.log", Size: 1024, ModTime: time.Now()},
		},
		logEntries: []*models.LogEntry{
			{
				ID:        1,
				LogName:   "app.log",
				LogPath:   "/var/log/app.log",
				Line:      "INFO: Application started",
				CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:        2,
				LogName:   "app.log",
				LogPath:   "/var/log/app.log",
				Line:      "ERROR: Connection failed",
				CreatedAt: time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC),
			},
		},
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)

	// Load log content
	err := provider.loadLogContent("app.log")
	if err != nil {
		t.Fatalf("loadLogContent failed: %v", err)
	}

	// Check viewer has content
	text := provider.viewer.GetText(false)
	if !strings.Contains(text, "INFO: Application started") {
		t.Errorf("expected viewer to contain first log line, got %q", text)
	}
	if !strings.Contains(text, "ERROR: Connection failed") {
		t.Errorf("expected viewer to contain second log line, got %q", text)
	}
	if !strings.Contains(text, "10:00:00") {
		t.Errorf("expected viewer to contain timestamp, got %q", text)
	}
}

func TestLogsProvider_ViewerTab_NoEntries(t *testing.T) {
	mock := &mockAPIClient{
		logEntries: []*models.LogEntry{}, // Empty log
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)

	err := provider.loadLogContent("empty.log")
	if err != nil {
		t.Fatalf("loadLogContent failed: %v", err)
	}

	// Check for "No log entries" message
	text := provider.viewer.GetText(false)
	if !strings.Contains(text, "No log entries") {
		t.Errorf("expected 'No log entries' message, got %q", text)
	}
}

func TestLogsProvider_SelectFile(t *testing.T) {
	mock := &mockAPIClient{
		logFiles: []models.LogFileInfo{
			{Name: "app.log", Path: "/var/log/app.log", Size: 1024, ModTime: time.Now()},
		},
		logEntries: []*models.LogEntry{
			{
				ID:        1,
				LogName:   "app.log",
				Line:      "Test log line",
				CreatedAt: time.Now(),
			},
		},
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)
	_ = provider.Refresh(context.Background(), mock)

	// Initially no file selected
	if provider.selectedLog != "" {
		t.Errorf("expected no selected log initially, got %q", provider.selectedLog)
	}

	// Select file (simulates Enter key on Files tab)
	err := provider.loadLogContent("app.log")
	if err != nil {
		t.Fatalf("loadLogContent failed: %v", err)
	}

	// Verify content was loaded
	text := provider.viewer.GetText(false)
	if !strings.Contains(text, "Test log line") {
		t.Errorf("expected viewer to show log content after selection, got %q", text)
	}
}

func TestLogsProvider_LoadLogContent_Error(t *testing.T) {
	mock := &mockAPIClient{
		logEntriesErr: context.DeadlineExceeded,
	}

	app := tview.NewApplication()
	provider := NewLogsDetailProvider(mock, app)

	err := provider.loadLogContent("error.log")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestLogsProvider_OnSelect(t *testing.T) {
	app := tview.NewApplication()
	provider := NewLogsDetailProvider(nil, app)

	// OnSelect should not panic
	provider.OnSelect(0)
	provider.OnSelect(1)
}

func TestLogsProvider_FileSizeFormat(t *testing.T) {
	tests := []struct {
		size     int64
		contains string
	}{
		{0, "-"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 512, "512.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		mock := &mockAPIClient{
			logFiles: []models.LogFileInfo{
				{Name: "test.log", Path: "/test.log", Size: tt.size, ModTime: time.Now()},
			},
		}

		app := tview.NewApplication()
		provider := NewLogsDetailProvider(mock, app)
		_ = provider.Refresh(context.Background(), mock)

		sizeCell := provider.filesTable.GetCell(1, 2)
		if sizeCell == nil {
			t.Fatalf("Size cell is nil for size %d", tt.size)
		}
		if !strings.Contains(sizeCell.Text, tt.contains) {
			t.Errorf("size %d: expected to contain %q, got %q", tt.size, tt.contains, sizeCell.Text)
		}
	}
}
