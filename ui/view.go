package ui

import (
	"context"

	"github.com/etlmon/etlmon/ui/client"
	"github.com/rivo/tview"
)

// View represents a TUI view component
type View interface {
	// Name returns the unique name of the view
	Name() string

	// Primitive returns the tview primitive for rendering
	Primitive() tview.Primitive

	// Refresh updates the view with fresh data from the API
	Refresh(ctx context.Context, client *client.Client) error

	// Focus sets focus on the appropriate element
	Focus()
}
