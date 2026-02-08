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

// ProcessAPIClient defines the interface for process API operations
type ProcessAPIClient interface {
	GetProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error)
}

// ProcessView displays process monitoring information
type ProcessView struct {
	table          *tview.Table
	onStatusChange func(msg string, isError bool)
}

// NewProcessView creates a new process view
func NewProcessView() *ProcessView {
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

	return &ProcessView{table: table}
}

// Name returns the view name
func (v *ProcessView) Name() string {
	return "process"
}

// Primitive returns the tview primitive
func (v *ProcessView) Primitive() tview.Primitive {
	return v.table
}

// Refresh updates the view with fresh data
func (v *ProcessView) Refresh(ctx context.Context, c *client.Client) error {
	return v.refresh(ctx, c)
}

// refresh is the internal method that accepts ProcessAPIClient interface
func (v *ProcessView) refresh(ctx context.Context, c ProcessAPIClient) error {
	procs, err := c.GetProcessInfo(ctx)
	if err != nil {
		return err
	}

	// Clear existing rows (keep header)
	for i := v.table.GetRowCount() - 1; i > 0; i-- {
		v.table.RemoveRow(i)
	}

	for i, p := range procs {
		row := i + 1

		// PID
		v.table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", p.PID)).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// User
		v.table.SetCell(row, 1, tview.NewTableCell(p.User).
			SetTextColor(theme.FgSecondary))

		// CPU%
		cpuColor := theme.FgPrimary
		if p.CPUPercent > 80 {
			cpuColor = theme.StatusCritical
		} else if p.CPUPercent > 50 {
			cpuColor = theme.StatusWarning
		}
		v.table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f", p.CPUPercent)).
			SetTextColor(cpuColor).
			SetAlign(tview.AlignRight))

		// Memory (RSS)
		v.table.SetCell(row, 3, tview.NewTableCell(ui.FormatBytes(uint64(p.MemRSS))).
			SetTextColor(theme.FgPrimary).
			SetAlign(tview.AlignRight))

		// Status with color
		statusColor := theme.FgSecondary
		switch p.Status {
		case "running":
			statusColor = theme.StatusOK
		case "zombie":
			statusColor = theme.StatusCritical
		case "stopped":
			statusColor = theme.StatusWarning
		}
		v.table.SetCell(row, 4, tview.NewTableCell(p.Status).
			SetTextColor(statusColor))

		// Elapsed
		v.table.SetCell(row, 5, tview.NewTableCell(p.Elapsed).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))

		// Name
		v.table.SetCell(row, 6, tview.NewTableCell(p.Name).
			SetTextColor(theme.FgPrimary).
			SetExpansion(1))
	}

	return nil
}

// Focus sets focus on the table
func (v *ProcessView) Focus() {}

// SetStatusCallback sets the callback for status messages
func (v *ProcessView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
}
