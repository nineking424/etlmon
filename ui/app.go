package ui

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main TUI application
type App struct {
	tview    *tview.Application
	client   *client.Client
	views    map[string]View
	current  string
	previous string
}

// NewApp creates a new TUI application
func NewApp(client *client.Client) *App {
	return &App{
		tview:  tview.NewApplication(),
		client: client,
		views:  make(map[string]View),
	}
}

// AddView registers a view with the application
func (a *App) AddView(v View) {
	a.views[v.Name()] = v
	if a.current == "" {
		a.current = v.Name()
	}

	// Setup close handler for help view
	if v.Name() == "help" {
		if helpView, ok := v.(interface{ SetCloseHandler(func()) }); ok {
			helpView.SetCloseHandler(func() {
				// Return to previous view
				if a.previous != "" && a.previous != "help" {
					a.SwitchView(a.previous)
				} else {
					// If no previous view, go to fs view
					a.SwitchView("fs")
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
		a.tview.SetRoot(view.Primitive(), true)
		view.Focus()

		// Refresh the view
		go func() {
			ctx := context.Background()
			view.Refresh(ctx, a.client)
			a.tview.Draw()
		}()
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	if a.current == "" {
		return fmt.Errorf("no views registered")
	}

	// Set initial view
	view := a.views[a.current]
	a.tview.SetRoot(view.Primitive(), true)

	// Initial refresh
	ctx := context.Background()
	if err := view.Refresh(ctx, a.client); err != nil {
		return fmt.Errorf("initial refresh: %w", err)
	}

	// Set up key bindings
	a.tview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
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
				go func() {
					ctx := context.Background()
					view.Refresh(ctx, a.client)
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
					go func() {
						ctx := context.Background()
						if err := pathsView.TriggerScan(ctx, a.client); err != nil {
							// Silently ignore errors for now
							// TODO: Add status bar to display errors
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
