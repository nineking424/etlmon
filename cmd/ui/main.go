package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/ui"
	"github.com/etlmon/etlmon/ui/client"
	"github.com/etlmon/etlmon/ui/views"
)

func main() {
	configPath := flag.String("c", "configs/ui.yaml", "path to config file")
	nodeOverride := flag.String("node", "", "node address override (e.g., http://localhost:8080)")
	flag.Parse()

	// Load config
	cfg, err := config.LoadUIConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Determine node address
	nodeAddr := getDefaultNodeAddress(cfg)
	if *nodeOverride != "" {
		nodeAddr = *nodeOverride
	}

	if nodeAddr == "" {
		fmt.Fprintln(os.Stderr, "Error: no node address configured. Use --node or configure in ui.yaml")
		os.Exit(1)
	}

	// Create HTTP client
	httpClient := client.NewClient(nodeAddr)

	// Create main UI app
	app := ui.NewApp(httpClient)

	// Create and register views (overview first = default view)
	overviewView := views.NewOverviewView()
	fsView := views.NewFSView()
	pathsView := views.NewPathsView()
	helpView := views.NewHelpView()

	app.AddView(overviewView)
	app.AddView(fsView)
	app.AddView(pathsView)
	app.AddView(helpView)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start refresh loop in background
	refreshInterval := cfg.UI.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				app.RefreshCurrentView()
			}
		}
	}()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
		// The app will exit when Run() returns
	}()

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}
}

// getDefaultNodeAddress returns the address of the default node from UIConfig
func getDefaultNodeAddress(cfg *config.UIConfig) string {
	if len(cfg.Nodes) == 0 {
		return ""
	}

	// If default node is specified, find it
	if cfg.UI.DefaultNode != "" {
		for _, node := range cfg.Nodes {
			if node.Name == cfg.UI.DefaultNode {
				return node.Address
			}
		}
	}

	// Otherwise, return the first node
	return cfg.Nodes[0].Address
}
