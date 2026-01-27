package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
	"github.com/etlmon/etlmon/internal/storage"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewRealtime ViewMode = iota
	ViewHistory
)

// Store interface for database operations
type Store interface {
	GetMetrics(opts storage.GetMetricsOptions) ([]*storage.AggregatedMetric, error)
	GetLatestMetrics(resourceType, windowSize string) ([]*storage.AggregatedMetric, error)
}

// App is the main TUI application
type App struct {
	app          *tview.Application
	pages        *tview.Pages
	realtimeView *RealtimeView
	historyView  *HistoryView
	statusBar    *StatusBar
	currentView  ViewMode
	store        Store
	running      bool
	mu           sync.RWMutex
}

// NewApp creates a new TUI application
func NewApp() *App {
	a := &App{
		app:          tview.NewApplication(),
		pages:        tview.NewPages(),
		realtimeView: NewRealtimeView(),
		historyView:  NewHistoryView(),
		statusBar:    NewStatusBar(),
		currentView:  ViewRealtime,
	}

	a.setupLayout()
	a.setupKeybindings()

	return a
}

// setupLayout creates the UI layout
func (a *App) setupLayout() {
	// Main layout with status bar at bottom
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.statusBar.view, 1, 0, false)

	// Add views to pages
	a.pages.AddPage("realtime", a.realtimeView.view, true, true)
	a.pages.AddPage("history", a.historyView.view, true, false)

	a.app.SetRoot(mainFlex, true)
}

// setupKeybindings configures keyboard shortcuts
func (a *App) setupKeybindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		case tcell.KeyTab:
			a.toggleView()
			return nil
		}

		switch event.Rune() {
		case 'q', 'Q':
			a.app.Stop()
			return nil
		case 'r', 'R':
			a.SwitchView(ViewRealtime)
			return nil
		case 'h', 'H':
			a.SwitchView(ViewHistory)
			return nil
		case '1':
			a.historyView.SetWindowFilter("1m")
			a.refreshHistory()
			return nil
		case '5':
			a.historyView.SetWindowFilter("5m")
			a.refreshHistory()
			return nil
		case '0':
			a.historyView.SetWindowFilter("1h")
			a.refreshHistory()
			return nil
		case 't', 'T':
			a.mu.RLock()
			currentView := a.currentView
			a.mu.RUnlock()
			if currentView == ViewRealtime {
				a.realtimeView.ToggleDisplayFormat()
				a.app.Draw()
			}
			return nil
		}

		return event
	})
}

// toggleView switches between realtime and history views
func (a *App) toggleView() {
	if a.currentView == ViewRealtime {
		a.SwitchView(ViewHistory)
	} else {
		a.SwitchView(ViewRealtime)
	}
}

// SwitchView changes the current view
func (a *App) SwitchView(mode ViewMode) {
	a.mu.Lock()
	a.currentView = mode
	a.mu.Unlock()

	switch mode {
	case ViewRealtime:
		a.pages.SwitchToPage("realtime")
	case ViewHistory:
		a.pages.SwitchToPage("history")
		a.refreshHistory()
	}
}

// refreshHistory loads and displays history data
func (a *App) refreshHistory() {
	if a.store == nil {
		return
	}

	a.mu.RLock()
	filter := a.historyView.windowFilter
	a.mu.RUnlock()

	metrics, err := a.store.GetMetrics(storage.GetMetricsOptions{
		WindowSize: filter,
		Limit:      100,
	})
	if err != nil {
		return
	}

	// Convert to AggregationResult for display
	var results []aggregator.AggregationResult
	for _, m := range metrics {
		windowSize, _ := time.ParseDuration(m.WindowSize)
		results = append(results, aggregator.AggregationResult{
			ResourceType:    m.ResourceType,
			MetricName:      m.MetricName,
			Value:           m.AggregatedValue,
			WindowSize:      windowSize,
			AggregationType: m.AggregationType,
			Timestamp:       time.Unix(m.Timestamp, 0),
		})
	}

	a.historyView.Update(results)
}

// SetStore sets the storage backend
func (a *App) SetStore(store Store) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store = store
}

// OnMetricsCollected is called when new metrics are collected
func (a *App) OnMetricsCollected(metrics []collector.Metric) {
	a.realtimeView.Update(metrics)
	a.statusBar.SetLastUpdate(time.Now())
	// Only draw if the app is running
	a.mu.RLock()
	running := a.running
	a.mu.RUnlock()
	if running {
		a.app.Draw()
	}
}

// OnAggregationComplete is called when aggregation results are ready
func (a *App) OnAggregationComplete(results []aggregator.AggregationResult) {
	// Add to history view
	a.historyView.Update(results)
	// Only draw if the app is running
	a.mu.RLock()
	running := a.running
	a.mu.RUnlock()
	if running {
		a.app.Draw()
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	a.mu.Lock()
	a.running = true
	a.mu.Unlock()
	a.statusBar.SetStatus("Running")
	err := a.app.Run()
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()
	return err
}

// Stop stops the TUI application
func (a *App) Stop() {
	a.app.Stop()
}

// QueueUpdateDraw queues a UI update
func (a *App) QueueUpdateDraw(f func()) {
	a.app.QueueUpdateDraw(f)
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", bytes/div, "KMGTPE"[exp])
}

// FormatDuration formats duration to short string
func FormatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
