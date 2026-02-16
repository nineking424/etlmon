package views

import (
	"fmt"
	"strings"

	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TabBar is a horizontal tab bar widget
type TabBar struct {
	view         *tview.TextView
	tabs         []string
	activeTab    int
	onChanged    func(int)
	renderedText string // cached rendered text for testing
}

// NewTabBar creates a new TabBar
func NewTabBar() *TabBar {
	tb := &TabBar{
		view:      tview.NewTextView(),
		tabs:      []string{},
		activeTab: 0,
	}

	tb.view.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Set up input capture for [ and ] keys
	tb.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '[':
			tb.PrevTab()
			return nil
		case ']':
			tb.NextTab()
			return nil
		}
		return event
	})

	tb.render()
	return tb
}

// SetTabs sets the tab names
func (tb *TabBar) SetTabs(names []string) {
	tb.tabs = names
	if tb.activeTab >= len(names) {
		tb.activeTab = 0
	}
	tb.render()
}

// SetActiveTab sets the active tab by index
func (tb *TabBar) SetActiveTab(index int) {
	if index >= 0 && index < len(tb.tabs) {
		tb.activeTab = index
		tb.render()
		if tb.onChanged != nil {
			tb.onChanged(index)
		}
	}
}

// GetActiveTab returns the current active tab index
func (tb *TabBar) GetActiveTab() int {
	return tb.activeTab
}

// NextTab moves to the next tab (with wrap-around)
func (tb *TabBar) NextTab() {
	if len(tb.tabs) == 0 {
		return
	}
	tb.activeTab = (tb.activeTab + 1) % len(tb.tabs)
	tb.render()
	if tb.onChanged != nil {
		tb.onChanged(tb.activeTab)
	}
}

// PrevTab moves to the previous tab (with wrap-around)
func (tb *TabBar) PrevTab() {
	if len(tb.tabs) == 0 {
		return
	}
	tb.activeTab = (tb.activeTab - 1 + len(tb.tabs)) % len(tb.tabs)
	tb.render()
	if tb.onChanged != nil {
		tb.onChanged(tb.activeTab)
	}
}

// SetChangedFunc sets the callback for tab changes
func (tb *TabBar) SetChangedFunc(fn func(int)) {
	tb.onChanged = fn
}

// Primitive returns the underlying tview primitive
func (tb *TabBar) Primitive() *tview.TextView {
	return tb.view
}

// GetRenderedText returns the cached rendered text (for testing)
func (tb *TabBar) GetRenderedText() string {
	return tb.renderedText
}

// render updates the tab bar display
func (tb *TabBar) render() {
	if len(tb.tabs) == 0 {
		tb.renderedText = ""
		tb.view.SetText("")
		return
	}

	var parts []string
	for i, name := range tb.tabs {
		if i == tb.activeTab {
			// Active tab: accent color + bold
			parts = append(parts, fmt.Sprintf("%s%s[%s]%s", theme.TagAccent, theme.TagBold, name, theme.TagReset))
		} else {
			// Inactive tab: secondary color
			parts = append(parts, fmt.Sprintf("%s[%s]%s", theme.TagSecondary, name, theme.TagReset))
		}
	}

	text := " " + strings.Join(parts, " ") + " "
	tb.renderedText = text
	tb.view.SetText(text)
}
