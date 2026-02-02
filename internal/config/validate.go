package config

import "fmt"

// ValidateNodeConfig validates a node configuration
func ValidateNodeConfig(cfg *NodeConfig) error {
	// Validate node settings
	if cfg.Node.NodeName == "" {
		return fmt.Errorf("node_name is required")
	}

	// Validate paths
	if len(cfg.Paths) == 0 {
		return fmt.Errorf("at least one path must be configured")
	}

	for i, path := range cfg.Paths {
		if path.Path == "" {
			return fmt.Errorf("path[%d]: path is required", i)
		}
	}

	return nil
}

// ValidateUIConfig validates a UI configuration
func ValidateUIConfig(cfg *UIConfig) error {
	// Validate nodes
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("at least one node must be configured")
	}

	for i, node := range cfg.Nodes {
		if node.Name == "" {
			return fmt.Errorf("nodes[%d]: name is required", i)
		}
		if node.Address == "" {
			return fmt.Errorf("nodes[%d]: address is required", i)
		}
	}

	return nil
}
