package views

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOverviewClient struct {
	usage   []*models.FilesystemUsage
	stats   []*models.PathStats
	fsErr   error
	pathErr error
}

func (m *mockOverviewClient) GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return m.usage, m.fsErr
}

func (m *mockOverviewClient) GetPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return m.stats, m.pathErr
}

func (m *mockOverviewClient) TriggerScan(ctx context.Context, paths []string) error {
	return nil
}

func TestOverviewView_Name_ReturnsOverview(t *testing.T) {
	view := NewOverviewView()
	assert.Equal(t, "overview", view.Name())
}

func TestOverviewView_Refresh_BothSucceed(t *testing.T) {
	mockClient := &mockOverviewClient{
		usage: []*models.FilesystemUsage{
			{
				MountPoint:  "/",
				TotalBytes:  245000000000,
				UsedBytes:   218000000000,
				AvailBytes:  27000000000,
				UsedPercent: 89.1,
				CollectedAt: time.Now(),
			},
		},
		stats: []*models.PathStats{
			{
				Path:           "/data/logs",
				FileCount:      1500,
				DirCount:       25,
				ScanDurationMs: 250,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
		},
	}

	view := NewOverviewView()
	err := view.refresh(context.Background(), mockClient)

	require.NoError(t, err)
	assert.Equal(t, 2, view.pathBox.GetRowCount()) // header + 1 data row
}

func TestOverviewView_Refresh_FSErrorOnly(t *testing.T) {
	mockClient := &mockOverviewClient{
		fsErr: fmt.Errorf("connection refused"),
		stats: []*models.PathStats{
			{
				Path:           "/data",
				FileCount:      100,
				DirCount:       10,
				ScanDurationMs: 50,
				Status:         "OK",
				CollectedAt:    time.Now(),
			},
		},
	}

	view := NewOverviewView()
	err := view.refresh(context.Background(), mockClient)

	// Should not return error if only one fails
	require.NoError(t, err)
}

func TestOverviewView_Refresh_BothFail(t *testing.T) {
	mockClient := &mockOverviewClient{
		fsErr:   fmt.Errorf("fs error"),
		pathErr: fmt.Errorf("path error"),
	}

	view := NewOverviewView()
	err := view.refresh(context.Background(), mockClient)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fs:")
	assert.Contains(t, err.Error(), "paths:")
}
