<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# controller

## Purpose

Command execution for administrative actions. Handles process killing, scan triggering, and other control operations with safety checks.

## Key Files

| File | Description |
|------|-------------|
| `controller.go` | Main controller with all commands |
| `kill.go` | Process kill implementation |
| `scan.go` | Path scan trigger |

## For AI Agents

### Working In This Directory

- All commands require explicit confirmation from caller
- Validate permissions before executing
- Log all command executions
- Return detailed error messages

### Controller Interface

```go
type Controller struct {
    pathScanner *path.Scanner
    logger      *log.Logger
}

func NewController(scanner *path.Scanner) *Controller {
    return &Controller{
        pathScanner: scanner,
        logger:      log.New(os.Stdout, "[controller] ", log.LstdFlags),
    }
}
```

### Kill Process

```go
func (c *Controller) KillProcess(ctx context.Context, pid int, signal string) error {
    // Validate PID exists
    proc, err := os.FindProcess(pid)
    if err != nil {
        return fmt.Errorf("process not found: %d", pid)
    }

    // Map signal string to syscall
    var sig syscall.Signal
    switch signal {
    case "SIGTERM", "":
        sig = syscall.SIGTERM
    case "SIGKILL":
        sig = syscall.SIGKILL
    default:
        return fmt.Errorf("unsupported signal: %s", signal)
    }

    c.logger.Printf("killing pid=%d signal=%s", pid, signal)

    if err := proc.Signal(sig); err != nil {
        return fmt.Errorf("kill failed: %w", err)
    }

    return nil
}
```

### Trigger Scan

```go
func (c *Controller) TriggerScan(ctx context.Context, paths []string) error {
    if len(paths) == 0 {
        // Scan all configured paths
        return c.pathScanner.ScanAll(ctx)
    }

    // Scan specific paths
    for _, p := range paths {
        if err := c.pathScanner.ScanPath(ctx, p); err != nil {
            c.logger.Printf("scan failed for %s: %v", p, err)
        }
    }
    return nil
}
```

### Safety Considerations

1. **Process Kill**:
   - Only allow killing processes that are in the watch list
   - Log all kill attempts with user/timestamp
   - Consider rate limiting kill requests

2. **Scan Trigger**:
   - Don't allow triggering if scan already in progress
   - Respect configured timeouts
   - Log all manual scan triggers

```go
// Check if process is in watch list
func (c *Controller) isWatched(pid int) bool {
    for _, watched := range c.watchedPIDs {
        if watched == pid {
            return true
        }
    }
    return false
}

func (c *Controller) KillProcess(ctx context.Context, pid int, signal string) error {
    if !c.isWatched(pid) {
        return fmt.Errorf("pid %d is not in watch list", pid)
    }
    // ... proceed with kill
}
```

<!-- MANUAL: -->
