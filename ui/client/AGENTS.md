<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# client

## Purpose

HTTP client for communicating with Node API. Provides typed methods for all API endpoints.

## Key Files

| File | Description |
|------|-------------|
| `client.go` | Main client with connection management |
| `fs.go` | Filesystem API methods |
| `paths.go` | Paths API methods |
| `logs.go` | Logs API methods |
| `processes.go` | Processes API methods |
| `cron.go` | Cron API methods |
| `xferlog.go` | Xferlog API methods |

## For AI Agents

### Working In This Directory

- Use standard `net/http` client
- Set reasonable timeouts
- Handle errors consistently
- Return typed responses (not raw JSON)

### Client Structure

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: strings.TrimSuffix(baseURL, "/"),
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *Client) SetTimeout(d time.Duration) {
    c.httpClient.Timeout = d
}
```

### Request Helper

```go
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
    url := c.baseURL + path

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
    }

    var apiResp struct {
        Data json.RawMessage `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return fmt.Errorf("decode response: %w", err)
    }

    return json.Unmarshal(apiResp.Data, result)
}

func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
    url := c.baseURL + path

    var bodyReader io.Reader
    if body != nil {
        data, _ := json.Marshal(body)
        bodyReader = bytes.NewReader(data)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    // ... similar to get()
}
```

### API Methods

```go
// Filesystem
func (c *Client) GetFilesystemUsage(ctx context.Context) ([]*FilesystemUsage, error) {
    var result []*FilesystemUsage
    err := c.get(ctx, "/api/v1/fs", &result)
    return result, err
}

// Paths
func (c *Client) GetPathStats(ctx context.Context) ([]*PathStats, error) {
    var result []*PathStats
    err := c.get(ctx, "/api/v1/paths", &result)
    return result, err
}

func (c *Client) TriggerScan(ctx context.Context, paths []string) error {
    req := map[string][]string{"paths": paths}
    return c.post(ctx, "/api/v1/paths/scan", req, nil)
}

// Logs
func (c *Client) GetLogLines(ctx context.Context, logName string, limit, offset int) ([]*LogLine, error) {
    var result []*LogLine
    path := fmt.Sprintf("/api/v1/logs/%s?limit=%d&offset=%d", logName, limit, offset)
    err := c.get(ctx, path, &result)
    return result, err
}

// Processes
func (c *Client) GetProcesses(ctx context.Context) ([]*ProcessStats, error) {
    var result []*ProcessStats
    err := c.get(ctx, "/api/v1/processes", &result)
    return result, err
}

func (c *Client) KillProcess(ctx context.Context, pid int, signal string, confirm bool) error {
    req := map[string]interface{}{
        "signal":  signal,
        "confirm": confirm,
    }
    path := fmt.Sprintf("/api/v1/processes/%d/kill", pid)
    return c.post(ctx, path, req, nil)
}

// Health
func (c *Client) Health(ctx context.Context) error {
    return c.get(ctx, "/api/v1/health", nil)
}
```

### Error Types

```go
type APIError struct {
    StatusCode int
    Message    string
    Code       string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
    var apiErr *APIError
    return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}
```

<!-- MANUAL: -->
