package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etlmon/etlmon/internal/aggregator"
	"github.com/etlmon/etlmon/internal/collector"
	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/storage"
	"github.com/etlmon/etlmon/internal/tui"
)

var (
	version = "0.1.0"
	commit  = "dev"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "", "Path to configuration file (required)")
	dbPath := flag.String("db", "", "Override database path from config")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("etlmon version %s (%s)\n", version, commit)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --config flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Override database path if specified
	if *dbPath != "" {
		cfg.Database.Path = *dbPath
	}

	// Initialize storage
	store, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Get window durations
	windows, err := cfg.GetWindowDurations()
	if err != nil {
		log.Fatalf("Failed to parse windows: %v", err)
	}

	// Create aggregator
	agg := aggregator.NewAggregator(windows, cfg.Aggregations)

	// Create collector manager
	collectorMgr := collector.NewManager(cfg.Interval)

	// Register collectors based on config
	for _, resource := range cfg.Resources {
		switch resource {
		case "cpu":
			collectorMgr.Register(collector.NewCPUCollector())
		case "memory":
			collectorMgr.Register(collector.NewMemoryCollector())
		case "disk":
			collectorMgr.Register(collector.NewDiskCollector())
		}
	}

	// Create TUI
	app := tui.NewApp()
	app.SetStore(store)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for metrics from collector
	metricsChan := make(chan []collector.Metric, 100)

	// Start collector in goroutine
	go func() {
		collectorMgr.Start(ctx, func(metrics []collector.Metric) {
			select {
			case metricsChan <- metrics:
			default:
				// Drop metrics if channel is full (backpressure)
			}
		})
	}()

	// Start aggregation checker in goroutine
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case metrics := <-metricsChan:
				// Add metrics to aggregator
				for _, m := range metrics {
					agg.Add(m)
				}

				// Update TUI with realtime metrics
				app.QueueUpdateDraw(func() {
					app.OnMetricsCollected(metrics)
				})

			case <-ticker.C:
				// Check for completed windows
				results := agg.CheckWindows(time.Now())
				if len(results) > 0 {
					// Save to storage
					batch := make([]*storage.AggregatedMetric, len(results))
					for i, r := range results {
						batch[i] = &storage.AggregatedMetric{
							Timestamp:       r.Timestamp.Unix(),
							ResourceType:    r.ResourceType,
							MetricName:      r.MetricName,
							AggregatedValue: r.Value,
							WindowSize:      tui.FormatDuration(r.WindowSize),
							AggregationType: r.AggregationType,
						}
					}

					if err := store.SaveBatch(batch); err != nil {
						log.Printf("Failed to save metrics: %v", err)
					}

					// Update TUI with aggregation results
					app.QueueUpdateDraw(func() {
						app.OnAggregationComplete(results)
					})
				}
			}
		}
	}()

	// Handle shutdown signal
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		app.Stop()
	}()

	// Run TUI (blocking)
	if err := app.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}

	log.Println("etlmon stopped")
}
