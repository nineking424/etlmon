# Contributing to etlmon

Thank you for your interest in contributing to etlmon! This document outlines our development practices, with a strong emphasis on **Test-Driven Development (TDD)**.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [TDD Requirements](#tdd-requirements-mandatory)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Commit Messages](#commit-messages)

---

## Code of Conduct

Be respectful, constructive, and collaborative. We're all here to build great software.

---

## TDD Requirements (MANDATORY)

> **This project strictly follows Test-Driven Development. PRs without tests will not be merged.**

### The TDD Cycle

Every feature, bug fix, or enhancement must follow the **Red-Green-Refactor** cycle:

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                  │
│   1. RED        2. GREEN       3. REFACTOR                      │
│   ─────────     ──────────     ───────────                      │
│   Write a       Write the      Improve the                      │
│   failing       minimum code   code while                       │
│   test first    to pass        tests stay                       │
│                                green                            │
│                                                                  │
│   ┌─────┐       ┌─────┐        ┌─────┐                         │
│   │FAIL │  ───► │PASS │  ───►  │PASS │  ───► (repeat)          │
│   └─────┘       └─────┘        └─────┘                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### The Golden Rules

1. **Never write production code without a failing test**
2. **Write only one failing test at a time**
3. **Write only enough code to make the test pass**
4. **Refactor only when all tests are green**
5. **Every line of production code must exist because a test required it**

### Example TDD Workflow

Let's say you're adding a feature to exclude certain filesystem types from monitoring.

#### Step 1: Write the failing test (RED)

```go
// internal/collector/disk/collector_test.go

func TestDiskCollector_Collect_ExcludesConfiguredTypes(t *testing.T) {
    cfg := DiskConfig{
        ExcludeTypes: []string{"tmpfs", "devtmpfs"},
    }
    collector := NewDiskCollector(cfg)

    results, err := collector.Collect(context.Background())

    require.NoError(t, err)
    for _, r := range results {
        assert.NotContains(t, []string{"tmpfs", "devtmpfs"}, r.FSType,
            "excluded filesystem types should not appear in results")
    }
}
```

Run the test:
```bash
go test ./internal/collector/disk/ -run TestDiskCollector_Collect_ExcludesConfiguredTypes -v
```

The test **fails** because `ExcludeTypes` doesn't exist yet. This is correct!

#### Step 2: Write minimal code to pass (GREEN)

```go
// internal/collector/disk/config.go
type DiskConfig struct {
    Interval     time.Duration
    ExcludeTypes []string  // Add this
}

// internal/collector/disk/collector.go
func (c *DiskCollector) Collect(ctx context.Context) ([]*FilesystemUsage, error) {
    // ... existing code ...

    // Add exclusion logic
    if c.isExcluded(fsType) {
        continue
    }

    // ... rest of code ...
}

func (c *DiskCollector) isExcluded(fsType string) bool {
    for _, excluded := range c.cfg.ExcludeTypes {
        if fsType == excluded {
            return true
        }
    }
    return false
}
```

Run the test again:
```bash
go test ./internal/collector/disk/ -run TestDiskCollector_Collect_ExcludesConfiguredTypes -v
```

The test **passes**!

#### Step 3: Refactor if needed (REFACTOR)

Maybe `isExcluded` could be more efficient with a map lookup:

```go
type DiskCollector struct {
    cfg          DiskConfig
    excludeSet   map[string]bool  // Add lookup map
}

func NewDiskCollector(cfg DiskConfig) *DiskCollector {
    excludeSet := make(map[string]bool)
    for _, t := range cfg.ExcludeTypes {
        excludeSet[t] = true
    }
    return &DiskCollector{cfg: cfg, excludeSet: excludeSet}
}

func (c *DiskCollector) isExcluded(fsType string) bool {
    return c.excludeSet[fsType]
}
```

Run all tests to ensure nothing broke:
```bash
go test ./internal/collector/disk/ -v
```

All tests still **pass**. Refactoring complete!

#### Step 4: Add edge case tests

```go
func TestDiskCollector_Collect_EmptyExcludeList(t *testing.T) {
    cfg := DiskConfig{ExcludeTypes: []string{}}
    collector := NewDiskCollector(cfg)

    results, err := collector.Collect(context.Background())

    require.NoError(t, err)
    assert.NotEmpty(t, results, "should return results when no exclusions")
}

func TestDiskCollector_Collect_CaseSensitiveExclusion(t *testing.T) {
    cfg := DiskConfig{ExcludeTypes: []string{"tmpfs"}}
    collector := NewDiskCollector(cfg)

    // "TMPFS" should NOT be excluded (case sensitive)
    // Test implementation depends on your design decision
}
```

---

## Test File Structure

### Naming Convention

```
Production File              Test File
────────────────────────────────────────────
internal/collector/disk/
  collector.go           →   collector_test.go
  config.go              →   config_test.go

internal/api/handler/
  fs.go                  →   fs_test.go
  paths.go               →   paths_test.go

pkg/models/
  fs.go                  →   fs_test.go
```

### Test Function Naming

Use descriptive names that explain what's being tested:

```go
// Pattern: Test<Unit>_<Scenario>_<ExpectedBehavior>

// Good
func TestPathScanner_Scan_ReturnsFileCount(t *testing.T)
func TestPathScanner_Scan_RespectsMaxDepth(t *testing.T)
func TestPathScanner_Scan_SkipsExcludedPatterns(t *testing.T)
func TestPathScanner_Scan_TimesOutOnLongScan(t *testing.T)
func TestPathScanner_Scan_HandlesPermissionDenied(t *testing.T)

// Bad (too vague)
func TestScan(t *testing.T)
func TestPathScanner(t *testing.T)
func TestScanWorks(t *testing.T)
```

### Table-Driven Tests

For functions with multiple input/output cases, use table-driven tests:

```go
func TestFormatBytes(t *testing.T) {
    tests := []struct {
        name  string
        bytes uint64
        want  string
    }{
        {"zero", 0, "0 B"},
        {"bytes", 500, "500 B"},
        {"kilobytes", 1024, "1.0 KB"},
        {"megabytes", 1048576, "1.0 MB"},
        {"gigabytes", 1073741824, "1.0 GB"},
        {"terabytes", 1099511627776, "1.0 TB"},
        {"large", 5368709120, "5.0 GB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatBytes(tt.bytes)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

---

## Coverage Requirements

| Component | Minimum Coverage |
|-----------|------------------|
| `internal/collector/*` | 80% |
| `internal/db/repository` | 85% |
| `internal/api/handler` | 80% |
| `internal/config` | 90% |
| `pkg/models` | 70% |
| `ui/*` | 60% |
| **Overall** | **80%** |

Check coverage with:

```bash
# Quick coverage check
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

---

## Development Workflow

### 1. Fork and Clone

```bash
git clone https://github.com/<your-username>/etlmon.git
cd etlmon
```

### 2. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Write Tests First (TDD)

```bash
# Create test file
touch internal/collector/newfeature/feature_test.go

# Write your failing test
# Run it to confirm it fails
go test ./internal/collector/newfeature/ -v
```

### 4. Implement the Feature

```bash
# Write minimal code to pass the test
# Run tests again
go test ./internal/collector/newfeature/ -v
```

### 5. Run Full Test Suite

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...
```

### 6. Commit and Push

```bash
git add .
git commit -m "feat(collector): add filesystem type exclusion"
git push origin feature/your-feature-name
```

### 7. Open a Pull Request

---

## Pull Request Process

### PR Checklist

Before submitting a PR, ensure:

- [ ] **Tests written first** (TDD followed)
- [ ] **All tests pass** (`go test ./...`)
- [ ] **Race detection passes** (`go test -race ./...`)
- [ ] **Coverage maintained** (no decrease)
- [ ] **Code is formatted** (`gofmt -s -w .`)
- [ ] **Linter passes** (`golangci-lint run`)
- [ ] **Commit messages follow convention**

### PR Description Template

```markdown
## Summary

Brief description of what this PR does.

## TDD Evidence

- Test file(s) added: `internal/collector/disk/collector_test.go`
- New test functions:
  - `TestDiskCollector_Collect_ExcludesConfiguredTypes`
  - `TestDiskCollector_Collect_EmptyExcludeList`
- Coverage before: 82%
- Coverage after: 83%

## Changes

- Added `ExcludeTypes` field to `DiskConfig`
- Implemented filesystem type filtering in `DiskCollector.Collect()`
- Added table-driven tests for exclusion logic

## Testing

```bash
go test ./internal/collector/disk/ -v -cover
```

## Related Issues

Fixes #123
```

### Review Criteria

Reviewers will check:

1. **Tests exist and are meaningful** - Not just for coverage
2. **TDD was followed** - Tests should describe behavior before implementation
3. **Test names are descriptive** - Clear what's being tested
4. **Edge cases covered** - Error paths, empty inputs, boundaries
5. **No test pollution** - Tests are isolated and repeatable

---

## Code Style

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` and `goimports`
- Maximum line length: 100 characters (soft limit)
- Use meaningful variable names

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("DiskCollector.Collect: failed to read mounts: %w", err)
}

// Bad: Naked error return
if err != nil {
    return err
}
```

### Context Usage

```go
// Good: Accept context, use it
func (c *Collector) Collect(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // ... do work
}

// Bad: Ignore context
func (c *Collector) Collect() error {
    // ... no cancellation support
}
```

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `test` | Adding or fixing tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `docs` | Documentation only |
| `chore` | Build process, dependencies, etc. |
| `perf` | Performance improvement |

### Examples

```
feat(collector): add filesystem type exclusion

Add ExcludeTypes configuration option to DiskCollector that filters
out specified filesystem types (e.g., tmpfs, devtmpfs) from collection.

Closes #123
```

```
test(path): add edge case tests for max depth

Add table-driven tests covering:
- max_depth=0 (scan only root)
- max_depth=1 (scan one level)
- max_depth exceeds actual depth
```

```
fix(api): handle empty path list in scan endpoint

Return 400 Bad Request when paths array is empty instead of
triggering a scan of all paths.

Includes test for the error case.
```

---

## Questions?

If you have questions about contributing:

1. Check existing issues and PRs
2. Open a new issue with your question
3. Tag it with `question` label

We're happy to help newcomers get started with TDD!

---

**Remember: Tests first, always. No exceptions.**
