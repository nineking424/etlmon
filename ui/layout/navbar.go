package layout

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	tv.SetBackgroundColor(tcell.ColorDarkBlue)

	return &NavBar{
		textView: tv,
		items: []NavItem{
			{Key: '1', Name: "FS", ViewName: "fs"},
			{Key: '2', Name: "Paths", ViewName: "paths"},
		},
	}
}

// SetActive sets the active view and updates display
func (n *NavBar) SetActive(viewName string) {
	n.active = viewName
	n.render()
}

// render updates the navbar display
func (n *NavBar) render() {
	var parts []string

	for _, item := range n.items {
		if item.ViewName == n.active {
			// Active item - highlighted
			parts = append(parts, fmt.Sprintf("[black:aqua] <%c> %s [-:-]", item.Key, item.Name))
		} else {
			// Inactive item
			parts = append(parts, fmt.Sprintf("[white:-] <%c> %s [-:-]", item.Key, item.Name))
		}
	}

	// Add separator and shortcuts
	shortcuts := "[gray]│[-]  [yellow]?[white]=help  [yellow]r[white]=refresh  [yellow]s[white]=scan  [yellow]T[white]=border  [yellow]q[white]=quit"

	text := " " + strings.Join(parts, "  ") + "  " + shortcuts + " "
	n.textView.SetText(text)
}

// Primitive returns the navbar's tview primitive
func (n *NavBar) Primitive() tview.Primitive {
	return n.textView
}

// AddItem adds a navigation item
func (n *NavBar) AddItem(key rune, name, viewName string) {
	n.items = append(n.items, NavItem{Key: key, Name: name, ViewName: viewName})
}
