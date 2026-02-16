package views

import (
	"context"
	"fmt"
	"sort"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// ProcessDetailProvider implements DetailProvider for process monitoring
type ProcessDetailProvider struct {
	data        []*models.ProcessInfo
	listTable   *tview.Table // List tab: all processes
	topCPUTable *tview.Table // Top CPU tab: sorted by CPU%
	topMemTable *tview.Table // Top Memory tab: sorted by Memory
}

// NewProcessDetailProvider creates a new process detail provider
func NewProcessDetailProvider() *ProcessDetailProvider {
	p := &ProcessDetailProvider{
		listTable:   createProcessTable(),
		topCPUTable: createProcessTable(),
		topMemTable: createProcessTable(),
	}
	return p
}

// createProcessTable creates a process table with headers
func createProcessTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Set headers
	headers := []string{"PID", "User", "CPU%", "Memory", "Status", "Elapsed", "Name"}
	aligns := []int{
		tview.AlignRight, // PID
		tview.AlignLeft,  // User
		tview.AlignRight, // CPU%
		tview.AlignRight, // Memory
		tview.AlignLeft,  // Status
		tview.AlignRight, // Elapsed
		tview.AlignLeft,  // Name
	}

	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetAlign(aligns[i]).
			SetSelectable(false)
		if i == 6 { // Name column expands
			cell.SetExpansion(1)
		}
		table.SetCell(0, i, cell)
	}

	return table
}

// Tabs returns the list of tab names
func (p *ProcessDetailProvider) Tabs() []string {
	return []string{"List", "Top CPU", "Top Memory"}
}

// TabContent returns the tview Primitive for the given tab index
func (p *ProcessDetailProvider) TabContent(tabIndex int) tview.Primitive {
	switch tabIndex {
	case 0:
		return p.listTable
	case 1:
		return p.topCPUTable
	case 2:
		return p.topMemTable
	default:
		return nil
	}
}

// Refresh fetches fresh data from the API and populates all tabs
func (p *ProcessDetailProvider) Refresh(ctx context.Context, client ui.APIClient) error {
	procs, err := client.GetProcessInfo(ctx)
	if err != nil {
		return err
	}

	p.data = procs
	p.populateListTable()
	p.populateTopCPUTable()
	p.populateTopMemTable()

	return nil
}

// OnSelect is called when this category is selected
func (p *ProcessDetailProvider) OnSelect(activeTabIndex int) {
	// No special action needed
}

// populateListTable fills the List tab with all processes
func (p *ProcessDetailProvider) populateListTable() {
	// Clear existing rows (keep header)
	for i := p.listTable.GetRowCount() - 1; i > 0; i-- {
		p.listTable.RemoveRow(i)
	}

	for i, proc := range p.data {
		row := i + 1
		p.addProcessRow(p.listTable, row, proc)
	}
}

// populateTopCPUTable fills the Top CPU tab with processes sorted by CPU% descending
func (p *ProcessDetailProvider) populateTopCPUTable() {
	// Clear existing rows (keep header)
	for i := p.topCPUTable.GetRowCount() - 1; i > 0; i-- {
		p.topCPUTable.RemoveRow(i)
	}

	// Sort by CPU% descending
	sorted := make([]*models.ProcessInfo, len(p.data))
	copy(sorted, p.data)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CPUPercent > sorted[j].CPUPercent
	})

	// Take top 10
	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}

	for i := 0; i < limit; i++ {
		row := i + 1
		p.addProcessRow(p.topCPUTable, row, sorted[i])
	}
}

// populateTopMemTable fills the Top Memory tab with processes sorted by Memory descending
func (p *ProcessDetailProvider) populateTopMemTable() {
	// Clear existing rows (keep header)
	for i := p.topMemTable.GetRowCount() - 1; i > 0; i-- {
		p.topMemTable.RemoveRow(i)
	}

	// Sort by Memory descending
	sorted := make([]*models.ProcessInfo, len(p.data))
	copy(sorted, p.data)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MemRSS > sorted[j].MemRSS
	})

	// Take top 10
	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}

	for i := 0; i < limit; i++ {
		row := i + 1
		p.addProcessRow(p.topMemTable, row, sorted[i])
	}
}

// addProcessRow adds a process row to the table with color coding
func (p *ProcessDetailProvider) addProcessRow(table *tview.Table, row int, proc *models.ProcessInfo) {
	// PID
	table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", proc.PID)).
		SetTextColor(theme.FgPrimary).
		SetAlign(tview.AlignRight))

	// User
	table.SetCell(row, 1, tview.NewTableCell(proc.User).
		SetTextColor(theme.FgSecondary))

	// CPU% with color coding
	cpuColor := theme.FgPrimary
	if proc.CPUPercent > 80 {
		cpuColor = theme.StatusCritical
	} else if proc.CPUPercent > 50 {
		cpuColor = theme.StatusWarning
	}
	table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f", proc.CPUPercent)).
		SetTextColor(cpuColor).
		SetAlign(tview.AlignRight))

	// Memory (RSS)
	table.SetCell(row, 3, tview.NewTableCell(ui.FormatBytes(uint64(proc.MemRSS))).
		SetTextColor(theme.FgPrimary).
		SetAlign(tview.AlignRight))

	// Status with color coding
	statusColor := theme.FgSecondary
	switch proc.Status {
	case "running":
		statusColor = theme.StatusOK
	case "zombie":
		statusColor = theme.StatusCritical
	case "stopped":
		statusColor = theme.StatusWarning
	}
	table.SetCell(row, 4, tview.NewTableCell(proc.Status).
		SetTextColor(statusColor))

	// Elapsed
	table.SetCell(row, 5, tview.NewTableCell(proc.Elapsed).
		SetTextColor(theme.FgSecondary).
		SetAlign(tview.AlignRight))

	// Name
	table.SetCell(row, 6, tview.NewTableCell(proc.Name).
		SetTextColor(theme.FgPrimary).
		SetExpansion(1))
}
