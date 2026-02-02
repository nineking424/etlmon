package views

import (
	"context"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPathsClient implements a mock client for testing
type mockPathsClient struct {
	stats []*models.PathStats
	err   error
}

func (m *mockPathsClient) GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return nil, nil
}

func (m *mockPathsClient) GetPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return m.stats, m.err
}

func (m *mockPathsClient) TriggerScan(ctx context.Context, paths []string) error {
	return nil
}

func TestPathsView_Name_ReturnsPaths(t *testing.T) {
	view := NewPathsView()
	assert.Equal(t, "paths", view.Name())
}

func TestPathsView_Refresh_PopulatesTable(t *testing.T) {
	// Setup mock client
	mockClient := &mockPathsClient{
		stats: []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      1500,
				DirCount:       25,
				ScanDurationMs: 250,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
			{
				Path:           "/data/error",
				FileCount:      0,
				DirCount:       0,
				ScanDurationMs: 0,
				Status:         "ERROR",
				ErrorMessage:   "permission denied",
				CollectedAt:    time.Now(),
			},
		},
	}

	// Create view and refresh
	view := NewPathsView()
	err := view.refresh(context.Background(), mockClient)

	// Assert
	require.NoError(t, err)

	// Verify table has rows (header + 2 data rows)
	assert.Equal(t, 3, view.table.GetRowCount())

	// Verify header
	cell := view.table.GetCell(0, 0)
	assert.Equal(t, "Path", cell.Text)

	// Verify first data row
	cell = view.table.GetCell(1, 0)
	assert.Equal(t, "/data/logs", cell.Text)

	// Verify second row shows ERROR status
	cell = view.table.GetCell(2, 4) // Status column
	assert.Equal(t, "ERROR", cell.Text)
}
