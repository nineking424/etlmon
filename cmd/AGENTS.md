<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# cmd

## Purpose

Entry points for the two executables that make up etlmon:
- `node/` - The daemon that runs on monitored servers
- `ui/` - The TUI client for operators

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `node/` | Node daemon entry point (see `node/AGENTS.md`) |
| `ui/` | TUI client entry point (see `ui/AGENTS.md`) |

## For AI Agents

### Working In This Directory

- Each subdirectory contains a single `main.go`
- Keep main.go minimal - delegate to internal packages
- Use cobra or similar for CLI argument parsing
- Both binaries should support `-c` flag for config file path

### Build Commands

```bash
# Build node daemon
go build -o bin/etlmon-node ./cmd/node

# Build UI client
go build -o bin/etlmon-ui ./cmd/ui

# Build both
make build
```

### Common Patterns

```go
// main.go structure
package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    configPath := flag.String("c", "", "config file path")
    flag.Parse()

    // Load config, initialize, run
    // Handle graceful shutdown via signal
}
```

<!-- MANUAL: -->
