# Excluded Scope - Table Format View

This document lists features and changes explicitly excluded from the table format view implementation (commit 9d9af24).

## Intentionally Excluded

### 1. History View Changes
- Table format applies only to Real-time view
- History view retains its original table format unchanged

### 2. Metric Struct Modifications
- No changes to `collector.Metric` struct
- No new fields or types in collector package

### 3. Storage Layer Changes
- No database schema changes for display format
- Display format is UI-only (not persisted)

### 4. Configuration Options
- No YAML config option for default display format
- Format preference not saved between sessions

### 5. External Dependencies
- No new third-party libraries added
- Uses only existing tview/tcell dependencies

### 6. Sorting/Filtering in Table View
- No column sorting capability
- No filtering by resource type in table view

### 7. Responsive Column Widths
- Fixed column widths (14, 12, 12, 12, 12)
- No dynamic width adjustment based on terminal size

### 8. Export Functionality
- No CSV/JSON export of table view data

## Rationale

These exclusions follow the "Must NOT Have" section from the implementation plan:
- Keep changes minimal and focused
- Preserve existing behavior as default
- Avoid scope creep
- Single-purpose implementation

## Future Considerations

The following could be added in future iterations:
- Persistent format preference in config
- Sortable columns
- Responsive column widths
- Export functionality
