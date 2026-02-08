package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// APIClient defines the interface for API operations
type APIClient interface {
	GetFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error)
	GetPathStats(ctx context.Context) ([]*models.PathStats, error)
	TriggerScan(ctx context.Context, paths []string) error
}

// FSView displays filesystem usage statistics
type FSView struct {
	table          *tview.Table
	onStatusChange func(msg string, isError bool)
}

// NewFSView creates a new filesystem view
func NewFSView() *FSView {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Set headers
	headers := []string{"Mount", "Total", "Used", "Avail", "Use%", "Usage"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		if i == 5 {
			cell.SetExpansion(1)
		}
		table.SetCell(0, i, cell)
	}

	v := &FSView{
		table: table,
	}

	return v
}

// Name returns the view name
func (v *FSView) Name() string {
	return "fs"
}

// Primitive returns the tview primitive
func (v *FSView) Primitive() tview.Primitive {
	return v.table
}

// Refresh updates the view with fresh data
func (v *FSView) Refresh(ctx context.Context, client *client.Client) error {
	return v.refresh(ctx, client)
}

// refresh is the internal method that accepts APIClient interface
func (v *FSView) refresh(ctx context.Context, client APIClient) error {
	usage, err := client.GetFilesystemUsage(ctx)
	if err != nil {
		return err
	}

	// Clear existing rows (keep header)
	for i := v.table.GetRowCount() - 1; i > 0; i-- {
		v.table.RemoveRow(i)
	}

	// Populate rows
	for i, fs := range usage {
		row := i + 1

		// Mount point
		v.table.SetCell(row, 0, tview.NewTableCell(fs.MountPoint).
			SetTextColor(theme.FgPrimary))

		// Total
		v.table.SetCell(row, 1, tview.NewTableCell(ui.FormatBytes(fs.TotalBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Used
		v.table.SetCell(row, 2, tview.NewTableCell(ui.FormatBytes(fs.UsedBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Available
		v.table.SetCell(row, 3, tview.NewTableCell(ui.FormatBytes(fs.AvailBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Use% with color coding
		color := theme.GaugeColor(fs.UsedPercent)
		v.table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%.1f%%", fs.UsedPercent)).
			SetTextColor(color).
			SetAlign(tview.AlignRight))

		// Usage gauge
		v.table.SetCell(row, 5, tview.NewTableCell(ui.FormatGauge(fs.UsedPercent, 25)).
			SetTextColor(theme.GaugeColor(fs.UsedPercent)).
			SetExpansion(1))
	}

	return nil
}

// Focus sets focus on the table
func (v *FSView) Focus() {
	// Nothing special needed for table focus
}

// SetStatusCallback sets the callback for status messages
func (v *FSView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
}
