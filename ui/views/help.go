package views

import (
	"context"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpView displays keybindings and help information
type HelpView struct {
	textView *tview.TextView
	onClose  func()
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetText(`[aqua::b]etlmon TUI - Keyboard Shortcuts[-::-]

[teal::b]Navigation:[-::-]
  [aqua]0[-]       Switch to Overview
  [aqua]1[-]       Switch to Filesystem view
  [aqua]2[-]       Switch to Paths view
  [aqua]?/h[-]     Show this help

[teal::b]Actions:[-::-]
  [aqua]r[-]       Refresh current view
  [aqua]s[-]       Trigger path scan (Paths view)

[teal::b]General:[-::-]
  [aqua]q[-]       Quit application
  [aqua]Ctrl+C[-]  Force quit

[silver]Press any key to return...[-]`)

	textView.SetBorder(true).
		SetTitle(" Help ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(theme.FgLabel)

	view := &HelpView{
		textView: textView,
	}

	// Set input capture to close help on any key press
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if view.onClose != nil {
			view.onClose()
		}
		return nil
	})

	return view
}

// SetCloseHandler sets the function to call when closing help
func (v *HelpView) SetCloseHandler(handler func()) {
	v.onClose = handler
}

// Name returns the view name
func (v *HelpView) Name() string {
	return "help"
}

// Primitive returns the tview primitive
func (v *HelpView) Primitive() tview.Primitive {
	return v.textView
}

// Refresh is a no-op for help view
func (v *HelpView) Refresh(ctx context.Context, client *client.Client) error {
	return nil
}

// Focus sets focus on the modal
func (v *HelpView) Focus() {
	// Nothing special needed for modal focus
}
