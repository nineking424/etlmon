package views

import (
	"context"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient implements a mock client for testing
type mockFSClient struct {
	usage []*models.FilesystemUsage
	err   error
}

func (m *mockFSClient) GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return m.usage, m.err
}

func (m *mockFSClient) GetPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return nil, nil
}

func (m *mockFSClient) TriggerScan(ctx context.Context, paths []string) error {
	return nil
}

func TestFSView_Name_ReturnsFS(t *testing.T) {
	view := NewFSView()
	assert.Equal(t, "fs", view.Name())
}

func TestFSView_Refresh_PopulatesTable(t *testing.T) {
	// Setup mock client
	mockClient := &mockFSClient{
		usage: []*models.FilesystemUsage{
			{
				MountPoint:  "/data",
				TotalBytes:  1000000000,
				UsedBytes:   600000000,
				AvailBytes:  400000000,
				UsedPercent: 60.0,
				CollectedAt: time.Now(),
			},
			{
				MountPoint:  "/logs",
				TotalBytes:  500000000,
				UsedBytes:   450000000,
				AvailBytes:  50000000,
				UsedPercent: 90.0,
				CollectedAt: time.Now(),
			},
		},
	}

	// Create view and refresh
	view := NewFSView()
	err := view.refresh(context.Background(), mockClient)

	// Assert
	require.NoError(t, err)

	// Verify table has rows (header + 2 data rows)
	assert.Equal(t, 3, view.table.GetRowCount())

	// Verify header
	cell := view.table.GetCell(0, 0)
	assert.Equal(t, "Mount", cell.Text)

	// Verify first data row
	cell = view.table.GetCell(1, 0)
	assert.Equal(t, "/data", cell.Text)

	// Verify second data row has high usage warning
	cell = view.table.GetCell(2, 4) // Use% column
	assert.Contains(t, cell.Text, "90.0%")
}
