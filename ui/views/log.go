package views

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// LogAPIClient defines the interface for log API operations
type LogAPIClient interface {
	GetLogEntries(ctx context.Context) ([]*models.LogEntry, error)
}

// LogView displays log entries in a scrollable text view
type LogView struct {
	flex           *tview.Flex
	logText        *tview.TextView
	onStatusChange func(msg string, isError bool)
}

// NewLogView creates a new log view
func NewLogView() *LogView {
	logText := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)
	logText.SetBorder(true).
		SetTitle(" Log Entries ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logText, 0, 1, true)

	return &LogView{
		flex:    flex,
		logText: logText,
	}
}

// Name returns the view name
func (v *LogView) Name() string {
	return "logs"
}

// Primitive returns the tview primitive
func (v *LogView) Primitive() tview.Primitive {
	return v.flex
}

// Refresh updates the view with fresh data
func (v *LogView) Refresh(ctx context.Context, c *client.Client) error {
	return v.refresh(ctx, c)
}

// refresh is the internal method that accepts LogAPIClient interface
func (v *LogView) refresh(ctx context.Context, c LogAPIClient) error {
	entries, err := c.GetLogEntries(ctx)
	if err != nil {
		return err
	}

	v.logText.Clear()

	if len(entries) == 0 {
		fmt.Fprintf(v.logText, " [darkgray]No log entries[-]\n")
		return nil
	}

	for _, entry := range entries {
		timestamp := entry.CreatedAt.Format("15:04:05")
		fmt.Fprintf(v.logText, "[teal]%s[-] [aqua]%-12s[-] %s\n",
			timestamp,
			entry.LogName,
			entry.Line,
		)
	}

	// Scroll to bottom (newest entries)
	v.logText.ScrollToEnd()

	return nil
}

// Focus sets focus on the text view
func (v *LogView) Focus() {}

// SetStatusCallback sets the callback for status messages
func (v *LogView) SetStatusCallback(cb func(msg string, isError bool)) {
	v.onStatusChange = cb
}
