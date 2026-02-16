package views

import (
	"context"

	"github.com/etlmon/etlmon/ui"
	"github.com/rivo/tview"
)

// DetailProvider supplies tab-based detail content for a category
type DetailProvider interface {
	// Tabs returns the list of tab names for this category
	Tabs() []string
	// TabContent returns the tview Primitive for the given tab index
	TabContent(tabIndex int) tview.Primitive
	// Refresh fetches fresh data from the API for all tabs
	Refresh(ctx context.Context, client ui.APIClient) error
	// OnSelect is called when this category is selected with the current active tab index
	OnSelect(activeTabIndex int)
}

// MockProvider implements DetailProvider for testing
type MockProvider struct {
	TabNames       []string
	Contents       []tview.Primitive
	RefreshErr     error
	RefreshCalled  bool
	OnSelectCalled bool
	LastTabIndex   int
}

func (m *MockProvider) Tabs() []string { return m.TabNames }

func (m *MockProvider) TabContent(tabIndex int) tview.Primitive {
	if tabIndex >= 0 && tabIndex < len(m.Contents) {
		return m.Contents[tabIndex]
	}
	return nil
}

func (m *MockProvider) Refresh(ctx context.Context, client ui.APIClient) error {
	m.RefreshCalled = true
	return m.RefreshErr
}

func (m *MockProvider) OnSelect(activeTabIndex int) {
	m.OnSelectCalled = true
	m.LastTabIndex = activeTabIndex
}
