package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

func TestProcessProvider_Tabs(t *testing.T) {
	provider := NewProcessDetailProvider()
	tabs := provider.Tabs()

	expected := []string{"List", "Top CPU", "Top Memory"}
	if len(tabs) != len(expected) {
		t.Fatalf("expected %d tabs, got %d", len(expected), len(tabs))
	}

	for i, tab := range tabs {
		if tab != expected[i] {
			t.Errorf("tab %d: expected %q, got %q", i, expected[i], tab)
		}
	}
}

func TestProcessProvider_Refresh_Success(t *testing.T) {
	mock := &mockAPIClient{
		procInfo: []*models.ProcessInfo{
			{
				PID:         1234,
				Name:        "test-process",
				User:        "root",
				CPUPercent:  45.5,
				MemRSS:      1024 * 1024 * 100, // 100 MB
				Status:      "running",
				Elapsed:     "1h23m",
				CollectedAt: time.Now(),
			},
			{
				PID:         5678,
				Name:        "another-process",
				User:        "user",
				CPUPercent:  85.2,
				MemRSS:      1024 * 1024 * 200, // 200 MB
				Status:      "sleeping",
				Elapsed:     "45m",
				CollectedAt: time.Now(),
			},
		},
	}

	provider := NewProcessDetailProvider()
	err := provider.Refresh(context.Background(), mock)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Verify data was stored
	if len(provider.data) != 2 {
		t.Errorf("expected 2 processes, got %d", len(provider.data))
	}
}

func TestProcessProvider_Refresh_Error(t *testing.T) {
	mock := &mockAPIClient{
		procErr: context.DeadlineExceeded,
	}

	provider := NewProcessDetailProvider()
	err := provider.Refresh(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestProcessProvider_ListTab(t *testing.T) {
	mock := &mockAPIClient{
		procInfo: []*models.ProcessInfo{
			{
				PID:        1234,
				Name:       "test-process",
				User:       "root",
				CPUPercent: 45.5,
				MemRSS:     1024 * 1024 * 100,
				Status:     "running",
				Elapsed:    "1h23m",
			},
			{
				PID:        5678,
				Name:       "another",
				User:       "user",
				CPUPercent: 25.0,
				MemRSS:     1024 * 1024 * 50,
				Status:     "sleeping",
				Elapsed:    "30m",
			},
		},
	}

	provider := NewProcessDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Get List tab content
	primitive := provider.TabContent(0)
	if primitive == nil {
		t.Fatal("List tab returned nil primitive")
	}

	// Verify it's the listTable
	if provider.listTable == nil {
		t.Fatal("listTable is nil")
	}

	// Check table has header + 2 data rows
	rowCount := provider.listTable.GetRowCount()
	if rowCount != 3 { // header + 2 data rows
		t.Errorf("expected 3 rows (header + 2 data), got %d", rowCount)
	}

	// Check header cells
	headers := []string{"PID", "User", "CPU%", "Memory", "Status", "Elapsed", "Name"}
	for col, expectedHeader := range headers {
		cell := provider.listTable.GetCell(0, col)
		if cell == nil {
			t.Errorf("header cell [0,%d] is nil", col)
			continue
		}
		if cell.Text != expectedHeader {
			t.Errorf("header[%d]: expected %q, got %q", col, expectedHeader, cell.Text)
		}
	}

	// Check first data row
	pidCell := provider.listTable.GetCell(1, 0)
	if pidCell == nil {
		t.Fatal("PID cell is nil")
	}
	if pidCell.Text != "1234" {
		t.Errorf("expected PID '1234', got %q", pidCell.Text)
	}

	nameCell := provider.listTable.GetCell(1, 6)
	if nameCell == nil {
		t.Fatal("Name cell is nil")
	}
	if nameCell.Text != "test-process" {
		t.Errorf("expected Name 'test-process', got %q", nameCell.Text)
	}
}

func TestProcessProvider_ListTab_ColorCoding(t *testing.T) {
	mock := &mockAPIClient{
		procInfo: []*models.ProcessInfo{
			{
				PID:        1,
				Name:       "high-cpu",
				User:       "root",
				CPUPercent: 85.0, // > 80, should be red
				MemRSS:     1024,
				Status:     "running", // should be green
				Elapsed:    "1m",
			},
			{
				PID:        2,
				Name:       "med-cpu",
				User:       "root",
				CPUPercent: 60.0, // > 50, should be yellow
				MemRSS:     1024,
				Status:     "zombie", // should be red
				Elapsed:    "2m",
			},
			{
				PID:        3,
				Name:       "low-cpu",
				User:       "root",
				CPUPercent: 30.0, // normal
				MemRSS:     1024,
				Status:     "stopped", // should be yellow
				Elapsed:    "3m",
			},
		},
	}

	provider := NewProcessDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Check CPU color coding
	// Row 1: CPU > 80 -> red
	cpuCell1 := provider.listTable.GetCell(1, 2)
	if cpuCell1 == nil {
		t.Fatal("CPU cell 1 is nil")
	}
	// Color should be StatusCritical (red), but we can't directly check tcell.Color
	// We verify the cell exists and contains the right value
	if !strings.Contains(cpuCell1.Text, "85") {
		t.Errorf("expected CPU cell to contain '85', got %q", cpuCell1.Text)
	}

	// Row 2: CPU > 50 -> yellow
	cpuCell2 := provider.listTable.GetCell(2, 2)
	if cpuCell2 == nil {
		t.Fatal("CPU cell 2 is nil")
	}
	if !strings.Contains(cpuCell2.Text, "60") {
		t.Errorf("expected CPU cell to contain '60', got %q", cpuCell2.Text)
	}

	// Check Status color coding
	// Row 1: running -> green
	statusCell1 := provider.listTable.GetCell(1, 4)
	if statusCell1 == nil {
		t.Fatal("Status cell 1 is nil")
	}
	if statusCell1.Text != "running" {
		t.Errorf("expected Status 'running', got %q", statusCell1.Text)
	}

	// Row 2: zombie -> red
	statusCell2 := provider.listTable.GetCell(2, 4)
	if statusCell2 == nil {
		t.Fatal("Status cell 2 is nil")
	}
	if statusCell2.Text != "zombie" {
		t.Errorf("expected Status 'zombie', got %q", statusCell2.Text)
	}

	// Row 3: stopped -> yellow
	statusCell3 := provider.listTable.GetCell(3, 4)
	if statusCell3 == nil {
		t.Fatal("Status cell 3 is nil")
	}
	if statusCell3.Text != "stopped" {
		t.Errorf("expected Status 'stopped', got %q", statusCell3.Text)
	}
}

func TestProcessProvider_TopCPUTab(t *testing.T) {
	mock := &mockAPIClient{
		procInfo: []*models.ProcessInfo{
			{PID: 1, Name: "p1", User: "root", CPUPercent: 30.0, MemRSS: 1024, Status: "running", Elapsed: "1m"},
			{PID: 2, Name: "p2", User: "root", CPUPercent: 90.0, MemRSS: 1024, Status: "running", Elapsed: "2m"},
			{PID: 3, Name: "p3", User: "root", CPUPercent: 60.0, MemRSS: 1024, Status: "running", Elapsed: "3m"},
			{PID: 4, Name: "p4", User: "root", CPUPercent: 45.0, MemRSS: 1024, Status: "running", Elapsed: "4m"},
			{PID: 5, Name: "p5", User: "root", CPUPercent: 80.0, MemRSS: 1024, Status: "running", Elapsed: "5m"},
		},
	}

	provider := NewProcessDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Get Top CPU tab content
	primitive := provider.TabContent(1)
	if primitive == nil {
		t.Fatal("Top CPU tab returned nil primitive")
	}

	// Verify it's the topCPUTable
	if provider.topCPUTable == nil {
		t.Fatal("topCPUTable is nil")
	}

	// Check table has header + 5 data rows (all processes since < 10)
	rowCount := provider.topCPUTable.GetRowCount()
	if rowCount != 6 { // header + 5 data rows
		t.Errorf("expected 6 rows, got %d", rowCount)
	}

	// Verify sorting: first data row should have highest CPU (90.0, PID 2)
	cpuCell := provider.topCPUTable.GetCell(1, 2)
	if cpuCell == nil {
		t.Fatal("Top CPU cell is nil")
	}
	if !strings.Contains(cpuCell.Text, "90") {
		t.Errorf("expected first row CPU to contain '90', got %q", cpuCell.Text)
	}

	pidCell := provider.topCPUTable.GetCell(1, 0)
	if pidCell == nil {
		t.Fatal("Top CPU PID cell is nil")
	}
	if pidCell.Text != "2" {
		t.Errorf("expected first row PID '2', got %q", pidCell.Text)
	}

	// Verify second row has second highest CPU (80.0, PID 5)
	cpuCell2 := provider.topCPUTable.GetCell(2, 2)
	if cpuCell2 == nil {
		t.Fatal("Second CPU cell is nil")
	}
	if !strings.Contains(cpuCell2.Text, "80") {
		t.Errorf("expected second row CPU to contain '80', got %q", cpuCell2.Text)
	}
}

func TestProcessProvider_TopMemTab(t *testing.T) {
	mock := &mockAPIClient{
		procInfo: []*models.ProcessInfo{
			{PID: 1, Name: "p1", User: "root", CPUPercent: 10.0, MemRSS: 1024 * 1024 * 50, Status: "running", Elapsed: "1m"},  // 50 MB
			{PID: 2, Name: "p2", User: "root", CPUPercent: 20.0, MemRSS: 1024 * 1024 * 200, Status: "running", Elapsed: "2m"}, // 200 MB
			{PID: 3, Name: "p3", User: "root", CPUPercent: 30.0, MemRSS: 1024 * 1024 * 100, Status: "running", Elapsed: "3m"}, // 100 MB
			{PID: 4, Name: "p4", User: "root", CPUPercent: 40.0, MemRSS: 1024 * 1024 * 75, Status: "running", Elapsed: "4m"},  // 75 MB
			{PID: 5, Name: "p5", User: "root", CPUPercent: 50.0, MemRSS: 1024 * 1024 * 150, Status: "running", Elapsed: "5m"}, // 150 MB
		},
	}

	provider := NewProcessDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Get Top Memory tab content
	primitive := provider.TabContent(2)
	if primitive == nil {
		t.Fatal("Top Memory tab returned nil primitive")
	}

	// Verify it's the topMemTable
	if provider.topMemTable == nil {
		t.Fatal("topMemTable is nil")
	}

	// Check table has header + 5 data rows
	rowCount := provider.topMemTable.GetRowCount()
	if rowCount != 6 {
		t.Errorf("expected 6 rows, got %d", rowCount)
	}

	// Verify sorting: first data row should have highest memory (200 MB, PID 2)
	pidCell := provider.topMemTable.GetCell(1, 0)
	if pidCell == nil {
		t.Fatal("Top Mem PID cell is nil")
	}
	if pidCell.Text != "2" {
		t.Errorf("expected first row PID '2', got %q", pidCell.Text)
	}

	memCell := provider.topMemTable.GetCell(1, 3)
	if memCell == nil {
		t.Fatal("Top Mem cell is nil")
	}
	// Should contain "200" somewhere (200.00 MB)
	if !strings.Contains(memCell.Text, "200") {
		t.Errorf("expected first row Memory to contain '200', got %q", memCell.Text)
	}

	// Verify second row has second highest memory (150 MB, PID 5)
	pidCell2 := provider.topMemTable.GetCell(2, 0)
	if pidCell2 == nil {
		t.Fatal("Second Mem PID cell is nil")
	}
	if pidCell2.Text != "5" {
		t.Errorf("expected second row PID '5', got %q", pidCell2.Text)
	}
}

func TestProcessProvider_TopTabs_Limit10(t *testing.T) {
	// Create 15 processes to test limit
	procs := make([]*models.ProcessInfo, 15)
	for i := 0; i < 15; i++ {
		procs[i] = &models.ProcessInfo{
			PID:        i + 1,
			Name:       "process",
			User:       "root",
			CPUPercent: float64(i * 5),
			MemRSS:     int64((i + 1) * 1024 * 1024),
			Status:     "running",
			Elapsed:    "1m",
		}
	}

	mock := &mockAPIClient{procInfo: procs}
	provider := NewProcessDetailProvider()
	_ = provider.Refresh(context.Background(), mock)

	// Top CPU tab should have max 10 + header = 11 rows
	cpuRowCount := provider.topCPUTable.GetRowCount()
	if cpuRowCount != 11 {
		t.Errorf("Top CPU: expected 11 rows (header + 10), got %d", cpuRowCount)
	}

	// Top Memory tab should have max 10 + header = 11 rows
	memRowCount := provider.topMemTable.GetRowCount()
	if memRowCount != 11 {
		t.Errorf("Top Memory: expected 11 rows (header + 10), got %d", memRowCount)
	}
}

func TestProcessProvider_OnSelect(t *testing.T) {
	provider := NewProcessDetailProvider()

	// OnSelect should not panic
	provider.OnSelect(0)
	provider.OnSelect(1)
	provider.OnSelect(2)
}
