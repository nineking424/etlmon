<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# internal

## Purpose

Private packages for etlmon implementation. Contains the core business logic:
- Data collection subsystem
- API Gateway
- Database layer
- Configuration management
- Command execution

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `collector/` | Data collection modules (see `collector/AGENTS.md`) |
| `api/` | HTTP API Gateway (see `api/AGENTS.md`) |
| `db/` | SQLite database layer (see `db/AGENTS.md`) |
| `config/` | Configuration parsing (see `config/AGENTS.md`) |
| `controller/` | Command execution (see `controller/AGENTS.md`) |

## For AI Agents

### Working In This Directory

- All packages here are **private** to etlmon
- Use clear interfaces between packages
- Each package should have a single responsibility
- Prefer composition over inheritance

### Package Dependencies (Allowed)

```
collector/* → db/repository (write)
api/handler → db/repository (read)
api/handler → controller (commands)
controller → collector/* (trigger scans)
config → (no internal deps, only stdlib + yaml)
```

### Common Patterns

1. **Interface-first design**: Define interfaces in the package that uses them
2. **Constructor functions**: `NewXxx(deps) *Xxx`
3. **Context propagation**: All functions accept `context.Context` as first param
4. **Error wrapping**: `fmt.Errorf("package.Function: %w", err)`

<!-- MANUAL: -->
