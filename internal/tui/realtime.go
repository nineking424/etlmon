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

// RealtimeView displays real-time metrics
type RealtimeView struct {
	view           *tview.TextView
	currentMetrics []collector.Metric
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
	v.mu.RUnlock()

	text := v.formatMetrics(metrics)

	// Need write lock when setting text to avoid race with GetText
	v.mu.Lock()
	v.view.SetText(text)
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

// GetText returns the current text content
func (v *RealtimeView) GetText() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.view.GetText(true)
}
