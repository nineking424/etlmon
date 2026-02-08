package ui

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/layout"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main TUI application
type App struct {
	tview    *tview.Application
	client   *client.Client
	layout   *layout.Layout
	views    map[string]View
	current  string
	previous string
}

// NewApp creates a new TUI application
func NewApp(client *client.Client) *App {
	return &App{
		tview:  tview.NewApplication(),
		client: client,
		layout: layout.NewLayout(),
		views:  make(map[string]View),
	}
}

// AddView registers a view with the application
func (a *App) AddView(v View) {
	a.views[v.Name()] = v
	if a.current == "" {
		a.current = v.Name()
	}

	// Set status callback if view supports it
	type statusCallbackSetter interface {
		SetStatusCallback(func(msg string, isError bool))
	}
	if setter, ok := v.(statusCallbackSetter); ok {
		setter.SetStatusCallback(func(msg string, isError bool) {
			a.layout.SetMessage(msg, isError)
			a.tview.Draw()
		})
	}

	// Setup close handler for help view
	if v.Name() == "help" {
		if helpView, ok := v.(interface{ SetCloseHandler(func()) }); ok {
			helpView.SetCloseHandler(func() {
				// Return to previous view
				if a.previous != "" && a.previous != "help" {
					a.SwitchView(a.previous)
				} else {
					// If no previous view, go to overview
					a.SwitchView("overview")
				}
			})
		}
	}
}

// SwitchView switches to the specified view
func (a *App) SwitchView(name string) {
	if view, ok := a.views[name]; ok {
		a.previous = a.current
		a.current = name

		// Use layout for content instead of setting view as root
		a.layout.SetContent(view.Primitive())
		a.layout.SetActiveView(name)
		view.Focus()

		// Refresh the view
		go func() {
			ctx := context.Background()
			view.Refresh(ctx, a.client)
			a.tview.Draw()
		}()
	}
}

// RefreshCurrentView refreshes the currently active view with fresh data
func (a *App) RefreshCurrentView() {
	a.tview.QueueUpdateDraw(func() {
		if view, ok := a.views[a.current]; ok {
			ctx := context.Background()
			if err := view.Refresh(ctx, a.client); err != nil {
				a.layout.SetMessage(err.Error(), true)
			} else {
				a.layout.SetMessage("Auto-refreshed", false)
			}
			a.layout.RefreshTimestamp()
		}
	})
}

// Run starts the TUI application
func (a *App) Run() error {
	if a.current == "" {
		return fmt.Errorf("no views registered")
	}

	// Set layout as root
	a.tview.SetRoot(a.layout.Root(), true)

	// Set initial context
	a.layout.SetContext("localhost:8080", "connecting...")

	// Set initial view
	view := a.views[a.current]
	a.layout.SetContent(view.Primitive())
	a.layout.SetActiveView(a.current)

	// Initial refresh
	ctx := context.Background()
	if err := view.Refresh(ctx, a.client); err != nil {
		a.layout.SetMessage(err.Error(), true)
		return fmt.Errorf("initial refresh: %w", err)
	}

	// Update context after successful connection
	a.layout.SetContext("localhost:8080", "connected")
	a.layout.SetMessage("Ready", false)

	// Set up key bindings
	a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '0':
			a.SwitchView("overview")
			return nil
		case '1':
			a.SwitchView("fs")
			return nil
		case '2':
			a.SwitchView("paths")
			return nil
		case '?', 'h':
			a.SwitchView("help")
			return nil
		case 'r':
			// Refresh current view
			if view, ok := a.views[a.current]; ok {
				a.layout.SetMessage("Refreshing...", false)
				a.tview.Draw()
				go func() {
					ctx := context.Background()
					if err := view.Refresh(ctx, a.client); err != nil {
						a.layout.SetMessage(err.Error(), true)
					} else {
						a.layout.SetMessage("Refreshed", false)
					}
					a.layout.RefreshTimestamp()
					a.tview.Draw()
				}()
			}
			return nil
		case 's':
			// Trigger scan (only works in paths view)
			if a.current == "paths" {
				type scanTrigger interface {
					TriggerScan(context.Context, *client.Client) error
				}
				if pathsView, ok := a.views["paths"].(scanTrigger); ok {
					a.layout.SetMessage("Scanning...", false)
					go func() {
						ctx := context.Background()
						if err := pathsView.TriggerScan(ctx, a.client); err != nil {
							a.layout.SetMessage(err.Error(), true)
						} else {
							a.layout.SetMessage("Scan complete", false)
						}
						// Refresh the view after triggering scan
						if view, ok := a.views["paths"]; ok {
							view.Refresh(ctx, a.client)
						}
						a.tview.Draw()
					}()
				}
			}
			return nil
		case 'q':
			// In help view, let the view handle it to return to previous view
			if a.current == "help" {
				return event
			}
			a.tview.Stop()
			return nil
		}
		return event
	})

	// Run the application
	return a.tview.Run()
}
