package views

import (
	"context"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/rivo/tview"
)

// HelpView displays keybindings and help information
type HelpView struct {
	modal *tview.Modal
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	modal := tview.NewModal().
		SetText(`etlmon - ETL Pipeline Monitor

Keybindings:
  F / f       - Switch to Filesystem view
  P / p       - Switch to Paths view
  R / r       - Refresh current view
  S / s       - Trigger scan (Paths view only)
  Q / q       - Quit application
  ? / h       - Show this help

Navigation:
  Arrow keys  - Move selection
  Tab         - Cycle through views
  Enter       - Select item (context-dependent)

Views:
  Filesystem  - Shows disk usage for mounted filesystems
  Paths       - Shows file/directory counts for monitored paths

For more information, visit:
  https://github.com/etlmon/etlmon`).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// This will be handled by the app controller
		})

	return &HelpView{
		modal: modal,
	}
}

// Name returns the view name
func (v *HelpView) Name() string {
	return "help"
}

// Primitive returns the tview primitive
func (v *HelpView) Primitive() tview.Primitive {
	return v.modal
}

// Refresh is a no-op for help view
func (v *HelpView) Refresh(ctx context.Context, client *client.Client) error {
	return nil
}

// Focus sets focus on the modal
func (v *HelpView) Focus() {
	// Nothing special needed for modal focus
}
