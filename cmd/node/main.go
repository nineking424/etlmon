package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etlmon/etlmon/internal/api"
	"github.com/etlmon/etlmon/internal/collector/disk"
	"github.com/etlmon/etlmon/internal/collector/path"
	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/db"
	"github.com/etlmon/etlmon/internal/db/repository"
)

func main() {
	configPath := flag.String("c", "configs/node.yaml", "path to config file")
	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.LoadNodeConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting etlmon node", "name", cfg.Node.NodeName)

	// Initialize database
	database, err := db.NewDB(cfg.Node.DBPath)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create repository
	repo := repository.NewRepository(database.GetDB())

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start disk collector
	diskCollector := disk.NewDiskCollector(repo.FS, cfg.Refresh.Disk)
	if err := diskCollector.Start(ctx); err != nil {
		slog.Error("failed to start disk collector", "error", err)
		os.Exit(1)
	}
	slog.Info("disk collector started", "interval", cfg.Refresh.Disk)

	// Convert config.PathConfig to path.PathConfig
	pathConfigs := make([]path.PathConfig, len(cfg.Paths))
	for i, p := range cfg.Paths {
		pathConfigs[i] = path.PathConfig{
			Path:         p.Path,
			ScanInterval: p.ScanInterval,
			MaxDepth:     p.MaxDepth,
			Exclude:      p.Exclude,
			Timeout:      p.Timeout,
		}
	}

	// Start path scanner
	pathScanner := path.NewPathScanner(repo.Paths, pathConfigs)
	pathScanner.Start(ctx)
	slog.Info("path scanner started", "paths", len(cfg.Paths))

	// Create and start API server
	server := api.NewServer(cfg.Node.Listen, repo, cfg.Node.NodeName)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("starting API server", "address", cfg.Node.Listen)
		if err := server.Start(); err != nil {
			slog.Error("server error", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	slog.Info("received shutdown signal", "signal", sig)

	// Graceful shutdown
	cancel() // Stop collectors

	// Stop collectors explicitly
	diskCollector.Stop()
	pathScanner.Stop()

	// Shutdown API server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("etlmon node stopped")
}
