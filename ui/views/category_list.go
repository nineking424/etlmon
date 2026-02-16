package views

import (
	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CategoryList is a vertical list of categories with vim-style navigation
type CategoryList struct {
	list       *tview.List
	categories []string
	onChanged  func(index int, name string)
}

// NewCategoryList creates a new CategoryList with 4 categories
func NewCategoryList() *CategoryList {
	categories := []string{"FS", "Paths", "Process", "Logs"}

	list := tview.NewList()
	list.SetBorder(true).
		SetTitle(" Categories ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.FgLabel)

	// Add categories
	for _, cat := range categories {
		list.AddItem(cat, "", 0, nil)
	}

	// Configure list styling
	list.SetHighlightFullLine(true).
		SetSelectedBackgroundColor(theme.BgSelected).
		SetSelectedTextColor(theme.FgPrimary).
		SetMainTextColor(theme.FgSecondary)

	cl := &CategoryList{
		list:       list,
		categories: categories,
	}

	// Set up vim-style navigation and quick jump
	list.SetInputCapture(cl.handleInput)

	// Set up selection change callback
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if cl.onChanged != nil {
			cl.onChanged(index, mainText)
		}
	})

	// Select first item by default
	list.SetCurrentItem(0)

	return cl
}

// handleInput processes keyboard input for vim-style navigation and quick jump
func (cl *CategoryList) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'j':
		// Move down
		current := cl.list.GetCurrentItem()
		if current < len(cl.categories)-1 {
			cl.list.SetCurrentItem(current + 1)
		}
		return nil
	case 'k':
		// Move up
		current := cl.list.GetCurrentItem()
		if current > 0 {
			cl.list.SetCurrentItem(current - 1)
		}
		return nil
	case '1':
		cl.list.SetCurrentItem(0)
		return nil
	case '2':
		cl.list.SetCurrentItem(1)
		return nil
	case '3':
		cl.list.SetCurrentItem(2)
		return nil
	case '4':
		cl.list.SetCurrentItem(3)
		return nil
	}
	return event
}

// SetChangedFunc sets the callback for selection changes
func (cl *CategoryList) SetChangedFunc(handler func(index int, name string)) {
	cl.onChanged = handler
}

// Primitive returns the underlying tview.Primitive
func (cl *CategoryList) Primitive() tview.Primitive {
	return cl.list
}

// GetCurrentItem returns the currently selected category index
func (cl *CategoryList) GetCurrentItem() int {
	return cl.list.GetCurrentItem()
}

// GetCurrentCategory returns the currently selected category name
func (cl *CategoryList) GetCurrentCategory() string {
	index := cl.list.GetCurrentItem()
	if index >= 0 && index < len(cl.categories) {
		return cl.categories[index]
	}
	return ""
}
