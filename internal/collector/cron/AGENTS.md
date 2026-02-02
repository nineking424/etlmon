<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# cron

## Purpose

Parses system and user crontabs, calculates next run times. Writes to `cron_jobs` table. Watches for crontab file changes.

## Key Files

| File | Description |
|------|-------------|
| `collector.go` | CronCollector implementation |
| `collector_test.go` | Unit tests |
| `parser.go` | Crontab file parsing |

## For AI Agents

### Working In This Directory

- Parse `/etc/crontab`, `/var/spool/cron/crontabs/*`
- Use `robfig/cron` for expression parsing and next_run calculation
- Watch crontab files with fsnotify for live updates
- Re-parse periodically as fallback

### Data Model

```go
type CronJob struct {
    JobID       int64
    Schedule    string // Cron expression
    Command     string
    User        string
    Source      string // "system" or "user"
    NextRun     time.Time
    LastChecked time.Time
}
```

### Crontab Locations

```go
var crontabPaths = []string{
    "/etc/crontab",           // System crontab
    "/etc/cron.d/*",          // System cron.d
    "/var/spool/cron/*",      // User crontabs (RedHat)
    "/var/spool/cron/crontabs/*", // User crontabs (Debian)
}
```

### Parsing Example

```go
import "github.com/robfig/cron/v3"

func parseCrontab(path string) ([]CronJob, error) {
    content, _ := os.ReadFile(path)
    var jobs []CronJob

    for _, line := range strings.Split(string(content), "\n") {
        // Skip comments and empty lines
        if strings.HasPrefix(line, "#") || line == "" {
            continue
        }

        // Parse: minute hour day month weekday [user] command
        parts := strings.Fields(line)
        schedule := strings.Join(parts[0:5], " ")

        parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
        sched, err := parser.Parse(schedule)
        if err != nil {
            continue
        }

        jobs = append(jobs, CronJob{
            Schedule: schedule,
            Command:  parts[len(parts)-1],
            NextRun:  sched.Next(time.Now()),
        })
    }
    return jobs, nil
}
```

<!-- MANUAL: -->
