<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# views

## Purpose

TUI views for each monitoring domain. Each view fetches data from Node API and renders it in the terminal.

## Key Files

| File | Description |
|------|-------------|
| `fs.go` | Filesystem usage view |
| `paths.go` | Path statistics view |
| `logs.go` | Log viewer with follow mode |
| `processes.go` | Process list with kill action |
| `cron.go` | Cron jobs view |
| `xferlog.go` | FTP transfer log view |
| `help.go` | Help overlay |

## For AI Agents

### Working In This Directory

- Each view implements the `View` interface
- Views should be reusable across nodes
- Handle loading states and errors gracefully
- Use color coding for warnings/errors

### FS View Example

```go
type FSView struct {
    table *tview.Table
}

func NewFSView() *FSView {
    v := &FSView{
        table: tview.NewTable().
            SetBorders(true).
            SetSelectable(true, false),
    }
    v.setupHeaders()
    return v
}

func (v *FSView) Name() string { return "fs" }

func (v *FSView) Primitive() tview.Primitive { return v.table }

func (v *FSView) Refresh(ctx context.Context, client *client.Client) error {
    usage, err := client.GetFilesystemUsage(ctx)
    if err != nil {
        return err
    }

    v.table.Clear()
    v.setupHeaders()

    for i, fs := range usage {
        row := i + 1
        v.table.SetCell(row, 0, tview.NewTableCell(fs.MountPoint))
        v.table.SetCell(row, 1, tview.NewTableCell(formatBytes(fs.TotalBytes)))
        v.table.SetCell(row, 2, tview.NewTableCell(formatBytes(fs.UsedBytes)))
        v.table.SetCell(row, 3, tview.NewTableCell(formatBytes(fs.AvailBytes)))

        // Color code by usage
        percentCell := tview.NewTableCell(fmt.Sprintf("%.1f%%", fs.UsedPercent))
        if fs.UsedPercent > 90 {
            percentCell.SetTextColor(tcell.ColorRed)
        } else if fs.UsedPercent > 75 {
            percentCell.SetTextColor(tcell.ColorYellow)
        }
        v.table.SetCell(row, 4, percentCell)
    }
    return nil
}

func (v *FSView) setupHeaders() {
    headers := []string{"Mount", "Total", "Used", "Avail", "Use%"}
    for i, h := range headers {
        v.table.SetCell(0, i,
            tview.NewTableCell(h).
                SetTextColor(tcell.ColorWhite).
                SetAttributes(tcell.AttrBold))
    }
}
```

### Log View (Special: Follow Mode)

```go
type LogView struct {
    textView *tview.TextView
    logName  string
    follow   bool
}

func (v *LogView) Refresh(ctx context.Context, client *client.Client) error {
    lines, err := client.GetLogLines(ctx, v.logName, 100, 0)
    if err != nil {
        return err
    }

    v.textView.Clear()
    for _, line := range lines {
        fmt.Fprintln(v.textView, line.Line)
    }

    if v.follow {
        v.textView.ScrollToEnd()
    }
    return nil
}

// Toggle follow mode with 'f' key
func (v *LogView) ToggleFollow() {
    v.follow = !v.follow
}
```

### Process View (Special: Kill Action)

```go
func (v *ProcessView) setupKeyBindings() {
    v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Rune() == 'd' {
            v.showKillDialog()
            return nil
        }
        return event
    })
}

func (v *ProcessView) showKillDialog() {
    // Show confirmation dialog
    // On confirm, call client.KillProcess()
}
```

<!-- MANUAL: -->
