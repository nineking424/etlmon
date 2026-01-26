package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/rivo/tview"
)

// StatusBar displays status information
type StatusBar struct {
	view       *tview.TextView
	status     string
	lastUpdate time.Time
	mu         sync.RWMutex
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	return &StatusBar{
		view:   view,
		status: "Initializing...",
	}
}

// SetStatus sets the status message
func (s *StatusBar) SetStatus(status string) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
	s.render()
}

// SetLastUpdate sets the last update time
func (s *StatusBar) SetLastUpdate(t time.Time) {
	s.mu.Lock()
	s.lastUpdate = t
	s.mu.Unlock()
	s.render()
}

// render updates the status bar text
func (s *StatusBar) render() {
	s.mu.RLock()
	status := s.status
	lastUpdate := s.lastUpdate
	s.mu.RUnlock()

	var text string
	if lastUpdate.IsZero() {
		text = fmt.Sprintf("[green]%s[white] | [gray]etlmon[white]", status)
	} else {
		text = fmt.Sprintf("[green]%s[white] | Last: %s | [gray]etlmon[white]",
			status, lastUpdate.Format("15:04:05"))
	}

	// Need write lock when setting text to avoid race with GetText
	s.mu.Lock()
	s.view.SetText(text)
	s.mu.Unlock()
}

// GetText returns the current status text
func (s *StatusBar) GetText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.view.GetText(true)
}
