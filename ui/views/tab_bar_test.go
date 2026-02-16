package views

import (
	"strings"
	"testing"

	"github.com/etlmon/etlmon/ui/theme"
	"github.com/gdamore/tcell/v2"
)

func TestTabBar_Render(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"Info", "Details", "Stats"})

	text := tb.GetRenderedText()

	// Check that all tab names are present in the rendered text
	if !strings.Contains(text, "Info") {
		t.Errorf("Expected tab 'Info' to be rendered, got: %s", text)
	}
	if !strings.Contains(text, "Details") {
		t.Errorf("Expected tab 'Details' to be rendered, got: %s", text)
	}
	if !strings.Contains(text, "Stats") {
		t.Errorf("Expected tab 'Stats' to be rendered, got: %s", text)
	}
}

func TestTabBar_ActiveHighlight(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"Tab1", "Tab2", "Tab3"})

	// Set active tab to index 1
	tb.SetActiveTab(1)

	text := tb.GetRenderedText()

	// Active tab should have accent color tag and bold
	// The exact format depends on implementation, but it should contain the accent color
	if !strings.Contains(text, theme.TagAccent) || !strings.Contains(text, theme.TagBold) {
		t.Errorf("Expected active tab to have accent color and bold formatting, got: %s", text)
	}
}

func TestTabBar_SwitchBrackets(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"A", "B", "C"})

	// Start at tab 0
	if tb.GetActiveTab() != 0 {
		t.Errorf("Expected initial active tab to be 0, got %d", tb.GetActiveTab())
	}

	// Next tab should move to 1
	tb.NextTab()
	if tb.GetActiveTab() != 1 {
		t.Errorf("Expected active tab to be 1 after NextTab, got %d", tb.GetActiveTab())
	}

	// Prev tab should move back to 0
	tb.PrevTab()
	if tb.GetActiveTab() != 0 {
		t.Errorf("Expected active tab to be 0 after PrevTab, got %d", tb.GetActiveTab())
	}
}

func TestTabBar_BoundaryWrap(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"A", "B", "C"})

	// Start at tab 0, PrevTab should wrap to last tab (2)
	tb.SetActiveTab(0)
	tb.PrevTab()
	if tb.GetActiveTab() != 2 {
		t.Errorf("Expected PrevTab from 0 to wrap to 2, got %d", tb.GetActiveTab())
	}

	// From last tab (2), NextTab should wrap to first tab (0)
	tb.SetActiveTab(2)
	tb.NextTab()
	if tb.GetActiveTab() != 0 {
		t.Errorf("Expected NextTab from 2 to wrap to 0, got %d", tb.GetActiveTab())
	}
}

func TestTabBar_SetChangedFunc(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"A", "B", "C"})

	callbackFired := false
	var callbackIndex int

	tb.SetChangedFunc(func(index int) {
		callbackFired = true
		callbackIndex = index
	})

	// Change tab should fire callback
	tb.NextTab()

	if !callbackFired {
		t.Errorf("Expected callback to fire after NextTab")
	}
	if callbackIndex != 1 {
		t.Errorf("Expected callback index to be 1, got %d", callbackIndex)
	}
}

func TestTabBar_EmptyTabs(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{})

	// Should handle empty tabs gracefully
	if tb.GetActiveTab() != 0 {
		t.Errorf("Expected active tab to be 0 for empty tabs, got %d", tb.GetActiveTab())
	}

	// NextTab/PrevTab should not panic
	tb.NextTab()
	tb.PrevTab()
}

func TestTabBar_InputCapture(t *testing.T) {
	tb := NewTabBar()
	tb.SetTabs([]string{"A", "B", "C"})

	// Simulate ] key press (NextTab)
	event := tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModNone)
	result := tb.Primitive().GetInputCapture()(event)

	// Event should be consumed (nil returned)
	if result != nil {
		t.Errorf("Expected ] key to be consumed by TabBar")
	}

	// Active tab should have moved
	if tb.GetActiveTab() != 1 {
		t.Errorf("Expected active tab to be 1 after ] key, got %d", tb.GetActiveTab())
	}

	// Simulate [ key press (PrevTab)
	event = tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone)
	result = tb.Primitive().GetInputCapture()(event)

	// Event should be consumed
	if result != nil {
		t.Errorf("Expected [ key to be consumed by TabBar")
	}

	// Active tab should have moved back
	if tb.GetActiveTab() != 0 {
		t.Errorf("Expected active tab to be 0 after [ key, got %d", tb.GetActiveTab())
	}
}
