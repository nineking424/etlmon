<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# handler

## Purpose

HTTP request handlers for each API endpoint. Each handler queries the repository and returns JSON responses.

## Key Files

| File | Description |
|------|-------------|
| `fs.go` | Filesystem usage endpoints |
| `paths.go` | Path statistics endpoints |
| `logs.go` | Log viewing endpoints |
| `processes.go` | Process monitoring endpoints |
| `cron.go` | Cron job endpoints |
| `xferlog.go` | Xferlog endpoints |
| `health.go` | Health check endpoint |
| `admin.go` | Admin/maintenance endpoints |

## For AI Agents

### Working In This Directory

- One file per domain (fs, paths, logs, etc.)
- Handlers receive repository and controller as dependencies
- Use `http.Request.Context()` for request-scoped context
- Parse query params for pagination (`limit`, `offset`)
- Validate input before processing

### Handler Pattern

```go
type FSHandler struct {
    repo *repository.Repository
}

func NewFSHandler(repo *repository.Repository) *FSHandler {
    return &FSHandler{repo: repo}
}

func (h *FSHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    usage, err := h.repo.FS.GetLatest(ctx)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err)
        return
    }

    writeJSON(w, http.StatusOK, Response{Data: usage})
}
```

### Pagination

```go
func getPagination(r *http.Request) (limit, offset int) {
    limit = 100 // default
    offset = 0

    if l := r.URL.Query().Get("limit"); l != "" {
        limit, _ = strconv.Atoi(l)
    }
    if o := r.URL.Query().Get("offset"); o != "" {
        offset, _ = strconv.Atoi(o)
    }

    // Cap limit
    if limit > 1000 {
        limit = 1000
    }
    return
}
```

### Command Handlers (Kill, Scan)

```go
func (h *ProcessHandler) Kill(w http.ResponseWriter, r *http.Request) {
    var req KillRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, err)
        return
    }

    // Require confirmation
    if !req.Confirm {
        writeError(w, http.StatusBadRequest, errors.New("confirmation required"))
        return
    }

    // Execute via controller
    if err := h.ctrl.KillProcess(r.Context(), req.PID, req.Signal); err != nil {
        writeError(w, http.StatusInternalServerError, err)
        return
    }

    writeJSON(w, http.StatusOK, Response{Data: "ok"})
}
```

<!-- MANUAL: -->
