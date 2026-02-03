package views

import (
	"context"

	"github.com/etlmon/etlmon/ui/client"
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
		SetText(`[yellow::b]etlmon TUI - Keyboard Shortcuts[-::-]

[green::b]Navigation:[-::-]
  1       Switch to Filesystem view
  2       Switch to Paths view
  ?/h     Show this help

[green::b]Actions:[-::-]
  r       Refresh current view
  s       Trigger path scan (Paths view)
  T       Toggle table borders

[green::b]General:[-::-]
  q       Quit application
  Ctrl+C  Force quit

[yellow]Press any key to return...[-]`)

	textView.SetBorder(true).
		SetTitle(" Help ").
		SetTitleAlign(tview.AlignCenter)

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
