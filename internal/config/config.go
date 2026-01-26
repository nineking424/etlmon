package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for etlmon
type Config struct {
	Interval     time.Duration  `yaml:"-"` // Parsed from IntervalStr
	IntervalStr  string         `yaml:"interval"`
	Resources    []string       `yaml:"resources"`
	Windows      []string       `yaml:"windows"`
	Aggregations []string       `yaml:"aggregations"`
	Database     DatabaseConfig `yaml:"database"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// Valid resources, windows, and aggregations
var (
	ValidResources = map[string]bool{
		"cpu":    true,
		"memory": true,
		"disk":   true,
	}
	
	ValidAggregations = map[string]bool{
		"avg":  true,
		"max":  true,
		"min":  true,
		"last": true,
	}
)

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}
	
	// Parse interval string to duration
	if cfg.IntervalStr != "" {
		interval, err := time.ParseDuration(cfg.IntervalStr)
		if err != nil {
			return nil, fmt.Errorf("parsing interval %q: %w", cfg.IntervalStr, err)
		}
		cfg.Interval = interval
	}
	
	// Set default database path if not specified
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./etlmon.db"
	}
	
	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	
	if len(c.Resources) == 0 {
		return fmt.Errorf("at least one resource must be specified")
	}
	
	for _, r := range c.Resources {
		if !ValidResources[r] {
			return fmt.Errorf("invalid resource: %s (valid: cpu, memory, disk)", r)
		}
	}
	
	if len(c.Windows) == 0 {
		return fmt.Errorf("at least one window must be specified")
	}
	
	for _, w := range c.Windows {
		if _, err := ParseWindow(w); err != nil {
			return fmt.Errorf("invalid window %q: %w", w, err)
		}
	}
	
	if len(c.Aggregations) == 0 {
		return fmt.Errorf("at least one aggregation must be specified")
	}
	
	for _, a := range c.Aggregations {
		if !ValidAggregations[a] {
			return fmt.Errorf("invalid aggregation: %s (valid: avg, max, min, last)", a)
		}
	}
	
	return nil
}

// ParseWindow parses a window string like "1m", "5m", "1h" into a duration
func ParseWindow(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty window string")
	}
	
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid window format: %w", err)
	}
	
	if d <= 0 {
		return 0, fmt.Errorf("window must be positive")
	}
	
	return d, nil
}

// GetWindowDurations returns all windows as time.Duration
func (c *Config) GetWindowDurations() ([]time.Duration, error) {
	durations := make([]time.Duration, len(c.Windows))
	for i, w := range c.Windows {
		d, err := ParseWindow(w)
		if err != nil {
			return nil, err
		}
		durations[i] = d
	}
	return durations, nil
}
