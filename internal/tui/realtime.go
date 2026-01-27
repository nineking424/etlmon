package tui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"

	"github.com/etlmon/etlmon/internal/collector"
)

// DisplayFormat represents the display format for metrics
type DisplayFormat int

const (
	FormatDetailed DisplayFormat = iota
	FormatTable
)

// RealtimeView displays real-time metrics
type RealtimeView struct {
	view           *tview.TextView
	currentMetrics []collector.Metric
	displayFormat  DisplayFormat
	mu             sync.RWMutex
}

// NewRealtimeView creates a new realtime view
func NewRealtimeView() *RealtimeView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)

	view.SetBorder(true).
		SetTitle(" Real-time Metrics (Press Tab to switch, Q to quit) ").
		SetTitleAlign(tview.AlignLeft)

	return &RealtimeView{
		view:           view,
		currentMetrics: make([]collector.Metric, 0),
		displayFormat:  FormatDetailed,
	}
}

// Update updates the view with new metrics
func (v *RealtimeView) Update(metrics []collector.Metric) {
	v.mu.Lock()
	v.currentMetrics = make([]collector.Metric, len(metrics))
	copy(v.currentMetrics, metrics)
	v.mu.Unlock()

	v.render()
}

// render updates the text content
func (v *RealtimeView) render() {
	v.mu.RLock()
	metrics := make([]collector.Metric, len(v.currentMetrics))
	copy(metrics, v.currentMetrics)
	format := v.displayFormat
	v.mu.RUnlock()

	var text string
	if format == FormatDetailed {
		text = v.formatMetrics(metrics)
	} else {
		text = v.formatMetricsTable(metrics)
	}

	// Need write lock when setting text to avoid race with GetText
	v.mu.Lock()
	v.view.SetText(text)
	v.updateTitle()
	v.mu.Unlock()
}

// formatMetrics formats metrics for display
func (v *RealtimeView) formatMetrics(metrics []collector.Metric) string {
	if len(metrics) == 0 {
		return "[yellow]Waiting for metrics...[white]"
	}

	var sb strings.Builder

	// Group by resource type
	grouped := make(map[string][]collector.Metric)
	for _, m := range metrics {
		grouped[m.ResourceType] = append(grouped[m.ResourceType], m)
	}

	// Sort resource types
	types := make([]string, 0, len(grouped))
	for t := range grouped {
		types = append(types, t)
	}
	sort.Strings(types)

	for _, resourceType := range types {
		resourceMetrics := grouped[resourceType]

		// Header
		sb.WriteString(fmt.Sprintf("[green]━━━ %s ━━━[white]\n", strings.ToUpper(resourceType)))

		// Sort metrics by name
		sort.Slice(resourceMetrics, func(i, j int) bool {
			return resourceMetrics[i].Name < resourceMetrics[j].Name
		})

		for _, m := range resourceMetrics {
			// Format value based on metric name
			var valueStr string
			if strings.Contains(m.Name, "bytes") {
				valueStr = FormatBytes(m.Value)
			} else if strings.Contains(m.Name, "percent") {
				valueStr = fmt.Sprintf("%.1f%%", m.Value)
				// Color code based on value
				if m.Value > 90 {
					valueStr = fmt.Sprintf("[red]%s[white]", valueStr)
				} else if m.Value > 70 {
					valueStr = fmt.Sprintf("[yellow]%s[white]", valueStr)
				} else {
					valueStr = fmt.Sprintf("[green]%s[white]", valueStr)
				}
			} else {
				valueStr = fmt.Sprintf("%.2f", m.Value)
			}

			// Add labels if present
			labelStr := ""
			if len(m.Labels) > 0 {
				if mp, ok := m.Labels["mountpoint"]; ok {
					labelStr = fmt.Sprintf(" [gray](%s)[white]", mp)
				}
			}

			sb.WriteString(fmt.Sprintf("  %-20s %s%s\n", m.Name, valueStr, labelStr))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("[gray]Last updated: %s[white]", time.Now().Format("15:04:05")))

	return sb.String()
}

// formatMetricsTable formats metrics as a table
func (v *RealtimeView) formatMetricsTable(metrics []collector.Metric) string {
	if len(metrics) == 0 {
		return "[yellow]Waiting for metrics...[white]"
	}

	var sb strings.Builder

	// Table header
	sb.WriteString(fmt.Sprintf("%-14s %-12s %-12s %-12s %-12s\n", "RESOURCE", "USAGE", "USED", "FREE", "TOTAL"))
	sb.WriteString(strings.Repeat("─", 62) + "\n")

	// Group metrics by resource type and mountpoint
	type resourceKey struct {
		resourceType string
		mountpoint   string
	}
	grouped := make(map[resourceKey]map[string]float64)

	for _, m := range metrics {
		key := resourceKey{resourceType: m.ResourceType}
		if mp, ok := m.Labels["mountpoint"]; ok {
			key.mountpoint = mp
		}

		if grouped[key] == nil {
			grouped[key] = make(map[string]float64)
		}
		grouped[key][m.Name] = m.Value
	}

	// Sort keys for consistent display
	var keys []resourceKey
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].resourceType != keys[j].resourceType {
			return keys[i].resourceType < keys[j].resourceType
		}
		return keys[i].mountpoint < keys[j].mountpoint
	})

	// Format each row
	for _, key := range keys {
		values := grouped[key]

		// Build resource name
		resourceName := key.resourceType
		if key.mountpoint != "" {
			resourceName = fmt.Sprintf("%s (%s)", key.resourceType, key.mountpoint)
		}

		// Get values
		usagePercent, hasUsage := values["usage_percent"]
		usedBytes, hasUsed := values["used_bytes"]
		totalBytes, hasTotal := values["total_bytes"]

		// For memory, available_bytes is FREE
		// For disk, free_bytes is FREE
		var freeBytes float64
		var hasFree bool
		if availBytes, ok := values["available_bytes"]; ok {
			freeBytes = availBytes
			hasFree = true
		} else if free, ok := values["free_bytes"]; ok {
			freeBytes = free
			hasFree = true
		}

		// Format usage with color coding
		var usageStr string
		if hasUsage {
			var color string
			if usagePercent > 90 {
				color = "[red]"
			} else if usagePercent > 70 {
				color = "[yellow]"
			} else {
				color = "[green]"
			}
			usageStr = fmt.Sprintf("%s%.1f%%[white]", color, usagePercent)
		} else {
			usageStr = "-"
		}

		// Format other columns
		var usedStr, freeStr, totalStr string
		if hasUsed {
			usedStr = FormatBytes(usedBytes)
		} else {
			usedStr = "-"
		}
		if hasFree {
			freeStr = FormatBytes(freeBytes)
		} else {
			freeStr = "-"
		}
		if hasTotal {
			totalStr = FormatBytes(totalBytes)
		} else {
			totalStr = "-"
		}

		sb.WriteString(fmt.Sprintf("%-14s %-12s %-12s %-12s %-12s\n",
			resourceName, usageStr, usedStr, freeStr, totalStr))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("[gray]Last updated: %s[white]", time.Now().Format("15:04:05")))

	return sb.String()
}

// SetDisplayFormat sets the display format
func (v *RealtimeView) SetDisplayFormat(format DisplayFormat) {
	v.mu.Lock()
	v.displayFormat = format
	v.mu.Unlock()
}

// ToggleDisplayFormat switches between display formats
func (v *RealtimeView) ToggleDisplayFormat() {
	v.mu.Lock()
	if v.displayFormat == FormatDetailed {
		v.displayFormat = FormatTable
	} else {
		v.displayFormat = FormatDetailed
	}
	v.mu.Unlock()
	v.render()
}

// GetDisplayFormat returns the current display format
func (v *RealtimeView) GetDisplayFormat() DisplayFormat {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.displayFormat
}

// updateTitle updates the view title based on display format
func (v *RealtimeView) updateTitle() {
	var title string
	if v.displayFormat == FormatDetailed {
		title = " Real-time Metrics [Detailed] (Tab: History | T: Table | Q: Quit) "
	} else {
		title = " Real-time Metrics [Table] (Tab: History | T: Detailed | Q: Quit) "
	}
	v.view.SetTitle(title)
}

// GetText returns the current text content
func (v *RealtimeView) GetText() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.view.GetText(true)
}
