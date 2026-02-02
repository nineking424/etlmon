package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadNodeConfig_ValidFile_LoadsAllFields(t *testing.T) {
	yamlContent := `
node:
  listen: "127.0.0.1:9090"
  node_name: "test-node"
  db_path: "/tmp/test.db"

refresh:
  disk: 30s
  default_path_scan: 120s
  process: 20s

paths:
  - path: "/data"
    scan_interval: 60s
    max_depth: 5
    exclude:
      - "*.tmp"
      - ".git"
    timeout: 10s
  - path: "/logs"
    scan_interval: 300s
    max_depth: 3
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "node.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadNodeConfig(configFile)
	if err != nil {
		t.Fatalf("LoadNodeConfig failed: %v", err)
	}

	// Verify node settings
	if cfg.Node.Listen != "127.0.0.1:9090" {
		t.Errorf("Expected listen '127.0.0.1:9090', got '%s'", cfg.Node.Listen)
	}
	if cfg.Node.NodeName != "test-node" {
		t.Errorf("Expected node_name 'test-node', got '%s'", cfg.Node.NodeName)
	}
	if cfg.Node.DBPath != "/tmp/test.db" {
		t.Errorf("Expected db_path '/tmp/test.db', got '%s'", cfg.Node.DBPath)
	}

	// Verify refresh settings
	if cfg.Refresh.Disk != 30*time.Second {
		t.Errorf("Expected disk refresh 30s, got %v", cfg.Refresh.Disk)
	}
	if cfg.Refresh.DefaultPathScan != 120*time.Second {
		t.Errorf("Expected default_path_scan 120s, got %v", cfg.Refresh.DefaultPathScan)
	}
	if cfg.Refresh.Process != 20*time.Second {
		t.Errorf("Expected process refresh 20s, got %v", cfg.Refresh.Process)
	}

	// Verify paths
	if len(cfg.Paths) != 2 {
		t.Fatalf("Expected 2 paths, got %d", len(cfg.Paths))
	}

	path1 := cfg.Paths[0]
	if path1.Path != "/data" {
		t.Errorf("Expected path '/data', got '%s'", path1.Path)
	}
	if path1.ScanInterval != 60*time.Second {
		t.Errorf("Expected scan_interval 60s, got %v", path1.ScanInterval)
	}
	if path1.MaxDepth != 5 {
		t.Errorf("Expected max_depth 5, got %d", path1.MaxDepth)
	}
	if len(path1.Exclude) != 2 || path1.Exclude[0] != "*.tmp" || path1.Exclude[1] != ".git" {
		t.Errorf("Expected exclude ['*.tmp', '.git'], got %v", path1.Exclude)
	}
	if path1.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", path1.Timeout)
	}

	path2 := cfg.Paths[1]
	if path2.Path != "/logs" {
		t.Errorf("Expected path '/logs', got '%s'", path2.Path)
	}
	if path2.ScanInterval != 300*time.Second {
		t.Errorf("Expected scan_interval 300s, got %v", path2.ScanInterval)
	}
	if path2.MaxDepth != 3 {
		t.Errorf("Expected max_depth 3, got %d", path2.MaxDepth)
	}
}

func TestLoadNodeConfig_AppliesDefaults(t *testing.T) {
	yamlContent := `
node:
  node_name: "minimal-node"

paths:
  - path: "/data"
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "node.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadNodeConfig(configFile)
	if err != nil {
		t.Fatalf("LoadNodeConfig failed: %v", err)
	}

	// Verify defaults applied
	if cfg.Node.Listen != "0.0.0.0:8080" {
		t.Errorf("Expected default listen '0.0.0.0:8080', got '%s'", cfg.Node.Listen)
	}
	if cfg.Node.DBPath != "./etlmon.db" {
		t.Errorf("Expected default db_path './etlmon.db', got '%s'", cfg.Node.DBPath)
	}
	if cfg.Refresh.Disk != 15*time.Second {
		t.Errorf("Expected default disk refresh 15s, got %v", cfg.Refresh.Disk)
	}
	if cfg.Refresh.DefaultPathScan != 60*time.Second {
		t.Errorf("Expected default path scan 60s, got %v", cfg.Refresh.DefaultPathScan)
	}
	if cfg.Refresh.Process != 10*time.Second {
		t.Errorf("Expected default process refresh 10s, got %v", cfg.Refresh.Process)
	}

	// Verify path defaults
	if len(cfg.Paths) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(cfg.Paths))
	}
	path := cfg.Paths[0]
	if path.ScanInterval != 60*time.Second {
		t.Errorf("Expected default scan_interval 60s, got %v", path.ScanInterval)
	}
	if path.MaxDepth != 10 {
		t.Errorf("Expected default max_depth 10, got %d", path.MaxDepth)
	}
	if path.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", path.Timeout)
	}
}

func TestLoadNodeConfig_InvalidYAML_ReturnsError(t *testing.T) {
	yamlContent := `
node:
  listen: "0.0.0.0:8080
  invalid yaml here
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "node.yaml")
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadNodeConfig(configFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestLoadNodeConfig_FileNotFound_ReturnsError(t *testing.T) {
	_, err := LoadNodeConfig("/nonexistent/path/node.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestValidateNodeConfig_MissingNodeName_ReturnsError(t *testing.T) {
	cfg := &NodeConfig{
		Node: NodeSettings{
			Listen: "0.0.0.0:8080",
			DBPath: "./etlmon.db",
			// NodeName missing
		},
		Paths: []PathConfig{
			{Path: "/data"},
		},
	}

	err := ValidateNodeConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for missing node_name, got nil")
	}
	if err.Error() != "node_name is required" {
		t.Errorf("Expected 'node_name is required', got '%s'", err.Error())
	}
}

func TestValidateNodeConfig_NoPaths_ReturnsError(t *testing.T) {
	cfg := &NodeConfig{
		Node: NodeSettings{
			Listen:   "0.0.0.0:8080",
			NodeName: "test-node",
			DBPath:   "./etlmon.db",
		},
		Paths: []PathConfig{},
	}

	err := ValidateNodeConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for no paths, got nil")
	}
	if err.Error() != "at least one path must be configured" {
		t.Errorf("Expected 'at least one path must be configured', got '%s'", err.Error())
	}
}

func TestValidateNodeConfig_PathMissingPath_ReturnsError(t *testing.T) {
	cfg := &NodeConfig{
		Node: NodeSettings{
			Listen:   "0.0.0.0:8080",
			NodeName: "test-node",
			DBPath:   "./etlmon.db",
		},
		Paths: []PathConfig{
			{Path: ""}, // Missing path
		},
	}

	err := ValidateNodeConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for missing path in PathConfig, got nil")
	}
}

func TestValidateNodeConfig_ValidConfig_NoError(t *testing.T) {
	cfg := &NodeConfig{
		Node: NodeSettings{
			Listen:   "0.0.0.0:8080",
			NodeName: "test-node",
			DBPath:   "./etlmon.db",
		},
		Refresh: RefreshSettings{
			Disk:            15 * time.Second,
			DefaultPathScan: 60 * time.Second,
			Process:         10 * time.Second,
		},
		Paths: []PathConfig{
			{
				Path:         "/data",
				ScanInterval: 60 * time.Second,
				MaxDepth:     10,
				Timeout:      30 * time.Second,
			},
		},
	}

	err := ValidateNodeConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}
