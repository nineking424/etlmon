<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# components

## Purpose

Reusable TUI components/widgets used across multiple views.

## Key Files

| File | Description |
|------|-------------|
| `statusbar.go` | Bottom status bar showing node info |
| `dialog.go` | Confirmation and input dialogs |
| `progress.go` | Progress/loading indicators |
| `filter.go` | Filter input for lists |
| `table.go` | Enhanced table with sorting |

## For AI Agents

### Working In This Directory

- Components should be stateless or minimally stateful
- Accept callbacks for user actions
- Follow tview patterns for primitives
- Support theming/colors

### Status Bar

```go
type StatusBar struct {
    *tview.TextView
    nodeName   string
    lastUpdate time.Time
    status     string
}

func NewStatusBar() *StatusBar {
    s := &StatusBar{
        TextView: tview.NewTextView().
            SetDynamicColors(true).
            SetTextAlign(tview.AlignLeft),
    }
    return s
}

func (s *StatusBar) Update(nodeName string, status string) {
    s.nodeName = nodeName
    s.status = status
    s.lastUpdate = time.Now()
    s.render()
}

func (s *StatusBar) render() {
    s.Clear()
    fmt.Fprintf(s, "[green]Node:[white] %s | [green]Status:[white] %s | [green]Updated:[white] %s",
        s.nodeName, s.status, s.lastUpdate.Format("15:04:05"))
}
```

### Confirmation Dialog

```go
type ConfirmDialog struct {
    *tview.Modal
    onConfirm func()
    onCancel  func()
}

func NewConfirmDialog(title, message string, onConfirm, onCancel func()) *ConfirmDialog {
    d := &ConfirmDialog{
        Modal:     tview.NewModal(),
        onConfirm: onConfirm,
        onCancel:  onCancel,
    }

    d.SetText(message).
        AddButtons([]string{"Confirm", "Cancel"}).
        SetDoneFunc(func(buttonIndex int, buttonLabel string) {
            if buttonLabel == "Confirm" {
                d.onConfirm()
            } else {
                d.onCancel()
            }
        })

    return d
}
```

### Filter Input

```go
type FilterInput struct {
    *tview.InputField
    onChange func(text string)
}

func NewFilterInput(placeholder string, onChange func(string)) *FilterInput {
    f := &FilterInput{
        InputField: tview.NewInputField().
            SetLabel("Filter: ").
            SetPlaceholder(placeholder).
            SetFieldWidth(30),
        onChange: onChange,
    }

    f.SetChangedFunc(func(text string) {
        f.onChange(text)
    })

    return f
}
```

### Progress Indicator

```go
type Progress struct {
    *tview.TextView
    running bool
    frames  []string
    current int
}

func NewProgress() *Progress {
    return &Progress{
        TextView: tview.NewTextView(),
        frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
    }
}

func (p *Progress) Start() {
    p.running = true
    go func() {
        for p.running {
            p.current = (p.current + 1) % len(p.frames)
            p.SetText(p.frames[p.current])
            time.Sleep(100 * time.Millisecond)
        }
    }()
}

func (p *Progress) Stop() {
    p.running = false
    p.SetText("")
}
```

<!-- MANUAL: -->
