package layout

import (
	"time"

	"github.com/rivo/tview"
)

// Layout manages the K9s-style layout structure
type Layout struct {
	root      *tview.Flex
	header    *Header
	navbar    *NavBar
	content   *tview.Flex // Container for content views
	statusbar *StatusBar
}

// NewLayout creates a new K9s-style layout
func NewLayout() *Layout {
	l := &Layout{
		root:      tview.NewFlex(),
		header:    NewHeader(),
		navbar:    NewNavBar(),
		content:   tview.NewFlex(),
		statusbar: NewStatusBar(),
	}

	// Vertical layout: Header | NavBar | Content | StatusBar
	l.root.SetDirection(tview.FlexRow).
		AddItem(l.header.Primitive(), 5, 0, false).   // Header: 5 lines for logo
		AddItem(l.navbar.Primitive(), 1, 0, false).   // NavBar: 1 line
		AddItem(l.content, 0, 1, true).               // Content: flexible
		AddItem(l.statusbar.Primitive(), 1, 0, false) // StatusBar: 1 line

	return l
}

// SetContent replaces the content area with a new primitive
func (l *Layout) SetContent(p tview.Primitive) {
	l.content.Clear()
	l.content.AddItem(p, 0, 1, true)
}

// SetActiveView updates navbar and statusbar for current view
func (l *Layout) SetActiveView(viewName string) {
	l.navbar.SetActive(viewName)
	l.statusbar.SetView(viewName)
	l.statusbar.SetLastRefresh(time.Now())
}

// SetContext updates header context info (node name, status)
func (l *Layout) SetContext(nodeName string, status string) {
	l.header.SetContext(nodeName, status)
}

// SetResource updates header resource info
func (l *Layout) SetResource(info string) {
	l.header.SetResource(info)
}

// SetMessage updates the statusbar message
func (l *Layout) SetMessage(msg string, isError bool) {
	l.statusbar.SetMessage(msg, isError)
}

// RefreshTimestamp updates the last refresh time in statusbar
func (l *Layout) RefreshTimestamp() {
	l.statusbar.SetLastRefresh(time.Now())
}

// Root returns the root tview primitive
func (l *Layout) Root() tview.Primitive {
	return l.root
}

// Header returns the header component for direct access
func (l *Layout) Header() *Header {
	return l.header
}

// NavBar returns the navbar component for direct access
func (l *Layout) NavBar() *NavBar {
	return l.navbar
}

// StatusBar returns the statusbar component for direct access
func (l *Layout) StatusBar() *StatusBar {
	return l.statusbar
}
