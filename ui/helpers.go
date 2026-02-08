package ui

import "fmt"

// FormatBytes formats bytes into a human-readable string
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// FormatDuration formats milliseconds into a human-readable duration string
func FormatDuration(ms int64) string {
	if ms == 0 {
		return "0ms"
	}

	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}

	seconds := ms / 1000
	remainingMs := ms % 1000

	if seconds < 60 {
		if remainingMs > 0 {
			return fmt.Sprintf("%.1fs", float64(ms)/1000.0)
		}
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%dm%ds", minutes, remainingSeconds)
}

// FormatGauge renders a text-based progress bar with percentage.
// Example output: [████████░░] 89.1%
func FormatGauge(percent float64, width int) string {
	if width < 3 {
		width = 3
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}

	return fmt.Sprintf("[%s] %5.1f%%", bar, percent)
}

// FormatNumber formats an integer with comma separators.
// Example: 1234567 → "1,234,567"
func FormatNumber(n int64) string {
	if n < 0 {
		return "-" + FormatNumber(-n)
	}

	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	// Insert commas from right to left
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
