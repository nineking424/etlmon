package ui

import (
	"context"
	"fmt"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/rivo/tview"
)

// App represents the main TUI application
type App struct {
	tview   *tview.Application
	client  *client.Client
	views   map[string]View
	current string
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
}

// SwitchView switches to the specified view
func (a *App) SwitchView(name string) {
	if view, ok := a.views[name]; ok {
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

	// Run the application
	return a.tview.Run()
}
