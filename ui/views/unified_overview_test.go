package views

import (
	"context"
	"errors"
	"testing"

	"github.com/etlmon/etlmon/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestUnifiedOverview_Name verifies the view name
func TestUnifiedOverview_Name(t *testing.T) {
	uo := NewUnifiedOverview(nil, nil)
	if uo.Name() != "overview" {
		t.Errorf("Expected name 'overview', got %q", uo.Name())
	}
}

// TestUnifiedOverview_ImplementsView is a compile-time check
func TestUnifiedOverview_ImplementsView(t *testing.T) {
	var _ ui.View = (*UnifiedOverview)(nil)
}

// TestUnifiedOverview_InitialState verifies FS is selected and first tab is active
func TestUnifiedOverview_InitialState(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	// Verify FS category (index 0) is selected
	if uo.currentCat != 0 {
		t.Errorf("Expected initial category 0 (FS), got %d", uo.currentCat)
	}

	// Verify first tab is active
	if uo.detailPanel.GetActiveTab() != 0 {
		t.Errorf("Expected initial tab 0, got %d", uo.detailPanel.GetActiveTab())
	}

	// Verify provider is set (should be FS provider)
	if len(uo.providers) != 4 {
		t.Errorf("Expected 4 providers, got %d", len(uo.providers))
	}
}

// TestUnifiedOverview_CategorySwitch verifies category switching updates provider
func TestUnifiedOverview_CategorySwitch(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	// Switch to Paths (index 1)
	uo.switchCategory(1)

	if uo.currentCat != 1 {
		t.Errorf("Expected category 1 (Paths), got %d", uo.currentCat)
	}

	// Verify provider tabs changed (Paths provider should have different tabs than FS)
	tabs := uo.providers[1].Tabs()
	if len(tabs) == 0 {
		t.Error("Expected Paths provider to have tabs")
	}
}

// TestUnifiedOverview_TabSwitch verifies tab switching within a category
func TestUnifiedOverview_TabSwitch(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	// Get initial tab count
	initialTab := uo.detailPanel.GetActiveTab()
	if initialTab != 0 {
		t.Errorf("Expected initial tab 0, got %d", initialTab)
	}

	// Switch to next tab
	uo.detailPanel.NextTab()

	newTab := uo.detailPanel.GetActiveTab()
	if newTab == initialTab {
		t.Error("Tab should have changed after NextTab()")
	}
}

// TestUnifiedOverview_FocusToggle verifies Tab key toggles focus between panes
func TestUnifiedOverview_FocusToggle(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	// Initial focus should be on category list (0)
	if uo.focusedPane != 0 {
		t.Errorf("Expected initial focus on pane 0, got %d", uo.focusedPane)
	}

	// Simulate Tab key
	event := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	if handler := uo.flex.GetInputCapture(); handler != nil {
		handler(event)
	}

	// Focus should move to detail panel (1)
	if uo.focusedPane != 1 {
		t.Errorf("After Tab, expected focus on pane 1, got %d", uo.focusedPane)
	}

	// Tab again should return to category list (0)
	if handler := uo.flex.GetInputCapture(); handler != nil {
		handler(event)
	}

	if uo.focusedPane != 0 {
		t.Errorf("After second Tab, expected focus on pane 0, got %d", uo.focusedPane)
	}
}

// TestUnifiedOverview_QuickJump verifies 1-4 keys jump to categories
func TestUnifiedOverview_QuickJump(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	tests := []struct {
		key      rune
		expected int
		name     string
	}{
		{'1', 0, "FS"},
		{'2', 1, "Paths"},
		{'3', 2, "Process"},
		{'4', 3, "Logs"},
	}

	for _, tt := range tests {
		event := tcell.NewEventKey(tcell.KeyRune, tt.key, tcell.ModNone)
		if handler := uo.flex.GetInputCapture(); handler != nil {
			handler(event)
		}

		if uo.currentCat != tt.expected {
			t.Errorf("After key '%c', expected category %d (%s), got %d", tt.key, tt.expected, tt.name, uo.currentCat)
		}
	}
}

// TestUnifiedOverview_Refresh_Success verifies Refresh propagates to provider
func TestUnifiedOverview_Refresh_Success(t *testing.T) {
	app := tview.NewApplication()
	mockClient := &mockAPIClient{}
	uo := NewUnifiedOverview(app, mockClient)

	// Replace first provider with mock
	mockProv := &MockProvider{
		TabNames: []string{"Tab1", "Tab2"},
		Contents: []tview.Primitive{tview.NewTextView(), tview.NewTextView()},
	}
	uo.providers[0] = mockProv
	uo.switchCategory(0)

	// Call Refresh
	ctx := context.Background()
	err := uo.Refresh(ctx, mockClient)

	if err != nil {
		t.Errorf("Refresh should not error, got: %v", err)
	}

	if !mockProv.RefreshCalled {
		t.Error("Provider Refresh should have been called")
	}
}

// TestUnifiedOverview_Refresh_PartialFailure verifies partial failure handling
func TestUnifiedOverview_Refresh_PartialFailure(t *testing.T) {
	app := tview.NewApplication()
	mockClient := &mockAPIClient{}
	uo := NewUnifiedOverview(app, mockClient)

	// Replace first provider with failing mock
	mockProv := &MockProvider{
		TabNames:   []string{"Tab1"},
		Contents:   []tview.Primitive{tview.NewTextView()},
		RefreshErr: errors.New("API error"),
	}
	uo.providers[0] = mockProv
	uo.switchCategory(0)

	// Call Refresh
	ctx := context.Background()
	err := uo.Refresh(ctx, mockClient)

	if err == nil {
		t.Error("Refresh should return error when provider fails")
	}

	if !mockProv.RefreshCalled {
		t.Error("Provider Refresh should have been called even if it fails")
	}
}

// TestUnifiedOverview_FocusRestore verifies focus is properly restored
func TestUnifiedOverview_FocusRestore(t *testing.T) {
	app := tview.NewApplication()
	uo := NewUnifiedOverview(app, nil)

	// Set focus to detail panel
	uo.focusedPane = 1

	// Call Focus() - should restore to detail panel
	uo.Focus()

	// Verify focusedPane is still 1 (this is a state check, actual tview focus requires running app)
	if uo.focusedPane != 1 {
		t.Errorf("Expected focusedPane to remain 1, got %d", uo.focusedPane)
	}
}

// mockAPIClient is already defined in provider_fs_test.go
