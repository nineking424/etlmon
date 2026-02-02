package client

import (
	"context"

	"github.com/etlmon/etlmon/pkg/models"
)

// GetPathStats retrieves path statistics from the API
func (c *Client) GetPathStats(ctx context.Context) ([]*models.PathStats, error) {
	var stats []*models.PathStats
	if err := c.get(ctx, "/api/v1/paths", &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// TriggerScan triggers a scan for the specified paths
func (c *Client) TriggerScan(ctx context.Context, paths []string) error {
	body := map[string]interface{}{
		"paths": paths,
	}
	var result map[string]string
	return c.post(ctx, "/api/v1/paths/scan", body, &result)
}
