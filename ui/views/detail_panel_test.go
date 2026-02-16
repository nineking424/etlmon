package views

import (
	"context"
	"errors"
	"testing"

	"github.com/rivo/tview"
)

func TestDetailPanel_SetProvider(t *testing.T) {
	dp := NewDetailPanel()

	// Create mock provider with 3 tabs
	provider := &MockProvider{
		TabNames: []string{"Tab1", "Tab2", "Tab3"},
		Contents: []tview.Primitive{
			tview.NewTextView().SetText("Content 1"),
			tview.NewTextView().SetText("Content 2"),
			tview.NewTextView().SetText("Content 3"),
		},
	}

	dp.SetProvider(provider)

	// Active tab should be 0
	if dp.GetActiveTab() != 0 {
		t.Errorf("Expected active tab to be 0 after SetProvider, got %d", dp.GetActiveTab())
	}

	// Tab bar should show 3 tabs
	tabBar := dp.GetTabBar()
	if len(tabBar.tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(tabBar.tabs))
	}
}

func TestDetailPanel_TabSwitch(t *testing.T) {
	dp := NewDetailPanel()

	content1 := tview.NewTextView().SetText("Content 1")
	content2 := tview.NewTextView().SetText("Content 2")

	provider := &MockProvider{
		TabNames: []string{"Tab1", "Tab2"},
		Contents: []tview.Primitive{content1, content2},
	}

	dp.SetProvider(provider)

	// Initially on tab 0
	if dp.GetActiveTab() != 0 {
		t.Errorf("Expected initial tab to be 0, got %d", dp.GetActiveTab())
	}

	// Switch to tab 1
	dp.SwitchTab(1)

	if dp.GetActiveTab() != 1 {
		t.Errorf("Expected active tab to be 1 after SwitchTab, got %d", dp.GetActiveTab())
	}

	// OnSelect should have been called
	if !provider.OnSelectCalled {
		t.Errorf("Expected OnSelect to be called on provider")
	}

	if provider.LastTabIndex != 1 {
		t.Errorf("Expected LastTabIndex to be 1, got %d", provider.LastTabIndex)
	}
}

func TestDetailPanel_NextPrevTab(t *testing.T) {
	dp := NewDetailPanel()

	provider := &MockProvider{
		TabNames: []string{"A", "B", "C"},
		Contents: []tview.Primitive{
			tview.NewTextView(),
			tview.NewTextView(),
			tview.NewTextView(),
		},
	}

	dp.SetProvider(provider)

	// Start at tab 0
	if dp.GetActiveTab() != 0 {
		t.Errorf("Expected initial tab to be 0, got %d", dp.GetActiveTab())
	}

	// NextTab -> 1
	dp.NextTab()
	if dp.GetActiveTab() != 1 {
		t.Errorf("Expected tab to be 1 after NextTab, got %d", dp.GetActiveTab())
	}

	// NextTab -> 2
	dp.NextTab()
	if dp.GetActiveTab() != 2 {
		t.Errorf("Expected tab to be 2 after NextTab, got %d", dp.GetActiveTab())
	}

	// NextTab -> 0 (wrap)
	dp.NextTab()
	if dp.GetActiveTab() != 0 {
		t.Errorf("Expected tab to wrap to 0, got %d", dp.GetActiveTab())
	}

	// PrevTab -> 2 (wrap backwards)
	dp.PrevTab()
	if dp.GetActiveTab() != 2 {
		t.Errorf("Expected tab to wrap to 2, got %d", dp.GetActiveTab())
	}
}

func TestDetailPanel_Refresh(t *testing.T) {
	dp := NewDetailPanel()

	provider := &MockProvider{
		TabNames: []string{"Tab1"},
		Contents: []tview.Primitive{tview.NewTextView()},
	}

	dp.SetProvider(provider)

	// Call Refresh
	err := dp.Refresh(context.Background(), nil)

	if err != nil {
		t.Errorf("Expected no error from Refresh, got %v", err)
	}

	if !provider.RefreshCalled {
		t.Errorf("Expected provider.Refresh to be called")
	}
}

func TestDetailPanel_RefreshError(t *testing.T) {
	dp := NewDetailPanel()

	expectedErr := errors.New("refresh failed")
	provider := &MockProvider{
		TabNames:   []string{"Tab1"},
		Contents:   []tview.Primitive{tview.NewTextView()},
		RefreshErr: expectedErr,
	}

	dp.SetProvider(provider)

	// Call Refresh
	err := dp.Refresh(context.Background(), nil)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestDetailPanel_NilProvider(t *testing.T) {
	dp := NewDetailPanel()

	// Should handle nil provider gracefully
	dp.SetProvider(nil)

	// These should not panic
	dp.NextTab()
	dp.PrevTab()
	dp.SwitchTab(0)

	err := dp.Refresh(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected no error with nil provider, got %v", err)
	}
}

func TestDetailPanel_InputCapture(t *testing.T) {
	dp := NewDetailPanel()

	provider := &MockProvider{
		TabNames: []string{"A", "B", "C"},
		Contents: []tview.Primitive{
			tview.NewTextView(),
			tview.NewTextView(),
			tview.NewTextView(),
		},
	}

	dp.SetProvider(provider)

	// Verify tab bar input capture is set up
	tabBar := dp.GetTabBar()
	if tabBar.Primitive().GetInputCapture() == nil {
		t.Errorf("Expected TabBar to have input capture set")
	}
}
