package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUIConfig_ValidFile_LoadsAllFields(t *testing.T) {
	yamlContent := `
nodes:
  - name: "node1"
    address: "http://localhost:8080"
  - name: "node2"
    address: "http://192.168.1.10:8080"

ui:
  refresh_interval: 5s
  default_node: "node1"
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "ui.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadUIConfig(configFile)
	if err != nil {
		t.Fatalf("LoadUIConfig failed: %v", err)
	}

	// Verify nodes
	if len(cfg.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(cfg.Nodes))
	}

	node1 := cfg.Nodes[0]
	if node1.Name != "node1" {
		t.Errorf("Expected node name 'node1', got '%s'", node1.Name)
	}
	if node1.Address != "http://localhost:8080" {
		t.Errorf("Expected address 'http://localhost:8080', got '%s'", node1.Address)
	}

	node2 := cfg.Nodes[1]
	if node2.Name != "node2" {
		t.Errorf("Expected node name 'node2', got '%s'", node2.Name)
	}
	if node2.Address != "http://192.168.1.10:8080" {
		t.Errorf("Expected address 'http://192.168.1.10:8080', got '%s'", node2.Address)
	}

	// Verify UI settings
	if cfg.UI.RefreshInterval != 5*time.Second {
		t.Errorf("Expected refresh_interval 5s, got %v", cfg.UI.RefreshInterval)
	}
	if cfg.UI.DefaultNode != "node1" {
		t.Errorf("Expected default_node 'node1', got '%s'", cfg.UI.DefaultNode)
	}
}

func TestLoadUIConfig_AppliesDefaults(t *testing.T) {
	yamlContent := `
nodes:
  - name: "node1"
    address: "http://localhost:8080"
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "ui.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadUIConfig(configFile)
	if err != nil {
		t.Fatalf("LoadUIConfig failed: %v", err)
	}

	// Verify defaults applied
	if cfg.UI.RefreshInterval != 3*time.Second {
		t.Errorf("Expected default refresh_interval 3s, got %v", cfg.UI.RefreshInterval)
	}
	if cfg.UI.DefaultNode != "" {
		t.Errorf("Expected empty default_node, got '%s'", cfg.UI.DefaultNode)
	}
}

func TestLoadUIConfig_InvalidYAML_ReturnsError(t *testing.T) {
	yamlContent := `
nodes:
  - name: "node1
    invalid yaml
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "ui.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadUIConfig(configFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestLoadUIConfig_FileNotFound_ReturnsError(t *testing.T) {
	_, err := LoadUIConfig("/nonexistent/path/ui.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestValidateUIConfig_NoNodes_ReturnsError(t *testing.T) {
	cfg := &UIConfig{
		Nodes: []NodeEntry{},
		UI: UISettings{
			RefreshInterval: 3 * time.Second,
		},
	}

	err := ValidateUIConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for no nodes, got nil")
	}
	if err.Error() != "at least one node must be configured" {
		t.Errorf("Expected 'at least one node must be configured', got '%s'", err.Error())
	}
}

func TestValidateUIConfig_NodeMissingName_ReturnsError(t *testing.T) {
	cfg := &UIConfig{
		Nodes: []NodeEntry{
			{Name: "", Address: "http://localhost:8080"},
		},
		UI: UISettings{
			RefreshInterval: 3 * time.Second,
		},
	}

	err := ValidateUIConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for missing node name, got nil")
	}
}

func TestValidateUIConfig_NodeMissingAddress_ReturnsError(t *testing.T) {
	cfg := &UIConfig{
		Nodes: []NodeEntry{
			{Name: "node1", Address: ""},
		},
		UI: UISettings{
			RefreshInterval: 3 * time.Second,
		},
	}

	err := ValidateUIConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for missing node address, got nil")
	}
}

func TestValidateUIConfig_ValidConfig_NoError(t *testing.T) {
	cfg := &UIConfig{
		Nodes: []NodeEntry{
			{Name: "node1", Address: "http://localhost:8080"},
		},
		UI: UISettings{
			RefreshInterval: 3 * time.Second,
			DefaultNode:     "node1",
		},
	}

	err := ValidateUIConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}
