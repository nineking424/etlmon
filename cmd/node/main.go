package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/etlmon/etlmon/internal/api"
	"github.com/etlmon/etlmon/internal/collector/disk"
	logcollector "github.com/etlmon/etlmon/internal/collector/log"
	"github.com/etlmon/etlmon/internal/collector/path"
	"github.com/etlmon/etlmon/internal/collector/process"
	"github.com/etlmon/etlmon/internal/config"
	"github.com/etlmon/etlmon/internal/db"
	"github.com/etlmon/etlmon/internal/db/repository"
)

// collectorManager manages the lifecycle of all collectors
type collectorManager struct {
	repo             *repository.Repository
	parentCtx        context.Context
	mu               sync.Mutex
	diskCollector    *disk.DiskCollector
	pathScanner      *path.PathScanner
	processCollector *process.Collector
	logTailer        *logcollector.LogTailer
}

func newCollectorManager(repo *repository.Repository, parentCtx context.Context) *collectorManager {
	return &collectorManager{
		repo:      repo,
		parentCtx: parentCtx,
	}
}

func (m *collectorManager) startAll(cfg *config.NodeConfig) error {
	// Start disk collector (static, not affected by config changes)
	m.diskCollector = disk.NewDiskCollector(m.repo.FS, cfg.Refresh.Disk)
	if err := m.diskCollector.Start(m.parentCtx); err != nil {
		return fmt.Errorf("failed to start disk collector: %w", err)
	}
	slog.Info("disk collector started", "interval", cfg.Refresh.Disk)

	return m.startDynamic(cfg)
}

func (m *collectorManager) startDynamic(cfg *config.NodeConfig) error {
	// Path scanner
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
	m.pathScanner = path.NewPathScanner(m.repo.Paths, pathConfigs)
	m.pathScanner.Start(m.parentCtx)
	slog.Info("path scanner started", "paths", len(cfg.Paths))

	// Process collector
	procConfig := process.Config{
		Patterns: cfg.Process.Patterns,
		TopN:     cfg.Process.TopN,
	}
	m.processCollector = process.NewCollector(m.repo.Process, cfg.Refresh.Process, procConfig)
	if err := m.processCollector.Start(m.parentCtx); err != nil {
		return fmt.Errorf("failed to start process collector: %w", err)
	}
	slog.Info("process collector started", "interval", cfg.Refresh.Process)

	// Log tailer
	if len(cfg.Logs) > 0 {
		tailerConfigs := make([]logcollector.TailerConfig, len(cfg.Logs))
		for i, l := range cfg.Logs {
			tailerConfigs[i] = logcollector.TailerConfig{
				Name:     l.Name,
				Path:     l.Path,
				MaxLines: l.MaxLines,
			}
		}
		m.logTailer = logcollector.NewLogTailer(m.repo.Log, tailerConfigs, cfg.Refresh.Log)
		if err := m.logTailer.Start(m.parentCtx); err != nil {
			return fmt.Errorf("failed to start log tailer: %w", err)
		}
		slog.Info("log tailer started", "logs", len(cfg.Logs), "interval", cfg.Refresh.Log)
	}

	return nil
}

func (m *collectorManager) stopDynamic() {
	if m.processCollector != nil {
		m.processCollector.Stop()
	}
	if m.logTailer != nil {
		m.logTailer.Stop()
		m.logTailer = nil
	}
	if m.pathScanner != nil {
		m.pathScanner.Stop()
	}
}

func (m *collectorManager) reload(cfg *config.NodeConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Info("reloading collectors with new config")
	m.stopDynamic()
	return m.startDynamic(cfg)
}

func (m *collectorManager) stopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.diskCollector != nil {
		m.diskCollector.Stop()
	}
	m.stopDynamic()
}

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

	// Create and start collectors
	cm := newCollectorManager(repo, ctx)
	if err := cm.startAll(cfg); err != nil {
		slog.Error("failed to start collectors", "error", err)
		os.Exit(1)
	}

	// Create and start API server
	server := api.NewServer(cfg.Node.Listen, repo, cfg.Node.NodeName, *configPath)
	server.SetPathScanner(cm.pathScanner)

	// Set config reload callback
	server.SetConfigReloadCallback(func() {
		newCfg, err := config.LoadNodeConfig(*configPath)
		if err != nil {
			slog.Error("failed to reload config", "error", err)
			return
		}
		if err := cm.reload(newCfg); err != nil {
			slog.Error("failed to restart collectors after config reload", "error", err)
			return
		}
		// Update scanner proxy with new path scanner
		server.SetPathScanner(cm.pathScanner)
		slog.Info("config reloaded successfully")
	})

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
	cancel()
	cm.stopAll()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("etlmon node stopped")
}
