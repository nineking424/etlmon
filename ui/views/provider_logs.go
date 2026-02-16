package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// LogsDetailProvider implements DetailProvider for log file monitoring
type LogsDetailProvider struct {
	logFiles    []models.LogFileInfo
	logEntries  []*models.LogEntry
	selectedLog string
	filesTable  *tview.Table    // Files tab: log file list
	viewer      *tview.TextView // Viewer tab: log content
	apiClient   ui.APIClient    // for GetLogEntriesByName
	tviewApp    *tview.Application
}

// NewLogsDetailProvider creates a new logs detail provider
func NewLogsDetailProvider(client ui.APIClient, app *tview.Application) *LogsDetailProvider {
	// Create files table
	filesTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	// Table headers
	headers := []string{"Name", "Path", "Size", "Modified"}
	aligns := []int{
		tview.AlignLeft,  // Name
		tview.AlignLeft,  // Path
		tview.AlignRight, // Size
		tview.AlignRight, // Modified
	}

	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.TableHeader).
			SetAttributes(theme.TableHeaderAttr).
			SetAlign(aligns[i]).
			SetSelectable(false)
		if i == 1 { // Path column expands
			cell.SetExpansion(1)
		}
		filesTable.SetCell(0, i, cell)
	}

	// Create viewer
	viewer := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)
	viewer.SetText("Select a log file from the Files tab")

	return &LogsDetailProvider{
		filesTable: filesTable,
		viewer:     viewer,
		apiClient:  client,
		tviewApp:   app,
	}
}

// Tabs returns the list of tab names
func (p *LogsDetailProvider) Tabs() []string {
	return []string{"Files", "Viewer"}
}

// TabContent returns the tview Primitive for the given tab index
func (p *LogsDetailProvider) TabContent(tabIndex int) tview.Primitive {
	switch tabIndex {
	case 0:
		return p.filesTable
	case 1:
		return p.viewer
	default:
		return nil
	}
}

// Refresh fetches fresh data from the API for the file list
func (p *LogsDetailProvider) Refresh(ctx context.Context, client ui.APIClient) error {
	p.apiClient = client

	files, err := client.GetLogFiles(ctx)
	if err != nil {
		return err
	}

	p.logFiles = files
	p.populateFilesTable()

	return nil
}

// OnSelect is called when this category is selected
func (p *LogsDetailProvider) OnSelect(activeTabIndex int) {
	// No special action needed
}

// populateFilesTable fills the Files tab with log file list
func (p *LogsDetailProvider) populateFilesTable() {
	// Clear existing rows (keep header)
	for i := p.filesTable.GetRowCount() - 1; i > 0; i-- {
		p.filesTable.RemoveRow(i)
	}

	if len(p.logFiles) == 0 {
		p.filesTable.SetCell(1, 0, tview.NewTableCell("(no log files configured)").
			SetTextColor(theme.FgMuted).
			SetExpansion(1))
		return
	}

	for i, file := range p.logFiles {
		row := i + 1

		// Name
		p.filesTable.SetCell(row, 0, tview.NewTableCell(file.Name).
			SetTextColor(theme.FgPrimary))

		// Path
		p.filesTable.SetCell(row, 1, tview.NewTableCell(file.Path).
			SetTextColor(theme.FgSecondary).
			SetExpansion(1))

		// Size
		p.filesTable.SetCell(row, 2, tview.NewTableCell(formatFileSize(file.Size)).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))

		// Modified
		modTimeStr := "-"
		if !file.ModTime.IsZero() {
			modTimeStr = file.ModTime.Format("2006-01-02 15:04:05")
		}
		p.filesTable.SetCell(row, 3, tview.NewTableCell(modTimeStr).
			SetTextColor(theme.FgSecondary).
			SetAlign(tview.AlignRight))
	}
}

// loadLogContent loads log entries for the specified log file name
// This is a synchronous method for testability; callers should wrap in goroutine if needed
func (p *LogsDetailProvider) loadLogContent(name string) error {
	if p.apiClient == nil {
		return fmt.Errorf("apiClient is nil")
	}

	ctx := context.Background()
	entries, err := p.apiClient.GetLogEntriesByName(ctx, name, 500)
	if err != nil {
		return err
	}

	p.logEntries = entries
	p.selectedLog = name
	p.populateViewer()

	return nil
}

// populateViewer fills the Viewer tab with log entries
func (p *LogsDetailProvider) populateViewer() {
	p.viewer.Clear()

	if len(p.logEntries) == 0 {
		fmt.Fprintf(p.viewer, " [darkgray]No log entries[-]\n")
		return
	}

	for _, entry := range p.logEntries {
		timestamp := entry.CreatedAt.Format("15:04:05")
		fmt.Fprintf(p.viewer, "[teal]%s[-] %s\n", timestamp, entry.Line)
	}

	p.viewer.ScrollToEnd()
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
