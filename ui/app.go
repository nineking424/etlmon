package ui

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/ui/layout"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main TUI application
type App struct {
	tview      *tview.Application
	client     APIClient
	layout     *layout.Layout
	mainPages  *tview.Pages
	overview   View
	settings   View
	help       View
	currentPage string
}

// NewApp creates a new TUI application
func NewApp(client APIClient) *App {
	return &App{
		tview:       tview.NewApplication(),
		client:      client,
		layout:      layout.NewLayout(),
		mainPages:   tview.NewPages(),
		currentPage: "overview",
	}
}

// GetTviewApp returns the underlying tview application
func (a *App) GetTviewApp() *tview.Application {
	return a.tview
}

// SetPages configures the main pages (overview, settings, help)
func (a *App) SetPages(overview, settings, help View) {
	a.overview = overview
	a.settings = settings
	a.help = help

	// Add pages to the Pages container
	a.mainPages.AddPage("overview", overview.Primitive(), true, true)
	a.mainPages.AddPage("settings", settings.Primitive(), true, false)
	a.mainPages.AddPage("help", help.Primitive(), true, false)

	// Set status callback if view supports it
	type statusCallbackSetter interface {
		SetStatusCallback(func(msg string, isError bool))
	}
	if setter, ok := overview.(statusCallbackSetter); ok {
		setter.SetStatusCallback(func(msg string, isError bool) {
			a.layout.SetMessage(msg, isError)
		})
	}
	if setter, ok := settings.(statusCallbackSetter); ok {
		setter.SetStatusCallback(func(msg string, isError bool) {
			a.layout.SetMessage(msg, isError)
		})
	}

	// Setup app reference for settings view
	if appSetter, ok := settings.(interface{ SetApp(*tview.Application) }); ok {
		appSetter.SetApp(a.tview)
	}

	// Setup close handler for help view
	if helpView, ok := help.(interface{ SetCloseHandler(func()) }); ok {
		helpView.SetCloseHandler(func() {
			a.SwitchPage("overview")
		})
	}
}

// SwitchPage switches to the specified page
func (a *App) SwitchPage(name string) {
	var view View
	switch name {
	case "overview":
		view = a.overview
	case "settings":
		view = a.settings
	case "help":
		view = a.help
	default:
		return
	}

	a.currentPage = name
	a.mainPages.SwitchToPage(name)
	a.layout.SetActiveView(name)
	a.tview.SetFocus(view.Primitive())
	view.Focus()

	// Refresh the view (except help)
	if name != "help" {
		go func() {
			ctx := context.Background()
			if err := view.Refresh(ctx, a.client); err != nil {
				a.tview.QueueUpdateDraw(func() {
					a.layout.SetMessage(err.Error(), true)
				})
			} else {
				a.tview.QueueUpdateDraw(func() {
					a.layout.RefreshTimestamp()
				})
			}
		}()
	}
}

// RefreshCurrentView refreshes the currently active view with fresh data
func (a *App) RefreshCurrentView() {
	var view View
	switch a.currentPage {
	case "overview":
		view = a.overview
	case "settings":
		view = a.settings
	case "help":
		return // help doesn't need refresh
	default:
		return
	}

	go func() {
		ctx := context.Background()
		if err := view.Refresh(ctx, a.client); err != nil {
			a.tview.QueueUpdateDraw(func() {
				a.layout.SetMessage(err.Error(), true)
			})
		} else {
			a.tview.QueueUpdateDraw(func() {
				a.layout.SetMessage("Auto-refreshed", false)
				a.layout.RefreshTimestamp()
			})
		}
	}()
}

// Run starts the TUI application
func (a *App) Run() error {
	if a.overview == nil || a.settings == nil || a.help == nil {
		return fmt.Errorf("pages not configured")
	}

	// Set layout as root
	a.tview.SetRoot(a.layout.Root(), true)

	// Set initial context
	a.layout.SetContext("localhost:8080", "connecting...")

	// Set mainPages as content
	a.layout.SetContent(a.mainPages)
	a.layout.SetActiveView("overview")

	// Initial refresh for overview
	ctx := context.Background()
	if err := a.overview.Refresh(ctx, a.client); err != nil {
		a.layout.SetMessage(err.Error(), true)
		return fmt.Errorf("initial refresh: %w", err)
	}

	// Update context after successful connection
	a.layout.SetContext("localhost:8080", "connected")
	a.layout.SetMessage("Ready", false)

	// Set up key bindings
	a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// When settings view has a modal open, pass all keys through
		if a.currentPage == "settings" {
			type editChecker interface {
				IsEditing() bool
			}
			if sv, ok := a.settings.(editChecker); ok && sv.IsEditing() {
				return event
			}
		}

		switch event.Rune() {
		case 's':
			// Switch to settings (unless already in settings)
			if a.currentPage != "settings" {
				a.SwitchPage("settings")
				return nil
			}
			// If already in settings, pass through for save functionality
			return event
		case '?', 'h':
			a.SwitchPage("help")
			return nil
		case 'r':
			// Refresh current view
			a.layout.SetMessage("Refreshing...", false)
			a.RefreshCurrentView()
			a.tview.QueueUpdateDraw(func() {
				a.layout.SetMessage("Refreshed", false)
				a.layout.RefreshTimestamp()
			})
			return nil
		case '1', '2', '3', '4':
			// Category quick jump (only works in overview page)
			if a.currentPage == "overview" {
				// Pass through to UnifiedOverview
				return event
			}
			return nil
		case 'q':
			a.tview.Stop()
			return nil
		}

		// Esc key handling
		if event.Key() == tcell.KeyEscape {
			// In settings (not editing), return to overview
			if a.currentPage == "settings" {
				type editChecker interface {
					IsEditing() bool
				}
				if sv, ok := a.settings.(editChecker); ok && !sv.IsEditing() {
					a.SwitchPage("overview")
					return nil
				}
			}
		}

		return event
	})

	// Run the application
	return a.tview.Run()
}
