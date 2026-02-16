package views

import (
	"context"

	"github.com/etlmon/etlmon/ui"
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

[teal::b]Pages:[-::-]
  [aqua]s[-]       Switch to Settings
  [aqua]?/h[-]     Show this help
  [aqua]Esc[-]     Return to Overview (from Settings)

[teal::b]Overview Navigation:[-::-]
  [aqua]1[-]       Jump to FS category
  [aqua]2[-]       Jump to Paths category
  [aqua]3[-]       Jump to Process category
  [aqua]4[-]       Jump to Logs category
  [aqua]j/k[-]     Move up/down in category list
  [aqua]Tab[-]     Toggle focus (category list ↔ detail panel)

[teal::b]Detail Panel:[-::-]
  [aqua][[silver]/[aqua]][-]     Previous/Next tab
  [aqua]j/k[-]     Navigate within detail content

[teal::b]Settings:[-::-]
  [aqua]a[-]       Add new entry
  [aqua]e[-]       Edit selected entry
  [aqua]d[-]       Delete selected entry
  [aqua]s[-]       Save settings
  [aqua]Tab[-]     Switch between sidebar and content
  [aqua]Esc[-]     Return to Overview (when not editing)

[teal::b]General:[-::-]
  [aqua]r[-]       Refresh current view
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
func (v *HelpView) Refresh(ctx context.Context, client ui.APIClient) error {
	return nil
}

// Focus sets focus on the modal
func (v *HelpView) Focus() {
	// Nothing special needed for modal focus
}
