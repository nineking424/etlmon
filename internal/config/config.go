package config

import "time"

// Common configuration types shared between node and UI configs

// NodeSettings contains node-level configuration
type NodeSettings struct {
	Listen   string `yaml:"listen"`
	NodeName string `yaml:"node_name"`
	DBPath   string `yaml:"db_path"`
}

// RefreshSettings contains refresh interval configuration
type RefreshSettings struct {
	Disk            time.Duration `yaml:"disk"`
	DefaultPathScan time.Duration `yaml:"default_path_scan"`
	Process         time.Duration `yaml:"process"`
}

// PathConfig defines a monitored path with its scan settings
type PathConfig struct {
	Path         string        `yaml:"path"`
	ScanInterval time.Duration `yaml:"scan_interval"`
	MaxDepth     int           `yaml:"max_depth"`
	Exclude      []string      `yaml:"exclude"`
	Timeout      time.Duration `yaml:"timeout"`
}

// NodeEntry represents a node in the UI configuration
type NodeEntry struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// UISettings contains UI-level configuration
type UISettings struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	DefaultNode     string        `yaml:"default_node"`
}
