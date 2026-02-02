<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# config

## Purpose

Configuration parsing for both Node and UI components. Loads and validates YAML configuration files.

## Key Files

| File | Description |
|------|-------------|
| `config.go` | Main config types and loader |
| `node.go` | Node-specific configuration |
| `ui.go` | UI-specific configuration |
| `validate.go` | Configuration validation |

## For AI Agents

### Working In This Directory

- Use `gopkg.in/yaml.v3` for YAML parsing
- Provide sensible defaults for all optional fields
- Validate configuration on load
- Return clear error messages for invalid config

### Node Configuration

```go
type NodeConfig struct {
    Node    NodeSettings    `yaml:"node"`
    Refresh RefreshSettings `yaml:"refresh"`
    Paths   []PathConfig    `yaml:"paths"`
    Logs    []LogConfig     `yaml:"logs"`
    Process ProcessConfig   `yaml:"process_watch"`
    Cron    CronConfig      `yaml:"cron"`
    Xferlog XferlogConfig   `yaml:"xferlog"`
}

type NodeSettings struct {
    Listen   string `yaml:"listen"`   // default: "0.0.0.0:8080"
    NodeName string `yaml:"node_name"`
    DBPath   string `yaml:"db_path"`  // default: "/var/lib/etlmon/etlmon.db"
}

type RefreshSettings struct {
    Disk            time.Duration `yaml:"disk"`             // default: 15s
    DefaultPathScan time.Duration `yaml:"default_path_scan"` // default: 5m
    Process         time.Duration `yaml:"process"`          // default: 5s
}

type PathConfig struct {
    Path         string        `yaml:"path"`
    ScanInterval time.Duration `yaml:"scan_interval"`
    MaxDepth     int           `yaml:"max_depth"`
    Exclude      []string      `yaml:"exclude"`
    Timeout      time.Duration `yaml:"timeout"`
}
```

### UI Configuration

```go
type UIConfig struct {
    Nodes []NodeEntry `yaml:"nodes"`
    UI    UISettings  `yaml:"ui"`
}

type NodeEntry struct {
    Name    string `yaml:"name"`
    Address string `yaml:"address"`
}

type UISettings struct {
    RefreshInterval time.Duration `yaml:"refresh_interval"` // default: 2s
    DefaultNode     string        `yaml:"default_node"`
}
```

### Loading Pattern

```go
func LoadNodeConfig(path string) (*NodeConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    cfg := &NodeConfig{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    // Apply defaults
    applyNodeDefaults(cfg)

    // Validate
    if err := validateNodeConfig(cfg); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }

    return cfg, nil
}

func applyNodeDefaults(cfg *NodeConfig) {
    if cfg.Node.Listen == "" {
        cfg.Node.Listen = "0.0.0.0:8080"
    }
    if cfg.Node.DBPath == "" {
        cfg.Node.DBPath = "/var/lib/etlmon/etlmon.db"
    }
    if cfg.Refresh.Disk == 0 {
        cfg.Refresh.Disk = 15 * time.Second
    }
    // ...
}
```

### Validation

```go
func validateNodeConfig(cfg *NodeConfig) error {
    if cfg.Node.NodeName == "" {
        return errors.New("node.node_name is required")
    }

    for i, path := range cfg.Paths {
        if path.Path == "" {
            return fmt.Errorf("paths[%d].path is required", i)
        }
        if _, err := os.Stat(path.Path); err != nil {
            return fmt.Errorf("paths[%d].path does not exist: %s", i, path.Path)
        }
    }

    return nil
}
```

<!-- MANUAL: -->
