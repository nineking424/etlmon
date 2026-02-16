package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PathsDetailProvider implements DetailProvider for path scanning monitoring
type PathsDetailProvider struct {
	data       []*models.PathStats
	statsTable *tview.Table        // Stats tab content
	scanFlex   *tview.Flex         // Scan tab content
	scanTable  *tview.Table        // Scan tab path list
	scanStatus *tview.TextView     // Scan status message
	apiClient  ui.APIClient        // needed for TriggerScan
	tviewApp   *tview.Application  // for QueueUpdateDraw
}

// NewPathsDetailProvider creates a new paths detail provider
func NewPathsDetailProvider(client ui.APIClient, app *tview.Application) *PathsDetailProvider {
	// Create stats table (reusing paths.go table setup logic)
	statsTable := tview.NewTable().
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
		statsTable.SetCell(0, i, cell)
	}

	// Create scan tab components
	scanTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Scan table headers
	scanTable.SetCell(0, 0, tview.NewTableCell("Path").
		SetTextColor(theme.TableHeader).
		SetAttributes(theme.TableHeaderAttr).
		SetSelectable(false))

	scanStatus := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	scanStatus.SetText(fmt.Sprintf("%sPress Enter to trigger scan for all paths%s", theme.TagLabel, theme.TagReset))

	scanFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(scanTable, 0, 1, true).
		AddItem(scanStatus, 3, 0, false)

	p := &PathsDetailProvider{
		statsTable: statsTable,
		scanFlex:   scanFlex,
		scanTable:  scanTable,
		scanStatus: scanStatus,
		apiClient:  client,
		tviewApp:   app,
	}

	// Set up Enter key handler for scan trigger
	scanTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			p.triggerScan()
			return nil
		}
		return event
	})

	return p
}

// Tabs returns the list of tab names
func (p *PathsDetailProvider) Tabs() []string {
	return []string{"Stats", "Scan"}
}

// TabContent returns the tview Primitive for the given tab index
func (p *PathsDetailProvider) TabContent(tabIndex int) tview.Primitive {
	switch tabIndex {
	case 0:
		return p.statsTable
	case 1:
		return p.scanFlex
	default:
		return nil
	}
}

// Refresh fetches fresh data from the API for all tabs
func (p *PathsDetailProvider) Refresh(ctx context.Context, client ui.APIClient) error {
	stats, err := client.GetPathStats(ctx)
	if err != nil {
		return err
	}

	p.data = stats
	p.updateStatsTab()
	p.updateScanTab()

	return nil
}

// OnSelect is called when this category is selected
func (p *PathsDetailProvider) OnSelect(activeTabIndex int) {
	// No special action needed
}

// updateStatsTab populates the stats table (reusing paths.go rendering logic)
func (p *PathsDetailProvider) updateStatsTab() {
	// Clear existing rows (keep header)
	for i := p.statsTable.GetRowCount() - 1; i > 0; i-- {
		p.statsTable.RemoveRow(i)
	}

	// Populate rows
	for i, ps := range p.data {
		row := i + 1

		// Path
		p.statsTable.SetCell(row, 0, tview.NewTableCell(ps.Path).
			SetTextColor(theme.FgPrimary))

		// Files
		p.statsTable.SetCell(row, 1, tview.NewTableCell(ui.FormatNumber(ps.FileCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Dirs
		p.statsTable.SetCell(row, 2, tview.NewTableCell(ui.FormatNumber(ps.DirCount)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Duration
		p.statsTable.SetCell(row, 3, tview.NewTableCell(ui.FormatDuration(ps.ScanDurationMs)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Status with color coding
		color := theme.StatusColor(ps.Status)
		p.statsTable.SetCell(row, 4, tview.NewTableCell(ps.Status).
			SetTextColor(color))
	}
}

// updateScanTab populates the scan tab with path list
func (p *PathsDetailProvider) updateScanTab() {
	// Clear existing rows (keep header)
	for i := p.scanTable.GetRowCount() - 1; i > 0; i-- {
		p.scanTable.RemoveRow(i)
	}

	// Populate path list
	for i, ps := range p.data {
		row := i + 1
		p.scanTable.SetCell(row, 0, tview.NewTableCell(ps.Path).
			SetTextColor(theme.FgPrimary))
	}

	// Reset status message
	p.scanStatus.SetText(fmt.Sprintf("%sPress Enter to trigger scan for all paths%s", theme.TagLabel, theme.TagReset))
}

// triggerScan triggers a manual scan for all paths
func (p *PathsDetailProvider) triggerScan() {
	if p.apiClient == nil {
		p.scanStatus.SetText(fmt.Sprintf("%s[red]Error: API client not available%s", theme.TagBold, theme.TagReset))
		return
	}

	// Collect all paths
	var paths []string
	for _, ps := range p.data {
		paths = append(paths, ps.Path)
	}

	if len(paths) == 0 {
		p.scanStatus.SetText(fmt.Sprintf("%s[yellow]No paths to scan%s", theme.TagBold, theme.TagReset))
		return
	}

	// Show scanning status
	p.scanStatus.SetText(fmt.Sprintf("%s[yellow]Scanning %d paths...%s", theme.TagBold, len(paths), theme.TagReset))
	if p.tviewApp != nil {
		p.tviewApp.Draw()
	}

	// Trigger scan
	ctx := context.Background()
	err := p.apiClient.TriggerScan(ctx, paths)

	// Update status
	if err != nil {
		p.scanStatus.SetText(fmt.Sprintf("%s[red]Scan failed: %v%s", theme.TagBold, err, theme.TagReset))
	} else {
		p.scanStatus.SetText(fmt.Sprintf("%s[green]Scan triggered successfully for %d paths%s", theme.TagBold, len(paths), theme.TagReset))
	}

	if p.tviewApp != nil {
		p.tviewApp.Draw()
	}
}
