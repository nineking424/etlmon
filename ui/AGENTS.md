<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# ui

## Purpose

TUI (Terminal User Interface) client for etlmon. Connects to Node daemons via HTTP API and displays monitoring data in a terminal interface using tview.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `views/` | TUI views for each domain (see `views/AGENTS.md`) |
| `components/` | Reusable TUI widgets (see `components/AGENTS.md`) |
| `client/` | HTTP client for Node API (see `client/AGENTS.md`) |

## For AI Agents

### Working In This Directory

- Use `github.com/rivo/tview` for TUI framework
- UI is stateless - all data comes from Node API
- Implement keyboard navigation (vim-style)
- Handle API errors gracefully with user feedback

### Application Structure

```go
type App struct {
    tview    *tview.Application
    client   *client.Client
    config   *config.UIConfig

    // Main layout
    pages    *tview.Pages
    nodeList *tview.List
    views    map[string]View

    // State
    currentNode string
    refreshing  bool
}

func NewApp(cfg *config.UIConfig) *App {
    app := &App{
        tview:  tview.NewApplication(),
        config: cfg,
        views:  make(map[string]View),
    }

    app.setupLayout()
    app.setupKeyBindings()

    return app
}
```

### View Interface

```go
// View is the interface for all TUI views
type View interface {
    // Name returns view identifier
    Name() string

    // Primitive returns the tview primitive
    Primitive() tview.Primitive

    // Refresh updates view with latest data
    Refresh(ctx context.Context, client *client.Client) error

    // Focus handles view becoming active
    Focus()
}
```

### Key Bindings

| Key | Scope | Action |
|-----|-------|--------|
| `q` | Global | Quit application |
| `?` | Global | Show help |
| `Tab` | Global | Next node |
| `Shift+Tab` | Global | Previous node |
| `1` | Global | FS view |
| `2` | Global | Paths view |
| `3` | Global | Logs view |
| `4` | Global | Process view |
| `5` | Global | Cron view |
| `6` | Global | Xferlog view |
| `j`/`k` | List | Navigate down/up |
| `Enter` | List | Select/expand |
| `Esc` | Any | Back/cancel |

### Refresh Loop

```go
func (a *App) startRefreshLoop(ctx context.Context) {
    ticker := time.NewTicker(a.config.UI.RefreshInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            a.refreshCurrentView(ctx)
        }
    }
}

func (a *App) refreshCurrentView(ctx context.Context) {
    if a.refreshing {
        return
    }
    a.refreshing = true
    defer func() { a.refreshing = false }()

    currentView := a.getCurrentView()
    if err := currentView.Refresh(ctx, a.client); err != nil {
        a.showError(err)
    }

    a.tview.Draw()
}
```

<!-- MANUAL: -->
