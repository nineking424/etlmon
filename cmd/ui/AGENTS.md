<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# ui

## Purpose

Entry point for the etlmon TUI client. The client:
1. Loads configuration from `ui.yaml` (Node addresses)
2. Initializes tview application
3. Connects to configured Nodes via HTTP API
4. Renders monitoring views
5. Handles user input and commands

## Key Files

| File | Description |
|------|-------------|
| `main.go` | TUI application entry point |

## For AI Agents

### Working In This Directory

- Single `main.go` file only
- All view logic should be in `ui/views/` package
- All API client logic should be in `ui/client/` package
- UI is stateless - all data comes from Node API
- Default config path: `~/.config/etlmon/ui.yaml`

### Startup Sequence

```go
1. Parse CLI flags (-c config path)
2. Load ui.yaml (Node list)
3. Initialize tview Application
4. Create main layout (Node tabs, views)
5. Connect to default Node
6. Start refresh loop (poll API)
7. Run application
```

### Key Bindings (Global)

| Key | Action |
|-----|--------|
| `q` | Quit application |
| `?` | Show help |
| `Tab` | Switch Node |
| `1-6` | Switch view (FS/Path/Log/Process/Cron/Xferlog) |

### Example Usage

```bash
# Run with default config
./etlmon-ui

# Run with custom config
./etlmon-ui -c /path/to/ui.yaml

# Connect to specific node
./etlmon-ui --node prod-etl-01
```

### Dependencies

#### Internal
- `internal/config` - Configuration loading (ui.yaml parsing)
- `ui/views` - All TUI views
- `ui/client` - HTTP client for Node API
- `ui/components` - Reusable TUI widgets

#### External
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal handling

<!-- MANUAL: -->
