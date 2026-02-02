<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# docs

## Purpose

Documentation for etlmon including user guides, API reference, and operational runbooks.

## Key Files

| File | Description |
|------|-------------|
| `README.md` | Main documentation entry point |
| `INSTALL.md` | Installation guide |
| `API.md` | API reference |
| `OPERATIONS.md` | Operational runbook |

## For AI Agents

### Working In This Directory

- Write clear, concise documentation
- Include examples for all operations
- Keep API docs in sync with implementation
- Target audience: ETL operators and SREs

### Documentation Structure

```
docs/
├── README.md         # Overview and quick start
├── INSTALL.md        # Installation steps
├── API.md            # Full API reference
├── OPERATIONS.md     # Running in production
└── TROUBLESHOOTING.md # Common issues
```

### API.md Structure

```markdown
# etlmon API Reference

Base URL: `http://<node-address>:8080/api/v1`

## Filesystem

### GET /fs
Returns filesystem usage for all mounts.

**Response:**
\`\`\`json
{
  "data": [
    {
      "mount_point": "/data",
      "total_bytes": 1073741824,
      "used_bytes": 536870912,
      "avail_bytes": 536870912,
      "used_percent": 50.0,
      "collected_at": "2026-01-15T10:00:00Z"
    }
  ]
}
\`\`\`

## Paths
...
```

<!-- MANUAL: -->
