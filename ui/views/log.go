package views

import (
	"context"
	"fmt"
	"strconv"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// logViewClient defines the interface for log view API operations
type logViewClient interface {
	GetLogEntriesByName(ctx context.Context, name string, limit int) ([]*models.LogEntry, error)
	GetConfig(ctx context.Context) (*config.NodeConfig, error)
}

// LogView displays log file list and real-time log content
type LogView struct {
	outerFlex      *tview.Flex       // root: contentFlex + hintBar
	contentFlex    *tview.Flex       // horizontal: logTable + optionally logDetail
	logTable       *tview.Table      // left: list of log files
	logDetail      *tview.TextView   // right: log content for selected file
	hintBar        *tview.TextView
	selectedLog    string
	showingDetail  bool
	tviewApp       *tview.Application
	apiClient      logViewClient
	logConfigs     []config.LogMonitorConfig
	onStatusChange func(msg string, isError bool)
}

// NewLogView creates a new log view
func NewLogView() *LogView {
	v := &LogView{}

	// Create log file table
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).
		SetTitle(" Log Files ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	// Table headers
	headers := []string{"Name", "Path", "MaxLines"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetSelectable(false)
		if i == 1 {
			cell.SetExpansion(1)
		}
		table.SetCell(0, i, cell)
	}

	// Table key handler: Enter to select log
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			v.selectLog()
			return nil
		}
		return event
	})

	v.logTable = table

	// Create detail view (hidden initially)
	detail := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)
	detail.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)
	v.logDetail = detail

	// Hint bar
	hintBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	hintBar.SetText("[teal]Enter[silver]=view log  [teal]Esc[silver]=close  [teal]Tab[silver]=switch pane  [teal]\u2191\u2193[silver]=navigate")
	hintBar.SetBackgroundColor(theme.BgStatusBar)
	v.hintBar = hintBar

	// Content flex (initially table only, full width)
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)
	v.contentFlex = contentFlex

	// Outer flex: content + hint bar
	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(hintBar, 1, 0, false)

	// View-level key handler: Esc and Tab
	outerFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if v.showingDetail {
				v.closeDetail()
				return nil
			}
			return event
		}
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			if v.showingDetail && v.tviewApp != nil {
				if v.logTable.HasFocus() {
					v.tviewApp.SetFocus(v.logDetail)
				} else {
					v.tviewApp.SetFocus(v.logTable)
				}
				return nil
			}
		}
		return event
	})

	v.outerFlex = outerFlex

	return v
}

// SetApp sets the tview application reference
func (v *LogView) SetApp(app *tview.Application) {
	v.tviewApp = app
}

// Name returns the view name
func (v *LogView) Name() string {
	return "logs"
}

// Primitive returns the tview primitive
func (v *LogView) Primitive() tview.Primitive {
	return v.outerFlex
}

// Refresh updates the view with fresh data
func (v *LogView) Refresh(ctx context.Context, c *client.Client) error {
	v.apiClient = c
	return v.refresh(ctx, c)
}

// refresh is the internal method that accepts logViewClient interface
func (v *LogView) refresh(ctx context.Context, c logViewClient) error {
	// Get config for log file list
	cfg, err := c.GetConfig(ctx)
	if err != nil {
		return err
	}
	v.logConfigs = cfg.Logs
	v.refreshTable()

	// If detail panel is showing, refresh its content
	if v.showingDetail && v.selectedLog != "" {
		entries, err := c.GetLogEntriesByName(ctx, v.selectedLog, 500)
		if err != nil {
			return err
		}
		v.refreshDetail(entries)
	}

	return nil
}

// Focus sets focus on the log table
func (v *LogView) Focus() {
	if v.tviewApp != nil {
		v.tviewApp.SetFocus(v.logTable)
	}
}

// SetStatusCallback sets the callback for status messages
func (v *LogView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
}

func (v *LogView) refreshTable() {
	// Clear existing rows (keep header)
	for i := v.logTable.GetRowCount() - 1; i > 0; i-- {
		v.logTable.RemoveRow(i)
	}

	if len(v.logConfigs) == 0 {
		v.logTable.SetCell(1, 0, tview.NewTableCell("(no log files configured)").
			SetTextColor(theme.FgMuted).
			SetExpansion(1))
		return
	}

	for i, l := range v.logConfigs {
		row := i + 1
		v.logTable.SetCell(row, 0, tview.NewTableCell(l.Name).
			SetTextColor(theme.FgPrimary))
		v.logTable.SetCell(row, 1, tview.NewTableCell(l.Path).
			SetTextColor(theme.FgSecondary).
			SetExpansion(1))
		v.logTable.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(l.MaxLines)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
	}
}

func (v *LogView) refreshDetail(entries []*models.LogEntry) {
	v.logDetail.Clear()

	if len(entries) == 0 {
		fmt.Fprintf(v.logDetail, " [darkgray]No log entries[-]\n")
		return
	}

	for _, entry := range entries {
		timestamp := entry.CreatedAt.Format("15:04:05")
		fmt.Fprintf(v.logDetail, "[teal]%s[-] %s\n", timestamp, entry.Line)
	}

	v.logDetail.ScrollToEnd()
}

func (v *LogView) selectLog() {
	row, _ := v.logTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(v.logConfigs) {
		return
	}

	logName := v.logConfigs[idx].Name
	v.selectedLog = logName
	v.logDetail.SetTitle(fmt.Sprintf(" %s ", logName))

	if !v.showingDetail {
		v.showDetail()
	}

	// Fetch log entries asynchronously
	if v.apiClient != nil && v.tviewApp != nil {
		go func() {
			ctx := context.Background()
			entries, err := v.apiClient.GetLogEntriesByName(ctx, logName, 500)
			if err != nil {
				v.tviewApp.QueueUpdateDraw(func() {
					v.logDetail.Clear()
					fmt.Fprintf(v.logDetail, " [red]Error: %v[-]\n", err)
				})
				return
			}
			v.tviewApp.QueueUpdateDraw(func() {
				v.refreshDetail(entries)
			})
		}()
	}
}

func (v *LogView) showDetail() {
	v.showingDetail = true
	v.contentFlex.Clear()
	v.contentFlex.
		AddItem(v.logTable, 30, 0, true).
		AddItem(v.logDetail, 0, 1, false)
}

func (v *LogView) closeDetail() {
	v.showingDetail = false
	v.selectedLog = ""
	v.logDetail.Clear()
	v.logDetail.SetTitle("")

	v.contentFlex.Clear()
	v.contentFlex.AddItem(v.logTable, 0, 1, true)

	if v.tviewApp != nil {
		v.tviewApp.SetFocus(v.logTable)
	}
}
