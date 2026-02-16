package layout

import (
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/rivo/tview"
)

// NavItem represents a navigation item
type NavItem struct {
	Key      rune
	Name     string
	ViewName string
}

// NavBar displays navigation tabs and shortcuts
type NavBar struct {
	textView *tview.TextView
	items    []NavItem
	active   string
}

// NewNavBar creates a new navigation bar
func NewNavBar() *NavBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.BgNavBar)

	return &NavBar{
		textView: tv,
		items:    []NavItem{},
	}
}

// SetActive sets the active view and updates display
func (n *NavBar) SetActive(viewName string) {
	n.active = viewName
	n.render()
}

// render updates the navbar display
func (n *NavBar) render() {
	shortcuts := " [teal]Tab[silver]=panel  [teal][[silver]/[teal]][silver]=tab  [teal]j[silver]/[teal]k[silver]=nav  [teal]s[silver]=settings  [teal]?[silver]=help  [teal]r[silver]=refresh  [teal]q[silver]=quit "
	n.textView.SetText(shortcuts)
}

// Primitive returns the navbar's tview primitive
func (n *NavBar) Primitive() tview.Primitive {
	return n.textView
}

// AddItem adds a navigation item
func (n *NavBar) AddItem(key rune, name, viewName string) {
	n.items = append(n.items, NavItem{Key: key, Name: name, ViewName: viewName})
}
