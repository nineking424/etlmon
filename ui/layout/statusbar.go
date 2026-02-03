package layout

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar displays view info, timestamp, and status messages
type StatusBar struct {
	flex      *tview.Flex
	viewInfo  *tview.TextView
	timestamp *tview.TextView
	message   *tview.TextView
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	s := &StatusBar{
		flex:      tview.NewFlex(),
		viewInfo:  tview.NewTextView().SetDynamicColors(true),
		timestamp: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter),
		message:   tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
	}

	// Set background color for status bar
	s.viewInfo.SetBackgroundColor(tcell.ColorDarkGreen)
	s.timestamp.SetBackgroundColor(tcell.ColorDarkGreen)
	s.message.SetBackgroundColor(tcell.ColorDarkGreen)

	// Horizontal layout
	s.flex.SetDirection(tview.FlexColumn).
		AddItem(s.viewInfo, 20, 0, false).
		AddItem(s.timestamp, 0, 1, false).
		AddItem(s.message, 30, 0, false)

	// Set initial values
	s.SetView("--")
	s.SetMessage("Ready", false)

	return s
}

// SetView updates the current view name
func (s *StatusBar) SetView(name string) {
	s.viewInfo.SetText(fmt.Sprintf(" [white::b]View:[white:-] %s ", name))
}

// SetLastRefresh updates the last refresh timestamp
func (s *StatusBar) SetLastRefresh(t time.Time) {
	s.timestamp.SetText(fmt.Sprintf("[gray]│[-] [white]Last:[gray] %s [gray]│[-]", t.Format("15:04:05")))
}

// SetMessage updates the status message
func (s *StatusBar) SetMessage(msg string, isError bool) {
	color := "[green]"
	if isError {
		color = "[red::b]"
	}
	s.message.SetText(fmt.Sprintf("%s%s[-:-] ", color, msg))
}

// Primitive returns the status bar's tview primitive
func (s *StatusBar) Primitive() tview.Primitive {
	return s.flex
}
