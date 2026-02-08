package theme

import "github.com/gdamore/tcell/v2"

// Background colors
const (
	BgDefault   = tcell.ColorDefault
	BgHeader    = tcell.ColorDarkCyan
	BgNavBar    = tcell.ColorNavy
	BgStatusBar = tcell.ColorDarkGreen
	BgSelected  = tcell.ColorDarkBlue
)

// Foreground colors
const (
	FgPrimary   = tcell.ColorWhite
	FgSecondary = tcell.ColorSilver
	FgMuted     = tcell.ColorDarkGray
	FgAccent    = tcell.ColorAqua
	FgLabel     = tcell.ColorTeal
)

// Status colors
const (
	StatusOK       = tcell.ColorGreen
	StatusWarning  = tcell.ColorYellow
	StatusCritical = tcell.ColorRed
)

// Gauge colors
const (
	GaugeFilled = tcell.ColorGreen
	GaugeWarn   = tcell.ColorYellow
	GaugeCrit   = tcell.ColorRed
	GaugeEmpty  = tcell.ColorDarkGray
)

// Table header
const (
	TableHeader     = tcell.ColorTeal
	TableHeaderAttr = tcell.AttrBold
)

// tview dynamic color tags
const (
	TagAccent    = "[aqua]"
	TagPrimary   = "[white]"
	TagSecondary = "[silver]"
	TagMuted     = "[darkgray]"
	TagLabel     = "[teal]"
	TagReset     = "[-:-:-]"
	TagBold      = "[::b]"
)

// GaugeColor returns the appropriate color for a usage percentage.
func GaugeColor(percent float64) tcell.Color {
	if percent > 90 {
		return GaugeCrit
	}
	if percent > 75 {
		return GaugeWarn
	}
	return GaugeFilled
}

// StatusColor returns the appropriate color for a status string.
func StatusColor(status string) tcell.Color {
	switch status {
	case "OK", "ok", "connected":
		return StatusOK
	case "SCANNING", "WARNING", "warning":
		return StatusWarning
	case "ERROR", "error", "CRITICAL", "critical":
		return StatusCritical
	default:
		return FgSecondary
	}
}
