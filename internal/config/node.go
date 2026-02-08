package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// NodeConfig represents the complete node configuration
type NodeConfig struct {
	Node    NodeSettings       `yaml:"node"`
	Refresh RefreshSettings    `yaml:"refresh"`
	Paths   []PathConfig       `yaml:"paths"`
	Process ProcessConfig      `yaml:"process"`
	Logs    []LogMonitorConfig `yaml:"logs"`
}

// LoadNodeConfig loads and validates a node configuration from a YAML file
func LoadNodeConfig(path string) (*NodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg NodeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply defaults
	applyNodeDefaults(&cfg)

	// Validate configuration
	if err := ValidateNodeConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyNodeDefaults applies default values to unset fields
func applyNodeDefaults(cfg *NodeConfig) {
	// Node defaults
	if cfg.Node.Listen == "" {
		cfg.Node.Listen = "0.0.0.0:8080"
	}
	if cfg.Node.DBPath == "" {
		cfg.Node.DBPath = "./etlmon.db"
	}

	// Refresh defaults
	if cfg.Refresh.Disk == 0 {
		cfg.Refresh.Disk = 15 * time.Second
	}
	if cfg.Refresh.DefaultPathScan == 0 {
		cfg.Refresh.DefaultPathScan = 60 * time.Second
	}
	if cfg.Refresh.Process == 0 {
		cfg.Refresh.Process = 10 * time.Second
	}
	if cfg.Refresh.Log == 0 {
		cfg.Refresh.Log = 2 * time.Second
	}

	// Path defaults
	for i := range cfg.Paths {
		path := &cfg.Paths[i]
		if path.ScanInterval == 0 {
			path.ScanInterval = cfg.Refresh.DefaultPathScan
		}
		if path.MaxDepth == 0 {
			path.MaxDepth = 10
		}
		if path.Timeout == 0 {
			path.Timeout = 30 * time.Second
		}
	}

	// Process defaults
	if cfg.Process.TopN == 0 {
		cfg.Process.TopN = 50
	}

	// Log defaults
	for i := range cfg.Logs {
		if cfg.Logs[i].MaxLines == 0 {
			cfg.Logs[i].MaxLines = 1000
		}
	}
}
