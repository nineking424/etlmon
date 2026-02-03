package views

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui/client"
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
	table     *tview.Table
	tableMode bool
}

// NewFSView creates a new filesystem view
func NewFSView() *FSView {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Set headers
	headers := []string{"Mount", "Total", "Used", "Avail", "Use%"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		table.SetCell(0, i, cell)
	}

	v := &FSView{
		table:     table,
		tableMode: false,
	}

	// Set up input capture for table toggle
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'T':
			v.tableMode = !v.tableMode
			v.table.SetBorders(v.tableMode)
			return nil
		}
		return event
	})

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
			SetTextColor(tcell.ColorWhite))

		// Total
		v.table.SetCell(row, 1, tview.NewTableCell(formatBytes(fs.TotalBytes)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Used
		v.table.SetCell(row, 2, tview.NewTableCell(formatBytes(fs.UsedBytes)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Available
		v.table.SetCell(row, 3, tview.NewTableCell(formatBytes(fs.AvailBytes)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight))

		// Use% with color coding
		color := tcell.ColorGreen
		if fs.UsedPercent > 90 {
			color = tcell.ColorRed
		} else if fs.UsedPercent > 75 {
			color = tcell.ColorYellow
		}
		v.table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%.1f%%", fs.UsedPercent)).
			SetTextColor(color).
			SetAlign(tview.AlignRight))
	}

	return nil
}

// Focus sets focus on the table
func (v *FSView) Focus() {
	// Nothing special needed for table focus
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}
