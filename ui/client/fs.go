package client

import (
	"context"

	"github.com/etlmon/etlmon/pkg/models"
)

// GetFilesystemUsage retrieves filesystem usage statistics from the API
func (c *Client) GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	var usage []*models.FilesystemUsage
	if err := c.get(ctx, "/api/v1/fs", &usage); err != nil {
		return nil, err
	}
	return usage, nil
}
