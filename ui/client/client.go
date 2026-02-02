package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an HTTP client for the etlmon API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// APIError represents an error response from the API
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(d time.Duration) {
	c.httpClient.Timeout = d
}

// get performs a GET request and unmarshals the response into result
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error   string `json:"error"`
			Code    string `json:"code"`
			Details string `json:"details"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// Parse successful response
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return fmt.Errorf("unmarshal response wrapper: %w", err)
	}

	// Unmarshal data into result
	if err := json.Unmarshal(wrapper.Data, result); err != nil {
		return fmt.Errorf("unmarshal response data: %w", err)
	}

	return nil
}

// post performs a POST request with a JSON body and unmarshals the response into result
func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	url := c.baseURL + path

	// Marshal body
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error   string `json:"error"`
			Code    string `json:"code"`
			Details string `json:"details"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// Parse successful response
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return fmt.Errorf("unmarshal response wrapper: %w", err)
	}

	// Unmarshal data into result
	if err := json.Unmarshal(wrapper.Data, result); err != nil {
		return fmt.Errorf("unmarshal response data: %w", err)
	}

	return nil
}
