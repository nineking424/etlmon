package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/etlmon/etlmon/internal/config"
)

// GetConfig retrieves the current node configuration
func (c *Client) GetConfig(ctx context.Context) (*config.NodeConfig, error) {
	var cfg config.NodeConfig
	if err := c.get(ctx, "/api/v1/config", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig sends updated configuration to the node
func (c *Client) SaveConfig(ctx context.Context, cfg *config.NodeConfig) error {
	url := c.baseURL + "/api/v1/config"

	bodyBytes, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return nil
}
