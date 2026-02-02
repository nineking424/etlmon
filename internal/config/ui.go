package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// UIConfig represents the complete UI configuration
type UIConfig struct {
	Nodes []NodeEntry `yaml:"nodes"`
	UI    UISettings  `yaml:"ui"`
}

// LoadUIConfig loads and validates a UI configuration from a YAML file
func LoadUIConfig(path string) (*UIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg UIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply defaults
	applyUIDefaults(&cfg)

	// Validate configuration
	if err := ValidateUIConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyUIDefaults applies default values to unset fields
func applyUIDefaults(cfg *UIConfig) {
	// UI defaults
	if cfg.UI.RefreshInterval == 0 {
		cfg.UI.RefreshInterval = 3 * time.Second
	}
}
