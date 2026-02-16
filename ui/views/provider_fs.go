package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// FSDetailProvider implements DetailProvider for filesystem monitoring
type FSDetailProvider struct {
	data       []*models.FilesystemUsage
	summaryBox *tview.TextView // Summary tab content
	usageTable *tview.Table    // Usage tab content
}

// NewFSDetailProvider creates a new filesystem detail provider
func NewFSDetailProvider() *FSDetailProvider {
	// Create summary TextView
	summaryBox := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	summaryBox.SetBorder(true).SetTitle("Filesystem Summary")

	// Create usage table (reusing fs.go table setup logic)
	usageTable := tview.NewTable().
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
		usageTable.SetCell(0, i, cell)
	}

	return &FSDetailProvider{
		summaryBox: summaryBox,
		usageTable: usageTable,
	}
}

// Tabs returns the list of tab names
func (p *FSDetailProvider) Tabs() []string {
	return []string{"Summary", "Usage"}
}

// TabContent returns the tview Primitive for the given tab index
func (p *FSDetailProvider) TabContent(tabIndex int) tview.Primitive {
	switch tabIndex {
	case 0:
		return p.summaryBox
	case 1:
		return p.usageTable
	default:
		return nil
	}
}

// Refresh fetches fresh data from the API for all tabs
func (p *FSDetailProvider) Refresh(ctx context.Context, client ui.APIClient) error {
	usage, err := client.GetFilesystemUsage(ctx)
	if err != nil {
		return err
	}

	p.data = usage
	p.updateSummaryTab()
	p.updateUsageTab()

	return nil
}

// OnSelect is called when this category is selected
func (p *FSDetailProvider) OnSelect(activeTabIndex int) {
	// No special action needed
}

// updateSummaryTab populates the summary text view with aggregate statistics
func (p *FSDetailProvider) updateSummaryTab() {
	if len(p.data) == 0 {
		p.summaryBox.SetText("No filesystem data available")
		return
	}

	// Calculate totals
	var totalCapacity, totalUsed, totalAvail uint64
	for _, fs := range p.data {
		totalCapacity += fs.TotalBytes
		totalUsed += fs.UsedBytes
		totalAvail += fs.AvailBytes
	}

	overallPercent := 0.0
	if totalCapacity > 0 {
		overallPercent = float64(totalUsed) / float64(totalCapacity) * 100.0
	}

	// Build summary text
	summary := fmt.Sprintf("%s%sFilesystem Summary%s\n\n", theme.TagBold, theme.TagAccent, theme.TagReset)
	summary += fmt.Sprintf("%sTotal Mounts:%s %d\n\n", theme.TagLabel, theme.TagReset, len(p.data))
	summary += fmt.Sprintf("%sTotal Capacity:%s %s\n", theme.TagLabel, theme.TagReset, ui.FormatBytes(totalCapacity))
	summary += fmt.Sprintf("%sTotal Used:%s     %s\n", theme.TagLabel, theme.TagReset, ui.FormatBytes(totalUsed))
	summary += fmt.Sprintf("%sTotal Available:%s %s\n", theme.TagLabel, theme.TagReset, ui.FormatBytes(totalAvail))
	summary += fmt.Sprintf("%sOverall Usage:%s   %.1f%%\n", theme.TagLabel, theme.TagReset, overallPercent)

	p.summaryBox.SetText(summary)
}

// updateUsageTab populates the usage table (reusing fs.go rendering logic)
func (p *FSDetailProvider) updateUsageTab() {
	// Clear existing rows (keep header)
	for i := p.usageTable.GetRowCount() - 1; i > 0; i-- {
		p.usageTable.RemoveRow(i)
	}

	// Populate rows
	for i, fs := range p.data {
		row := i + 1

		// Mount point
		p.usageTable.SetCell(row, 0, tview.NewTableCell(fs.MountPoint).
			SetTextColor(theme.FgPrimary))

		// Total
		p.usageTable.SetCell(row, 1, tview.NewTableCell(ui.FormatBytes(fs.TotalBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Used
		p.usageTable.SetCell(row, 2, tview.NewTableCell(ui.FormatBytes(fs.UsedBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Available
		p.usageTable.SetCell(row, 3, tview.NewTableCell(ui.FormatBytes(fs.AvailBytes)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Use% with color coding
		color := theme.GaugeColor(fs.UsedPercent)
		p.usageTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%.1f%%", fs.UsedPercent)).
			SetTextColor(color).
			SetAlign(tview.AlignRight))

		// Usage gauge
		p.usageTable.SetCell(row, 5, tview.NewTableCell(ui.FormatGauge(fs.UsedPercent, 25)).
			SetTextColor(theme.GaugeColor(fs.UsedPercent)).
			SetExpansion(1))
	}
}
