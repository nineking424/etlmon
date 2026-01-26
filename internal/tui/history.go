package tui

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rivo/tview"

	"github.com/etlmon/etlmon/internal/aggregator"
)

// HistoryView displays aggregated historical metrics
type HistoryView struct {
	view           *tview.TextView
	results        []aggregator.AggregationResult
	windowFilter   string
	resourceFilter string
	mu             sync.RWMutex
}

// NewHistoryView creates a new history view
func NewHistoryView() *HistoryView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)

	view.SetBorder(true).
		SetTitle(" Aggregated History (1/5/0 for window, Tab to switch, Q to quit) ").
		SetTitleAlign(tview.AlignLeft)

	return &HistoryView{
		view:         view,
		results:      make([]aggregator.AggregationResult, 0),
		windowFilter: "1m",
	}
}

// Update updates the view with new aggregation results
func (v *HistoryView) Update(results []aggregator.AggregationResult) {
	v.mu.Lock()
	// Append new results
	v.results = append(v.results, results...)

	// Keep only last 1000 results
	if len(v.results) > 1000 {
		v.results = v.results[len(v.results)-1000:]
	}
	v.mu.Unlock()

	v.render()
}

// SetWindowFilter sets the window size filter
func (v *HistoryView) SetWindowFilter(filter string) {
	v.mu.Lock()
	v.windowFilter = filter
	v.mu.Unlock()
	v.render()
}

// SetResourceFilter sets the resource type filter
func (v *HistoryView) SetResourceFilter(filter string) {
	v.mu.Lock()
	v.resourceFilter = filter
	v.mu.Unlock()
	v.render()
}

// render updates the text content
func (v *HistoryView) render() {
	v.mu.RLock()
	results := make([]aggregator.AggregationResult, len(v.results))
	copy(results, v.results)
	windowFilter := v.windowFilter
	resourceFilter := v.resourceFilter
	v.mu.RUnlock()

	text := v.formatResults(results, windowFilter, resourceFilter)

	// Need write lock when setting text to avoid race with GetText
	v.mu.Lock()
	v.view.SetText(text)
	v.mu.Unlock()
}

// formatResults formats aggregation results for display
func (v *HistoryView) formatResults(results []aggregator.AggregationResult, windowFilter, resourceFilter string) string {
	if len(results) == 0 {
		return "[yellow]No aggregated data yet...[white]\n\n[gray]Data will appear after the first window completes.[white]"
	}

	var sb strings.Builder

	// Filter results
	var filtered []aggregator.AggregationResult
	for _, r := range results {
		windowStr := FormatDuration(r.WindowSize)
		if windowFilter != "" && windowStr != windowFilter {
			continue
		}
		if resourceFilter != "" && r.ResourceType != resourceFilter {
			continue
		}
		filtered = append(filtered, r)
	}

	// Header
	sb.WriteString(fmt.Sprintf("[blue]Window: %s[white]  ", windowFilter))
	if resourceFilter != "" {
		sb.WriteString(fmt.Sprintf("[blue]Resource: %s[white]", resourceFilter))
	}
	sb.WriteString("\n\n")

	if len(filtered) == 0 {
		sb.WriteString("[gray]No data matching filters[white]")
		return sb.String()
	}

	// Sort by timestamp descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Show recent results (limit to 50)
	limit := 50
	if len(filtered) < limit {
		limit = len(filtered)
	}

	// Table header
	sb.WriteString(fmt.Sprintf("[green]%-12s %-15s %-10s %-8s %s[white]\n",
		"TIME", "RESOURCE", "METRIC", "TYPE", "VALUE"))
	sb.WriteString(strings.Repeat("─", 70) + "\n")

	for i := 0; i < limit; i++ {
		r := filtered[i]

		// Format value
		var valueStr string
		if strings.Contains(r.MetricName, "bytes") {
			valueStr = FormatBytes(r.Value)
		} else if strings.Contains(r.MetricName, "percent") {
			valueStr = fmt.Sprintf("%.1f%%", r.Value)
		} else {
			valueStr = fmt.Sprintf("%.2f", r.Value)
		}

		sb.WriteString(fmt.Sprintf("%-12s %-15s %-10s %-8s %s\n",
			r.Timestamp.Format("15:04:05"),
			r.ResourceType,
			truncate(r.MetricName, 10),
			r.AggregationType,
			valueStr,
		))
	}

	sb.WriteString(fmt.Sprintf("\n[gray]Showing %d of %d results[white]", limit, len(filtered)))

	return sb.String()
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// GetText returns the current text content
func (v *HistoryView) GetText() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.view.GetText(true)
}
