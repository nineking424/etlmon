package ui

import (
	"context"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui/client"
)

// APIClient defines the interface for all API operations needed by views
type APIClient interface {
	// Filesystem operations
	GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error)

	// Path operations
	GetPathStats(ctx context.Context) ([]*models.PathStats, error)
	TriggerScan(ctx context.Context, paths []string) error

	// Process operations
	GetProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error)

	// Log operations
	GetLogFiles(ctx context.Context) ([]models.LogFileInfo, error)
	GetLogEntriesByName(ctx context.Context, name string, limit int) ([]*models.LogEntry, error)

	// Config operations
	GetConfig(ctx context.Context) (*config.NodeConfig, error)
	SaveConfig(ctx context.Context, cfg *config.NodeConfig) error
}

// Compile-time check to ensure *client.Client implements APIClient
var _ APIClient = (*client.Client)(nil)
