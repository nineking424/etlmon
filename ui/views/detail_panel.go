package views

import (
	"context"

	"github.com/etlmon/etlmon/ui"
	"github.com/rivo/tview"
)

// DetailPanel is a container for tab-based detail content
type DetailPanel struct {
	flex     *tview.Flex
	tabBar   *TabBar
	content  *tview.Flex
	provider DetailProvider
}

// NewDetailPanel creates a new DetailPanel
func NewDetailPanel() *DetailPanel {
	dp := &DetailPanel{
		flex:    tview.NewFlex(),
		tabBar:  NewTabBar(),
		content: tview.NewFlex(),
	}

	// Set up layout: TabBar (1 line) + Content (flex)
	dp.flex.SetDirection(tview.FlexRow)
	dp.flex.AddItem(dp.tabBar.Primitive(), 1, 0, false)
	dp.flex.AddItem(dp.content, 0, 1, true)

	// Connect tab bar changes to content updates
	dp.tabBar.SetChangedFunc(func(index int) {
		dp.updateContent(index)
		if dp.provider != nil {
			dp.provider.OnSelect(index)
		}
	})

	return dp
}

// SetProvider sets the detail provider and updates the display
func (dp *DetailPanel) SetProvider(p DetailProvider) {
	dp.provider = p

	if p == nil {
		dp.tabBar.SetTabs([]string{})
		dp.content.Clear()
		return
	}

	// Update tab bar with new tabs
	tabs := p.Tabs()
	dp.tabBar.SetTabs(tabs)

	// Show first tab content
	if len(tabs) > 0 {
		dp.tabBar.SetActiveTab(0)
		dp.updateContent(0)
		p.OnSelect(0)
	}
}

// SwitchTab changes the active tab
func (dp *DetailPanel) SwitchTab(index int) {
	dp.tabBar.SetActiveTab(index)
	// SetActiveTab will trigger the changed callback which updates content
}

// GetActiveTab returns the current active tab index
func (dp *DetailPanel) GetActiveTab() int {
	return dp.tabBar.GetActiveTab()
}

// NextTab moves to the next tab
func (dp *DetailPanel) NextTab() {
	dp.tabBar.NextTab()
}

// PrevTab moves to the previous tab
func (dp *DetailPanel) PrevTab() {
	dp.tabBar.PrevTab()
}

// Refresh calls the provider's Refresh method
func (dp *DetailPanel) Refresh(ctx context.Context, client ui.APIClient) error {
	if dp.provider == nil {
		return nil
	}
	return dp.provider.Refresh(ctx, client)
}

// Primitive returns the underlying tview primitive
func (dp *DetailPanel) Primitive() *tview.Flex {
	return dp.flex
}

// Focus sets focus on the content area
func (dp *DetailPanel) Focus(app *tview.Application) {
	app.SetFocus(dp.content)
}

// GetTabBar returns the tab bar (for testing)
func (dp *DetailPanel) GetTabBar() *TabBar {
	return dp.tabBar
}

// updateContent updates the content area with the primitive for the given tab
func (dp *DetailPanel) updateContent(index int) {
	dp.content.Clear()

	if dp.provider == nil {
		return
	}

	primitive := dp.provider.TabContent(index)
	if primitive != nil {
		dp.content.AddItem(primitive, 0, 1, true)
	}
}
