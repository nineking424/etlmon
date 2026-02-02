<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# api

## Purpose

HTTP API Gateway for the Node daemon. Exposes REST endpoints for UI clients to query collected data and execute commands.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `handler/` | HTTP request handlers (see `handler/AGENTS.md`) |
| `middleware/` | Middleware (logging, auth, rate-limit) (see `middleware/AGENTS.md`) |

## Key Files

| File | Description |
|------|-------------|
| `server.go` | HTTP server setup and routing |
| `routes.go` | Route definitions |

## For AI Agents

### Working In This Directory

- Use `net/http` standard library (or chi/gorilla mux for routing)
- All responses are JSON
- Query endpoints: GET (read-only transactions)
- Command endpoints: POST/DELETE (with confirmation)
- Always include proper error responses

### API Structure

```
/api/v1/
├── fs                  GET    - Filesystem usage
├── paths               GET    - Path statistics
├── paths/scan          POST   - Trigger scan
├── logs                GET    - Log file list
├── logs/{name}         GET    - Log lines (paginated)
├── processes           GET    - Process list
├── processes/{pid}/kill POST  - Kill process
├── cron                GET    - Cron jobs
├── cron/refresh        POST   - Refresh cron
├── xferlog             GET    - Xferlog entries
├── health              GET    - Health check
└── admin/db/compact    POST   - DB maintenance
```

### Response Format

```go
// Success response
type Response struct {
    Data  interface{} `json:"data"`
    Meta  *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total  int `json:"total,omitempty"`
    Limit  int `json:"limit,omitempty"`
    Offset int `json:"offset,omitempty"`
}

// Error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}
```

### Server Setup

```go
func NewServer(cfg *config.API, repo *repository.Repository, ctrl *controller.Controller) *Server {
    mux := http.NewServeMux()

    // Apply middleware
    handler := middleware.Logging(
        middleware.Recovery(mux),
    )

    return &Server{
        httpServer: &http.Server{
            Addr:    cfg.Listen,
            Handler: handler,
        },
    }
}
```

### Dependencies

#### Internal
- `internal/db/repository` - For data queries
- `internal/controller` - For command execution
- `internal/config` - For API configuration

<!-- MANUAL: -->
