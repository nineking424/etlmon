package views

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/rivo/tview"
)

// PathsView displays path statistics
type PathsView struct {
	table     *tview.Table
	tableMode bool // true = with borders, false = compact
}

// NewPathsView creates a new paths view
func NewPathsView() *PathsView {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Set headers
	headers := []string{"Path", "Files", "Dirs", "Duration", "Status"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		table.SetCell(0, i, cell)
	}

	v := &PathsView{
		table:     table,
		tableMode: false,
	}

	// Set up key handlers
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'T':
			// Toggle table format
			v.tableMode = !v.tableMode
			v.table.SetBorders(v.tableMode)
			return nil
		}
		return event
	})

	return v
}

// Name returns the view name
func (v *PathsView) Name() string {
	return "paths"
}

// Primitive returns the tview primitive
func (v *PathsView) Primitive() tview.Primitive {
	return v.table
}

// Refresh updates the view with fresh data
func (v *PathsView) Refresh(ctx context.Context, client *client.Client) error {
	return v.refresh(ctx, client)
}

// refresh is the internal method that accepts APIClient interface
func (v *PathsView) refresh(ctx context.Context, client APIClient) error {
	stats, err := client.GetPathStats(ctx)
	if err != nil {
		return err
	}

	// Clear existing rows (keep header)
	for i := v.table.GetRowCount() - 1; i > 0; i-- {
		v.table.RemoveRow(i)
	}

	// Populate rows
	for i, ps := range stats {
		row := i + 1

		// Path
		v.table.SetCell(row, 0, tview.NewTableCell(ps.Path).
			SetTextColor(tcell.ColorWhite))

		// Files
		v.table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", ps.FileCount)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Dirs
		v.table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", ps.DirCount)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Duration
		v.table.SetCell(row, 3, tview.NewTableCell(formatDuration(ps.ScanDurationMs)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Status with color coding
		color := tcell.ColorGreen
		if ps.Status == "ERROR" {
			color = tcell.ColorRed
		} else if ps.Status == "SCANNING" {
			color = tcell.ColorYellow
		}
		v.table.SetCell(row, 4, tview.NewTableCell(ps.Status).
			SetTextColor(color))
	}

	return nil
}

// Focus sets focus on the table
func (v *PathsView) Focus() {
	// Nothing special needed for table focus
}

// TriggerScan triggers a manual scan for all paths currently displayed
func (v *PathsView) TriggerScan(ctx context.Context, client *client.Client) error {
	// Collect all paths from the current table
	var paths []string
	rowCount := v.table.GetRowCount()
	for i := 1; i < rowCount; i++ { // Skip header row
		if cell := v.table.GetCell(i, 0); cell != nil {
			paths = append(paths, cell.Text)
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("no paths to scan")
	}

	return client.TriggerScan(ctx, paths)
}

// formatDuration formats milliseconds into human-readable string
func formatDuration(ms int64) string {
	if ms == 0 {
		return "0ms"
	}

	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}

	seconds := ms / 1000
	remainingMs := ms % 1000

	if seconds < 60 {
		if remainingMs > 0 {
			return fmt.Sprintf("%.1fs", float64(ms)/1000.0)
		}
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
}
