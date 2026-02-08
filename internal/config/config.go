package config

import "time"

// Common configuration types shared between node and UI configs

// NodeSettings contains node-level configuration
type NodeSettings struct {
	Listen   string `yaml:"listen" json:"listen"`
	NodeName string `yaml:"node_name" json:"node_name"`
	DBPath   string `yaml:"db_path" json:"db_path"`
}

// RefreshSettings contains refresh interval configuration
type RefreshSettings struct {
	Disk            time.Duration `yaml:"disk" json:"disk"`
	DefaultPathScan time.Duration `yaml:"default_path_scan" json:"default_path_scan"`
	Process         time.Duration `yaml:"process" json:"process"`
	Log             time.Duration `yaml:"log" json:"log"`
}

// PathConfig defines a monitored path with its scan settings
type PathConfig struct {
	Path         string        `yaml:"path" json:"path"`
	ScanInterval time.Duration `yaml:"scan_interval" json:"scan_interval"`
	MaxDepth     int           `yaml:"max_depth" json:"max_depth"`
	Exclude      []string      `yaml:"exclude" json:"exclude"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`
}

// ProcessConfig defines process monitoring settings
type ProcessConfig struct {
	Patterns []string `yaml:"patterns" json:"patterns"`
	TopN     int      `yaml:"top_n" json:"top_n"`
}

// LogMonitorConfig defines a single log file to monitor
type LogMonitorConfig struct {
	Name     string `yaml:"name" json:"name"`
	Path     string `yaml:"path" json:"path"`
	MaxLines int    `yaml:"max_lines" json:"max_lines"`
}

// NodeEntry represents a node in the UI configuration
type NodeEntry struct {
	Name    string `yaml:"name" json:"name"`
	Address string `yaml:"address" json:"address"`
}

// UISettings contains UI-level configuration
type UISettings struct {
	RefreshInterval time.Duration `yaml:"refresh_interval" json:"refresh_interval"`
	DefaultNode     string        `yaml:"default_node" json:"default_node"`
}
