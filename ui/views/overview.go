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

// OverviewView displays a btop-style combined dashboard
type OverviewView struct {
	flex    *tview.Flex
	fsBox   *tview.Table
	pathBox *tview.Table
}

// NewOverviewView creates a new overview dashboard
func NewOverviewView() *OverviewView {
	fsBox := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false).
		SetFixed(1, 0)
	fsBox.SetBorder(true).
		SetTitle(" Filesystem Usage ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	// Set FS table headers
	fsHeaders := []string{"Mount", "Usage", "Used", "Total"}
	for i, header := range fsHeaders {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetSelectable(false)
		if i == 0 {
			cell.SetAlign(tview.AlignLeft)
		} else if i == 1 {
			cell.SetAlign(tview.AlignLeft).SetExpansion(1)
		} else {
			cell.SetAlign(tview.AlignRight)
		}
		fsBox.SetCell(0, i, cell)
	}

	pathBox := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false).
		SetFixed(1, 0)
	pathBox.SetBorder(true).
		SetTitle(" Path Statistics ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	// Set path table headers
	pathHeaders := []string{"Path", "Files", "Dirs", "Duration", "Status"}
	for i, header := range pathHeaders {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		pathBox.SetCell(0, i, cell)
	}

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(fsBox, 0, 1, false).
		AddItem(pathBox, 0, 2, true)

	return &OverviewView{
		flex:    flex,
		fsBox:   fsBox,
		pathBox: pathBox,
	}
}

// Name returns the view name
func (v *OverviewView) Name() string {
	return "overview"
}

// Primitive returns the tview primitive
func (v *OverviewView) Primitive() tview.Primitive {
	return v.flex
}

// Refresh updates the view with fresh data
func (v *OverviewView) Refresh(ctx context.Context, c *client.Client) error {
	return v.refresh(ctx, c)
}

// refresh is the internal method that accepts APIClient interface
func (v *OverviewView) refresh(ctx context.Context, c APIClient) error {
	var fsErr, pathErr error

	// Fetch filesystem usage
	usage, err := c.GetFilesystemUsage(ctx)
	if err != nil {
		fsErr = err
	}

	// Fetch path stats
	stats, err := c.GetPathStats(ctx)
	if err != nil {
		pathErr = err
	}

	// Render FS box
	v.renderFS(usage, fsErr)

	// Render Path box
	v.renderPaths(stats, pathErr)

	// Return first error if both failed
	if fsErr != nil && pathErr != nil {
		return fmt.Errorf("fs: %v, paths: %v", fsErr, pathErr)
	}

	return nil
}

// renderFS renders filesystem usage as a table with aligned columns
func (v *OverviewView) renderFS(usage []*models.FilesystemUsage, err error) {
	// Clear existing data rows (keep header)
	for i := v.fsBox.GetRowCount() - 1; i > 0; i-- {
		v.fsBox.RemoveRow(i)
	}

	if err != nil {
		v.fsBox.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %s", err.Error())).
			SetTextColor(theme.StatusCritical))
		return
	}

	if len(usage) == 0 {
		v.fsBox.SetCell(1, 0, tview.NewTableCell("No filesystem data").
			SetTextColor(theme.FgMuted))
		return
	}

	for i, fs := range usage {
		row := i + 1
		color := theme.GaugeColor(fs.UsedPercent)

		// Mount point
		v.fsBox.SetCell(row, 0, tview.NewTableCell(fs.MountPoint).
			SetTextColor(theme.FgPrimary))

		// Gauge bar
		v.fsBox.SetCell(row, 1, tview.NewTableCell(ui.FormatGauge(fs.UsedPercent, 30)).
			SetTextColor(color).
			SetExpansion(1))

		// Used bytes
		v.fsBox.SetCell(row, 2, tview.NewTableCell(ui.FormatBytes(fs.UsedBytes)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))

		// Total bytes
		v.fsBox.SetCell(row, 3, tview.NewTableCell(ui.FormatBytes(fs.TotalBytes)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
	}
}

// renderPaths renders path statistics table
func (v *OverviewView) renderPaths(stats []*models.PathStats, err error) {
	// Clear existing data rows (keep header)
	for i := v.pathBox.GetRowCount() - 1; i > 0; i-- {
		v.pathBox.RemoveRow(i)
	}

	if err != nil {
		v.pathBox.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %s", err.Error())).
			SetTextColor(theme.StatusCritical))
		return
	}

	if len(stats) == 0 {
		v.pathBox.SetCell(1, 0, tview.NewTableCell("No path data").
			SetTextColor(theme.FgMuted))
		return
	}

	for i, ps := range stats {
		row := i + 1

		v.pathBox.SetCell(row, 0, tview.NewTableCell(ps.Path).
			SetTextColor(theme.FgPrimary))

		v.pathBox.SetCell(row, 1, tview.NewTableCell(ui.FormatNumber(ps.FileCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		v.pathBox.SetCell(row, 2, tview.NewTableCell(ui.FormatNumber(ps.DirCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		v.pathBox.SetCell(row, 3, tview.NewTableCell(ui.FormatDuration(ps.ScanDurationMs)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		v.pathBox.SetCell(row, 4, tview.NewTableCell(ps.Status).
			SetTextColor(theme.StatusColor(ps.Status)))
	}
}

// Focus sets focus on the path table
func (v *OverviewView) Focus() {
	// Nothing special needed
}
