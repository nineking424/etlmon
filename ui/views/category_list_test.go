package views

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestCategoryList_Creation verifies that CategoryList is created with 4 categories
func TestCategoryList_Creation(t *testing.T) {
	cl := NewCategoryList()
	if cl == nil {
		t.Fatal("NewCategoryList returned nil")
	}

	// Verify 4 categories exist
	expectedCount := 4
	if cl.list.GetItemCount() != expectedCount {
		t.Errorf("Expected %d categories, got %d", expectedCount, cl.list.GetItemCount())
	}

	// Verify category names
	expectedCategories := []string{"FS", "Paths", "Process", "Logs"}
	for i, expected := range expectedCategories {
		main, _ := cl.list.GetItemText(i)
		if main != expected {
			t.Errorf("Category %d: expected %q, got %q", i, expected, main)
		}
	}
}

// TestCategoryList_InitialSelection verifies FS is selected by default
func TestCategoryList_InitialSelection(t *testing.T) {
	cl := NewCategoryList()

	// Verify first item (FS) is selected
	current := cl.list.GetCurrentItem()
	if current != 0 {
		t.Errorf("Expected initial selection at index 0, got %d", current)
	}
}

// TestCategoryList_NavigationJK verifies j/k vim-style navigation
func TestCategoryList_NavigationJK(t *testing.T) {
	app := tview.NewApplication()
	cl := NewCategoryList()

	// Set up the list in the app
	app.SetRoot(cl.list, true)

	// Initial position should be 0
	if cl.list.GetCurrentItem() != 0 {
		t.Fatalf("Initial position should be 0, got %d", cl.list.GetCurrentItem())
	}

	// Simulate 'j' key press (move down)
	event := tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)
	if handler := cl.list.GetInputCapture(); handler != nil {
		handler(event)
	}

	// Should move to index 1
	if cl.list.GetCurrentItem() != 1 {
		t.Errorf("After 'j', expected position 1, got %d", cl.list.GetCurrentItem())
	}

	// Simulate 'k' key press (move up)
	event = tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)
	if handler := cl.list.GetInputCapture(); handler != nil {
		handler(event)
	}

	// Should move back to index 0
	if cl.list.GetCurrentItem() != 0 {
		t.Errorf("After 'k', expected position 0, got %d", cl.list.GetCurrentItem())
	}
}

// TestCategoryList_QuickJump verifies 1-4 number keys for quick jump
func TestCategoryList_QuickJump(t *testing.T) {
	app := tview.NewApplication()
	cl := NewCategoryList()
	app.SetRoot(cl.list, true)

	tests := []struct {
		key      rune
		expected int
	}{
		{'1', 0}, // FS
		{'2', 1}, // Paths
		{'3', 2}, // Process
		{'4', 3}, // Logs
	}

	for _, tt := range tests {
		event := tcell.NewEventKey(tcell.KeyRune, tt.key, tcell.ModNone)
		if handler := cl.list.GetInputCapture(); handler != nil {
			handler(event)
		}

		if cl.list.GetCurrentItem() != tt.expected {
			t.Errorf("After key '%c', expected position %d, got %d", tt.key, tt.expected, cl.list.GetCurrentItem())
		}
	}
}

// TestCategoryList_SelectionCallback verifies that the callback is called on selection change
func TestCategoryList_SelectionCallback(t *testing.T) {
	cl := NewCategoryList()

	var callbackIndex int
	var callbackName string
	var callbackCalled bool

	cl.SetChangedFunc(func(index int, name string) {
		callbackIndex = index
		callbackName = name
		callbackCalled = true
	})

	// Trigger selection change programmatically
	cl.list.SetCurrentItem(2) // Select "Process"

	// The changed callback should be triggered
	if !callbackCalled {
		t.Error("Expected callback to be called, but it wasn't")
	}

	if callbackIndex != 2 {
		t.Errorf("Expected callback index 2, got %d", callbackIndex)
	}

	if callbackName != "Process" {
		t.Errorf("Expected callback name 'Process', got %q", callbackName)
	}
}

// TestCategoryList_FocusHighlight verifies focus highlighting behavior
func TestCategoryList_FocusHighlight(t *testing.T) {
	cl := NewCategoryList()

	// When focused, the list should have highlight enabled
	// This is a visual test, but we can verify the list has the right configuration
	if cl.list == nil {
		t.Fatal("List should not be nil")
	}

	// Verify the primitive can be focused (non-nil check)
	if cl.Primitive() == nil {
		t.Error("Primitive() should return non-nil")
	}
}
