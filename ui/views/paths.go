package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// PathsView displays path statistics
type PathsView struct {
	table          *tview.Table
	onStatusChange func(msg string, isError bool)
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
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		table.SetCell(0, i, cell)
	}

	v := &PathsView{
		table: table,
	}

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
			SetTextColor(theme.FgPrimary))

		// Files
		v.table.SetCell(row, 1, tview.NewTableCell(ui.FormatNumber(ps.FileCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Dirs
		v.table.SetCell(row, 2, tview.NewTableCell(ui.FormatNumber(ps.DirCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Duration
		v.table.SetCell(row, 3, tview.NewTableCell(ui.FormatDuration(ps.ScanDurationMs)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Status with color coding
		color := theme.StatusColor(ps.Status)
		v.table.SetCell(row, 4, tview.NewTableCell(ps.Status).
			SetTextColor(color))
	}

	return nil
}

// Focus sets focus on the table
func (v *PathsView) Focus() {
	// Nothing special needed for table focus
}

// SetStatusCallback sets the callback for status messages
func (v *PathsView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
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
