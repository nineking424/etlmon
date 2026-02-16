package views

import (
	"context"

	"github.com/etlmon/etlmon/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UnifiedOverview is a unified overview view combining CategoryList and DetailPanel
type UnifiedOverview struct {
	flex           *tview.Flex
	categoryList   *CategoryList
	detailPanel    *DetailPanel
	providers      []DetailProvider
	currentCat     int
	focusedPane    int // 0=category, 1=detail
	tviewApp       *tview.Application
	apiClient      ui.APIClient
	onStatusChange func(msg string, isError bool)
}

// NewUnifiedOverview creates a new UnifiedOverview
func NewUnifiedOverview(app *tview.Application, client ui.APIClient) *UnifiedOverview {
	uo := &UnifiedOverview{
		flex:         tview.NewFlex(),
		categoryList: NewCategoryList(),
		detailPanel:  NewDetailPanel(),
		providers:    make([]DetailProvider, 4),
		currentCat:   0,
		focusedPane:  0,
		tviewApp:     app,
		apiClient:    client,
	}

	// Initialize providers for each category
	uo.providers[0] = NewFSDetailProvider()                    // FS
	uo.providers[1] = NewPathsDetailProvider(client, app)      // Paths
	uo.providers[2] = NewProcessDetailProvider()               // Process
	uo.providers[3] = NewLogsDetailProvider(client, app)       // Logs

	// Set up layout: CategoryList (20 fixed) + DetailPanel (flex)
	uo.flex.SetDirection(tview.FlexColumn)
	uo.flex.AddItem(uo.categoryList.Primitive(), 20, 0, true)
	uo.flex.AddItem(uo.detailPanel.Primitive(), 0, 1, false)

	// Set up category selection callback
	uo.categoryList.SetChangedFunc(func(index int, name string) {
		uo.switchCategory(index)
	})

	// Set up global input capture for navigation
	uo.flex.SetInputCapture(uo.handleInput)

	// Initialize with first category (FS)
	uo.switchCategory(0)

	return uo
}

// Name returns the view name
func (uo *UnifiedOverview) Name() string {
	return "overview"
}

// Primitive returns the root tview primitive
func (uo *UnifiedOverview) Primitive() tview.Primitive {
	return uo.flex
}

// Refresh updates the current provider with fresh data
func (uo *UnifiedOverview) Refresh(ctx context.Context, client ui.APIClient) error {
	if uo.currentCat >= 0 && uo.currentCat < len(uo.providers) {
		provider := uo.providers[uo.currentCat]
		if provider != nil {
			return provider.Refresh(ctx, client)
		}
	}
	return nil
}

// Focus sets focus based on focusedPane
func (uo *UnifiedOverview) Focus() {
	if uo.tviewApp == nil {
		return
	}

	if uo.focusedPane == 0 {
		// Focus on category list
		uo.tviewApp.SetFocus(uo.categoryList.Primitive())
	} else {
		// Focus on detail panel
		uo.detailPanel.Focus(uo.tviewApp)
	}
}

// SetStatusCallback sets the status message callback
func (uo *UnifiedOverview) SetStatusCallback(cb func(msg string, isError bool)) {
	uo.onStatusChange = cb
}

// switchCategory changes the active category and updates the detail panel
func (uo *UnifiedOverview) switchCategory(index int) {
	if index < 0 || index >= len(uo.providers) {
		return
	}

	uo.currentCat = index
	provider := uo.providers[index]

	// Set provider to detail panel
	uo.detailPanel.SetProvider(provider)

	// Trigger OnSelect for the active tab
	activeTab := uo.detailPanel.GetActiveTab()
	if provider != nil {
		provider.OnSelect(activeTab)
	}
}

// handleInput processes keyboard input for navigation
func (uo *UnifiedOverview) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		// Toggle focus between category list and detail panel
		uo.focusedPane = (uo.focusedPane + 1) % 2
		uo.Focus()
		return nil
	}

	switch event.Rune() {
	case '1':
		// Quick jump to FS
		uo.categoryList.list.SetCurrentItem(0)
		return nil
	case '2':
		// Quick jump to Paths
		uo.categoryList.list.SetCurrentItem(1)
		return nil
	case '3':
		// Quick jump to Process
		uo.categoryList.list.SetCurrentItem(2)
		return nil
	case '4':
		// Quick jump to Logs
		uo.categoryList.list.SetCurrentItem(3)
		return nil
	case '[':
		// Previous tab (delegate to detail panel)
		if uo.focusedPane == 1 {
			uo.detailPanel.PrevTab()
			return nil
		}
	case ']':
		// Next tab (delegate to detail panel)
		if uo.focusedPane == 1 {
			uo.detailPanel.NextTab()
			return nil
		}
	case 'j':
		// Move down in active pane
		if uo.focusedPane == 0 {
			// Delegate to category list
			return event
		} else {
			// Delegate to detail panel content (pass through)
			return event
		}
	case 'k':
		// Move up in active pane
		if uo.focusedPane == 0 {
			// Delegate to category list
			return event
		} else {
			// Delegate to detail panel content (pass through)
			return event
		}
	}

	return event
}
