package client

import (
	"context"

	"github.com/etlmon/etlmon/pkg/models"
)

// GetProcessInfo retrieves process info from the API
func (c *Client) GetProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error) {
	var procs []*models.ProcessInfo
	if err := c.get(ctx, "/api/v1/processes", &procs); err != nil {
		return nil, err
	}
	return procs, nil
}
