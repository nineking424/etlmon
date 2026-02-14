package client

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
)

// GetLogEntries retrieves log entries from the API
func (c *Client) GetLogEntries(ctx context.Context) ([]*models.LogEntry, error) {
	var entries []*models.LogEntry
	if err := c.get(ctx, "/api/v1/logs", &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetLogEntriesByName retrieves log entries for a specific log name
func (c *Client) GetLogEntriesByName(ctx context.Context, name string, limit int) ([]*models.LogEntry, error) {
	var entries []*models.LogEntry
	path := fmt.Sprintf("/api/v1/logs?name=%s&limit=%d", name, limit)
	if err := c.get(ctx, path, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetLogFiles retrieves log file metadata from the API
func (c *Client) GetLogFiles(ctx context.Context) ([]models.LogFileInfo, error) {
	var files []models.LogFileInfo
	if err := c.get(ctx, "/api/v1/logs/files", &files); err != nil {
		return nil, err
	}
	return files, nil
}
