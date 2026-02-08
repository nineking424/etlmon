package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestGaugeColor_ReturnsCorrectColor(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		expected tcell.Color
	}{
		{"0% returns green", 0, GaugeFilled},
		{"50% returns green", 50, GaugeFilled},
		{"75% returns green", 75, GaugeFilled},
		{"75.1% returns yellow", 75.1, GaugeWarn},
		{"85% returns yellow", 85, GaugeWarn},
		{"90% returns yellow", 90, GaugeWarn},
		{"90.1% returns red", 90.1, GaugeCrit},
		{"100% returns red", 100, GaugeCrit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GaugeColor(tt.percent))
		})
	}
}

func TestStatusColor_ReturnsCorrectColor(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected tcell.Color
	}{
		{"OK returns green", "OK", StatusOK},
		{"ok returns green", "ok", StatusOK},
		{"connected returns green", "connected", StatusOK},
		{"SCANNING returns yellow", "SCANNING", StatusWarning},
		{"WARNING returns yellow", "WARNING", StatusWarning},
		{"ERROR returns red", "ERROR", StatusCritical},
		{"CRITICAL returns red", "CRITICAL", StatusCritical},
		{"unknown returns secondary", "UNKNOWN", FgSecondary},
		{"empty returns secondary", "", FgSecondary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, StatusColor(tt.status))
		})
	}
}
